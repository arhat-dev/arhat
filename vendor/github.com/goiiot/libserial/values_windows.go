// +build windows

/*
 * Copyright Go-IIoT (https://github.com/goiiot)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package libserial

import "unsafe"

type Parity byte
type StopBit byte

const (
	ParityNone       Parity  = 0
	ParityOdd        Parity  = 1
	ParityEven       Parity  = 2
	ParityMark       Parity  = 3
	ParitySpace      Parity  = 4
	StopBitOne       StopBit = 0
	StopBitOneHalf   StopBit = 1
	StopBitTwo       StopBit = 2
	_dcbSize                 = uint32(unsafe.Sizeof(_dcb{}))
	maskDataBits             = uint64(0)
	maskBaudRate             = uint64(0)
	softwareCtrlFlag         = 0
	hardwareCtrlFlag         = 0
	parityEnable             = 0
	dataBits5                = 0
	dataBits6                = 0
	dataBits7                = 0
	dataBits8                = 0
)

type _dcb struct {
	DCBLength, BaudRate                            uint32
	flags                                          [4]byte
	wReserved, XOnLim, XOffLim                     uint16
	ByteSize, Parity, StopBits                     byte
	XonChar, XOffChar, ErrorChar, EofChar, EvtChar byte
	wReserved1                                     uint16
}

// ignored in windows
var validBaudRates map[int]uint32
