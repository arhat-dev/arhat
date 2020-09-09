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

import (
	"fmt"
	"math"
	"os"

	"golang.org/x/sys/unix"
)

func (s *SerialPort) open() error {
	var (
		f   *os.File
		err error
	)

	f, err = os.OpenFile(s.dev, serialFileFlag, 0666)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			f.Close()
		} else {
			s.f = f
		}
	}()

	// get posix timeout value (seconds / 10)
	timeout := int64(0)
	if s.readTimeout > 0 {
		timeout = s.readTimeout.Nanoseconds() / 1e8
		if timeout > math.MaxUint8 {
			timeout = math.MaxUint8
		}
	}

	// check sys baud rate when baud rate not present
	if s.baudRate == unix.B0 {
		tty, err := unix.IoctlGetTermios(int(f.Fd()), termiosReqGet)
		if err != nil {
			return fmt.Errorf("fail to get serial port config: %v", err)
		}

		s.baudRate = uint64(tty.Cflag) & maskBaudRate
		if s.baudRate == unix.B0 {
			s.baudRate = uint64(tty.Ispeed)
		}

		if s.baudRate == unix.B0 {
			return fmt.Errorf("fail to determine serial port baud rate: %v", err)
		}
	}

	tty := &unix.Termios{
		Cflag:  unix.CREAD | unix.CLOCAL | termiosFlagType(s.controlOptions),
		Iflag:  termiosFlagType(s.inputOptions),
		Ispeed: termiosSpeedType(s.baudRate),
		Ospeed: termiosSpeedType(s.baudRate),
	}

	if timeout == 0 {
		// set blocking read with at least 1 byte have read if no timeout defined
		tty.Cc[unix.VMIN] = 1
	}
	// set read timeout
	tty.Cc[unix.VTIME] = uint8(timeout)

	err = unix.IoctlSetTermios(int(f.Fd()), termiosReqSet, tty)
	if err != nil {
		return err
	}

	s.flush = mkFlushFunc(f.Fd())

	return nil
}
