// +build !noext

package extensionutil

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"arhat.dev/arhat-proto/arhatgopb"
	"arhat.dev/pkg/log"
)

type SyncHandleFunc func(stopSig <-chan struct{}, msgCh <-chan *arhatgopb.Msg, cmdCh chan<- *arhatgopb.Cmd) error

func NewHandler(handler SyncHandleFunc) *Handler {
	return &Handler{processReq: handler, logger: log.Log.WithName("ext")}
}

type Handler struct {
	logger     log.Interface
	processReq SyncHandleFunc
}

func (h *Handler) handleJSONStream(reqExit <-chan struct{}, r io.Reader, w *flushWriter) error {
	dec := json.NewDecoder(r)
	enc := json.NewEncoder(w)

	cmdCh := make(chan *arhatgopb.Cmd, 1)
	msgCh := make(chan *arhatgopb.Msg, 1)
	errCh := make(chan error)

	// process requests in agent
	go func() {
		select {
		case errCh <- h.processReq(reqExit, msgCh, cmdCh):
		case <-reqExit:
		}
	}()

	// process commands from agent
	go func() {
		for {
			select {
			case cmd, more := <-cmdCh:
				if !more {
					return
				}

				err := enc.Encode(cmd)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("failed to marshal and write cmd: %w", err):
					case <-reqExit:
					}

					return
				}
				w.Flush()
			case <-reqExit:
				return
			}
		}
	}()

	defer func() {
		close(msgCh)
	}()

	// receive requests from extension driver
	for {
		msg := new(arhatgopb.Msg)
		err := dec.Decode(msg)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("failed to unmarshal json as proto message: %w", err)
			}

			return nil
		}

		select {
		case err = <-errCh:
			return err
		case msgCh <- msg:
		case <-reqExit:
			return nil
		}
	}
}

func (h *Handler) handleProtobufStream(reqExit <-chan struct{}, r io.Reader, w *flushWriter) error {
	br := bufio.NewReader(r)

	cmdCh := make(chan *arhatgopb.Cmd, 1)
	msgCh := make(chan *arhatgopb.Msg, 1)
	errCh := make(chan error)
	// process requests in agent
	go func() {
		select {
		case errCh <- h.processReq(reqExit, msgCh, cmdCh):
		case <-reqExit:
		}
	}()

	// process commands from agent
	go func() {
		for {
			sizeBuf := make([]byte, 10)
			select {
			case cmd, more := <-cmdCh:
				if !more {
					return
				}

				data, err := cmd.Marshal()
				if err != nil {
					select {
					case errCh <- fmt.Errorf("failed to marshal cmd: %w", err):
					case <-reqExit:
					}
				}

				i := binary.PutUvarint(sizeBuf, uint64(len(data)))
				_, _ = w.Write(sizeBuf[:i])
				_, err = w.Write(data)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("failed to write cmd data: %w", err):
					case <-reqExit:
					}
					return
				}

				w.Flush()
			case <-reqExit:
				return
			}
		}
	}()

	defer func() {
		close(msgCh)
	}()

	for {
		size, err := binary.ReadUvarint(br)
		if err != nil {
			return fmt.Errorf("unexpected non size")
		}
		buf := make([]byte, size)
		_, err = io.ReadFull(br, buf)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("failed to read full message: %w", err)
			}

			return nil
		}
		msg := new(arhatgopb.Msg)
		err = msg.Unmarshal(buf)
		if err != nil {
			return fmt.Errorf("failed to unmarshal proto message: %w", err)
		}

		select {
		case err = <-errCh:
			return err
		case msgCh <- msg:
		case <-reqExit:
			return nil
		}
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.WithFields(log.String("path", req.URL.Path))
	logger.V("serving extension")

	if req.Method != http.MethodPost {
		http.Error(w, "invalid non post request", http.StatusBadRequest)
		return
	}

	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "unable to get http.Flusher, cannot show establish connection", http.StatusInternalServerError)
		return
	}

	fw := &flushWriter{
		flusher: f,
		writer:  w,
	}

	header := req.Header.Get("Content-Type")
	i := strings.Index(header, ";")
	if i == -1 {
		i = len(header)
	}
	contentType := strings.TrimSpace(strings.ToLower(header[:i]))

	var err error
	w.Header().Set("Transfer-Encoding", "chunked")

	switch contentType {
	case "application/json", "":
		logger.D("serving json stream")
		err = h.handleJSONStream(req.Context().Done(), req.Body, fw)
	case "application/protobuf":
		logger.D("serving protobuf stream")
		err = h.handleProtobufStream(req.Context().Done(), req.Body, fw)
	default:
		http.Error(w, "unexpected Content-Type: "+contentType, http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "request finished with error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

type flushWriter struct {
	flusher http.Flusher
	writer  io.Writer
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.writer.Write(p)
	if err != nil {
		return
	}

	return
}

func (fw *flushWriter) Flush() {
	fw.flusher.Flush()
}
