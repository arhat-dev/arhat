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

// Package nethelper wraps various kind of network support with universal
// api to dial and listen
// you need to import according packages to enabled certain network supprot
//
// to support
// 		tcp/tcp4/tcp6/unix (with or without tls)
// 		udp/udp4/udp6 (without tls)
// 		unixpacket/unixgram (with or without tls)
//  import _ "arhat.dev/pkg/nethelper/stdnet"
//
// to support
// 		udp/udp4/udp6 (with tls)
//  import _ "arhat.dev/pkg/nethelper/piondtls"
//
// to support
// 		pipe (with or without tls)
//  import _ "arhat.dev/pkg/nethelper/pipenet"
package nethelper
