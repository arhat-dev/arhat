// +build openbsd

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

type Parity uint32
type StopBit uint32
type termiosFlagType = uint32
type termiosSpeedType = int32

var validBaudRates = map[int]uint64{
	0:      unix.B0, // detect baud rate automatically
	50:     unix.B50,
	75:     unix.B75,
	110:    unix.B110,
	134:    unix.B134,
	150:    unix.B150,
	200:    unix.B200,
	300:    unix.B300,
	600:    unix.B600,
	1200:   unix.B1200,
	1800:   unix.B1800,
	2400:   unix.B2400,
	4800:   unix.B4800,
	9600:   unix.B9600,
	19200:  unix.B19200,
	38400:  unix.B38400,
	57600:  unix.B57600,
	115200: unix.B115200,
	230400: unix.B230400,
}
