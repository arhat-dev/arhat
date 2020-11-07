/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package codecpb

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"

	"arhat.dev/arhat-proto/arhatgopb"
	"github.com/gogo/protobuf/proto"

	"arhat.dev/libext/types"
	"arhat.dev/libext/util"
)

var (
	ErrNotProtobufMessage = errors.New("invalid not protobuf message")
)

var pbBufPool = &sync.Pool{
	New: func() interface{} {
		return proto.NewBuffer(nil)
	},
}

func getPbBuf(data []byte) *proto.Buffer {
	buf := pbBufPool.Get().(*proto.Buffer)
	buf.SetBuf(data)
	return buf
}

type Codec struct{}

func (c *Codec) Type() arhatgopb.CodecType {
	return arhatgopb.CODEC_PROTOBUF
}

func (c *Codec) NewEncoder(w io.Writer) types.Encoder {
	return &Encoder{w}
}

func (c *Codec) NewDecoder(r io.Reader) types.Decoder {
	return &Decoder{bufio.NewReader(r)}
}

func (c *Codec) Unmarshal(data []byte, out interface{}) error {
	m, ok := out.(proto.Message)
	if !ok {
		return ErrNotProtobufMessage
	}

	buf := getPbBuf(data)

	err := buf.Unmarshal(m)
	pbBufPool.Put(buf)
	return err
}

func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, ErrNotProtobufMessage
	}

	return proto.Marshal(m)
}

type Encoder struct {
	w io.Writer
}

func (enc *Encoder) Encode(any interface{}) error {
	var (
		data []byte
		err  error
	)
	switch t := any.(type) {
	case proto.Marshaler:
		data, err = t.Marshal()
	case proto.Message:
		data, err = proto.Marshal(t)
	default:
		return ErrNotProtobufMessage
	}
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	sizeBuf := util.GetBytesBuf(10)
	i := binary.PutUvarint(sizeBuf, uint64(len(data)))
	_, err = enc.w.Write(sizeBuf[:i])
	if err != nil {
		util.PutBytesBuf(&sizeBuf)
		return fmt.Errorf("failed to write message size: %w", err)
	}
	util.PutBytesBuf(&sizeBuf)

	_, err = enc.w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write message body: %w", err)
	}

	return nil
}

type Decoder struct {
	r *bufio.Reader
}

func (dec *Decoder) Decode(out interface{}) error {
	m, ok := out.(proto.Message)
	if !ok {
		return ErrNotProtobufMessage
	}

	size, err := binary.ReadUvarint(dec.r)
	if err != nil {
		if err == io.EOF {
			return err
		}
		return fmt.Errorf("failed to read size of the message: %w", err)
	}

	data := util.GetBytesBuf(int(size))
	_, err = io.ReadFull(dec.r, data[:size])
	if err != nil {
		util.PutBytesBuf(&data)
		return fmt.Errorf("failed to read message body: %w", err)
	}
	buf := getPbBuf(data[:size])
	err = buf.Unmarshal(m)

	util.PutBytesBuf(&data)
	pbBufPool.Put(buf)
	return err
}
