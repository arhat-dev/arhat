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

package types

import (
	"io"

	"arhat.dev/arhat-proto/arhatgopb"
)

// Encoder for stream data encoding
type Encoder interface {
	// Encode one message to stream
	Encode(any interface{}) error
}

// Decoder for stream data decoding
type Decoder interface {
	// Decode one message from stream
	Decode(out interface{}) error
}

type (
	MarshalFunc   func(v interface{}) ([]byte, error)
	UnmarshalFunc func(data []byte, out interface{}) error
)

// Codec knows how to encode/decode all kinds of data,
// including stream data
type Codec interface {
	// Type is the codec kind
	Type() arhatgopb.CodecType

	// NewEncoder creates a new Encoder which will encode messages to the Writer
	NewEncoder(w io.Writer) Encoder

	// NewEncoder creates a new Decoder which will decode messages from the Reader
	NewDecoder(r io.Reader) Decoder

	// Unmarshal non stream data
	Unmarshal(data []byte, out interface{}) error

	// Marshal v as non stream data
	Marshal(v interface{}) ([]byte, error)
}
