// +build !windows,!linux

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
	termiosReqGet = uint(unix.TIOCGETA)
	termiosReqSet = uint(unix.TIOCSETA)
	maskBaudRate  = uint64(0)
	ParityMark    = Parity(0)
	ParitySpace   = Parity(0)
)

func mkFlushFunc(fd uintptr) func() error {
	return func() error {
		tty, err := unix.IoctlGetTermios(int(fd), termiosReqGet)
		if err != nil {
			return err
		}

		// set serial port again for input/output flush
		return unix.IoctlSetTermios(int(fd), unix.TIOCSETAF, tty)
	}
}
