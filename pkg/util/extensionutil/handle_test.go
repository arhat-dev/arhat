package extensionutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"arhat.dev/arhat-proto/arhatgopb"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ServeHTTP(t *testing.T) {
	registerMsg, err := arhatgopb.NewMsg(1, 0, &arhatgopb.RegisterMsg{
		Name: "test",
	})
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed msg creation")
	}

	cmds := []*arhatgopb.Cmd{
		{
			Kind:    arhatgopb.CMD_PERIPHERAL_CONNECT,
			Id:      1,
			Seq:     1,
			Payload: nil,
		},
		{
			Kind:    arhatgopb.CMD_PERIPHERAL_CONNECT,
			Id:      1,
			Seq:     1,
			Payload: nil,
		},
	}

	h := NewHandler(func(stopSig <-chan struct{}, msgCh <-chan *arhatgopb.Msg, cmdCh chan<- *arhatgopb.Cmd) error {
		i := 0
		for msg := range msgCh {
			_ = msg
			cmdCh <- cmds[i]
			i++
		}

		return nil
	})

	//buf := new(bytes.Buffer)
	for i, msg := range []*arhatgopb.Msg{
		registerMsg,
	} {
		msgBytes, err := json.Marshal(msg)
		assert.NoError(t, err)

		pr, pw := io.Pipe()
		jsonRec := httptest.NewRecorder()
		jsonReq := httptest.NewRequest(http.MethodPost, "/peripherals", pr)

		go func() {
			for j := 0; j < len(cmds); j++ {
				_, err = pw.Write(msgBytes)
				assert.NoError(t, err)

				time.Sleep(time.Second)
			}

			assert.NoError(t, pw.Close())
		}()

		h.ServeHTTP(jsonRec, jsonReq)

		expectedCell, err := json.Marshal(cmds[i])
		assert.NoError(t, err)
		expected := ""
		for range cmds {
			expected += string(expectedCell)
			expected += "\n"
		}

		assert.Equal(t, expected, jsonRec.Body.String())
	}
}
