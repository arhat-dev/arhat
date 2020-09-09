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

import (
	"fmt"
	"math"
	"os"
	"syscall"
	"unsafe"

	win "golang.org/x/sys/windows"
)

const (
	SetCommState    = "SetCommState"
	SetCommTimeouts = "SetCommTimeouts"
	SetCommMask     = "SetCommMask"
	SetupComm       = "SetupComm"
	PurgeComm       = "PurgeComm"
)

var (
	comSyscall     = map[string]func(s *SerialPort) error{}
	comSyscallList = []string{
		SetCommState, SetupComm, SetCommTimeouts, SetCommMask, PurgeComm,
	}
)

func (s *SerialPort) open() error {
	if s.dev[0] != '\\' {
		s.dev = `\\.\` + s.dev
	}

	f, err := os.OpenFile(s.dev, os.O_RDWR, 0)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	s.f = f
	for _, name := range comSyscallList {
		if name != PurgeComm {
			if err = comSyscall[name](s); err != nil {
				return err
			}
		}
	}

	s.flush = func() error {
		return comSyscall[PurgeComm](s)
	}

	return nil
}

func init() {
	dll, err := win.LoadLibrary("kernel32.dll")

	if err != nil {
		// you probably should use another serial port library in windows if this happens
		panic("init kernel32.dll failed")
	}

	// set raw syscall via proc addresses
	var rawSyscall = make(map[string]func(args ...uintptr) (uintptr, error))
	for _, name := range comSyscallList {
		addr, err := win.GetProcAddress(dll, name)
		if err != nil {
			panic("init raw syscall failed")
		}

		rawSyscall[name] = func(args ...uintptr) (uintptr, error) {
			n := uintptr(len(args))
			args = append(args, make([]uintptr, 3-n)...)
			r, _, err := syscall.Syscall(addr, n, args[0], args[1], args[2])
			return r, fmt.Errorf("syscall %s failed: %v", name, err)
		}
	}

	// wrap raw syscalls for setup ease
	{
		comSyscall[PurgeComm] = func(s *SerialPort) error {
			// 0x000F = (PURGE_TXABORT | PURGE_RXABORT | PURGE_TXCLEAR | PURGE_RXCLEAR)
			r, err := rawSyscall[PurgeComm](s.f.Fd(), 0x000F)
			if r == 0 {
				return err
			}
			return nil
		}

		comSyscall[SetupComm] = func(s *SerialPort) error {
			r, err := rawSyscall[SetupComm](s.f.Fd(), 64, 64)
			if r == 0 {
				return err
			}
			return nil
		}

		comSyscall[SetCommMask] = func(s *SerialPort) error {
			// EV_RXCHAR
			r, err := rawSyscall[SetCommMask](s.f.Fd(), 0x0001)
			if r == 0 {
				return err
			}
			return nil
		}

		comSyscall[SetCommTimeouts] = func(s *SerialPort) error {
			timeout := &struct {
				ReadIntervalTimeout         uint32
				ReadTotalTimeoutMultiplier  uint32
				ReadTotalTimeoutConstant    uint32
				WriteTotalTimeoutMultiplier uint32
				WriteTotalTimeoutConstant   uint32
			}{
				ReadIntervalTimeout:        math.MaxUint32,
				ReadTotalTimeoutMultiplier: math.MaxUint32,
				ReadTotalTimeoutConstant: func() uint32 {
					if s.readTimeout == 0 {
						return math.MaxUint32 - 1
					}
					return uint32(s.readTimeout.Nanoseconds() / 1e6)
				}(),
			}

			r, err := rawSyscall[SetCommTimeouts](s.f.Fd(), uintptr(unsafe.Pointer(timeout)))
			if r == 0 {
				return err
			}
			return nil
		}

		comSyscall[SetCommState] = func(s *SerialPort) error {
			d := &_dcb{
				DCBLength: _dcbSize,
				flags:     [4]byte{0x11, 0x00, 0x00, 0x00},
				BaudRate:  uint32(s.baudRate),
				ByteSize:  s.dataBits,
				StopBits:  s.stopBits,
				Parity:    s.parity,
			}

			r, err := rawSyscall[SetCommState](s.f.Fd(), uintptr(unsafe.Pointer(d)))
			if r == 0 {
				return err
			}
			return nil
		}
	}
}
