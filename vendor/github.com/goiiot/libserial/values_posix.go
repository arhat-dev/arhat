// +build !windows

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

import "golang.org/x/sys/unix"

const (
	StopBitOne       = StopBit(0) // default to stop bit one
	StopBitTwo       = StopBit(unix.CSTOPB)
	parityEnable     = unix.PARENB
	ParityNone       = Parity(0)
	ParityOdd        = Parity(unix.PARODD)
	ParityEven       = unix.PARENB // enable parity will default to even mode
	maskDataBits     = unix.CSIZE
	serialFileFlag   = unix.O_RDWR | unix.O_NOCTTY | unix.O_NONBLOCK
	softwareCtrlFlag = unix.IXON | unix.IXOFF | unix.IXANY
	hardwareCtrlFlag = unix.CRTSCTS
	dataBits5        = unix.CS5
	dataBits6        = unix.CS6
	dataBits7        = unix.CS7
	dataBits8        = unix.CS8
)
