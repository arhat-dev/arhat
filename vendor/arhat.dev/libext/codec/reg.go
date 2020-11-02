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

// Package codec is the registration center for supported codecs
// you can register whatever codec implementation you want
//
// e.g. for json codec, you can register you own codec implementation
// based on other json library (other than encoding/json)
//
// by default, using encoding/json for json codec and
// gogoproto for protobuf codec
package codec

import (
	"arhat.dev/arhat-proto/arhatgopb"

	"arhat.dev/libext/types"
)

var (
	supportedCodec = make(map[arhatgopb.CodecType]types.Codec)
)

func RegisterCodec(kind arhatgopb.CodecType, codec types.Codec) {
	supportedCodec[kind] = codec
}

func GetCodec(kind arhatgopb.CodecType) types.Codec {
	codec, ok := supportedCodec[kind]
	if !ok {
		panic("no codec registered for this kind")
	}

	return codec
}
