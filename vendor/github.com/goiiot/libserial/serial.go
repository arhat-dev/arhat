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
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"
)

var (
	// ErrDeviceNameEmpty happens when opening a device with empty name
	ErrDeviceNameEmpty = errors.New("device name should not be empty")
)

// SerialPort of serial
type SerialPort struct {
	// actions performer
	f     *os.File
	flush func() error

	// options
	// common options
	dev         string
	readTimeout time.Duration

	// options for windows
	baudRate uint64
	dataBits byte
	stopBits byte
	parity   byte

	// options for posix/linux
	inputOptions   uint64
	controlOptions uint64
}

// Write bytes to serial connection
func (s *SerialPort) Write(data []byte) (int, error) {
	return s.f.Write(data)
}

// Read bytes from serial connection
func (s *SerialPort) Read(data []byte) (int, error) {
	return s.f.Read(data)
}

// Close serial connection
func (s *SerialPort) Close() error {
	return s.f.Close()
}

// Flush serial input/output queue
func (s *SerialPort) Flush() error {
	return s.flush()
}

// File returns the underlying file handler for this serial port
func (s *SerialPort) File() *os.File {
	return s.f
}

// Open serial port
func Open(device string, options ...Option) (*SerialPort, error) {
	if device == "" {
		return nil, ErrDeviceNameEmpty
	}

	port := &SerialPort{dev: device}

	// set defaults 9600 8N1
	defaultOptions := []Option{
		WithBaudRate(9600),
		WithDataBits(8),
		WithParity(ParityNone),
		WithStopBits(StopBitOne),
	}

	for _, setDefaultOption := range defaultOptions {
		if err := setDefaultOption(port); err != nil {
			return nil, err
		}
	}

	// set user defined options
	for _, setOption := range options {
		if err := setOption(port); err != nil {
			return nil, err
		}
	}

	// open platform specific serial port
	if err := port.open(); err != nil {
		return nil, err
	}

	return port, nil
}

// Option for serial conn options
type Option func(c *SerialPort) error

// WithReadTimeout set timeout timer for read operations
// if no read timeout set, use blocking read
func WithReadTimeout(timeout time.Duration) Option {
	return func(s *SerialPort) error {
		if timeout > 0 {
			s.readTimeout = timeout
		}
		return nil
	}
}

// WithBaudRate set serial baud rate
// default is 9600
func WithBaudRate(rate int) Option {
	return func(c *SerialPort) error {
		// do not check baud rate on windows
		if runtime.GOOS == "windows" {
			c.baudRate = uint64(rate)
			return nil
		}

		// check baud rate on non windows platform
		baudRate, ok := validBaudRates[rate]
		if !ok {
			return fmt.Errorf("invalid baud rate: %v", rate)
		}

		c.baudRate = uint64(baudRate)

		// clear baud rate flags and set new baud rate
		if runtime.GOOS == "linux" {
			c.controlOptions &= ^uint64(maskBaudRate)
			c.controlOptions |= uint64(baudRate)
		}

		return nil
	}
}

// WithDataBits set the data bits for SerialPort
// available values are {5, 6, 7, 8}
// default is 8
func WithDataBits(d int) Option {
	return func(c *SerialPort) error {
		c.dataBits = byte(d)

		// clear flags
		c.controlOptions &= ^uint64(maskDataBits)

		switch d {
		case 5:
			c.controlOptions |= dataBits5
		case 6:
			c.controlOptions |= dataBits6
		case 7:
			c.controlOptions |= dataBits7
		case 8:
			c.controlOptions |= dataBits8
		default:
			return fmt.Errorf("invalid data bits: %v", d)
		}
		return nil
	}
}

// WithParity set parity mode
// available values are {ParityNone, ParityOdd, ParityEven}
// default is ParityNone
func WithParity(p Parity) Option {
	return func(c *SerialPort) error {
		c.parity = byte(p)

		// clear flags
		c.controlOptions &= ^uint64(ParityOdd | ParityMark | parityEnable)

		// keep ParityMark and ParitySpace separated for darwin compatibility
		if p == ParityMark {
			c.controlOptions |= uint64(p) | parityEnable
			return nil
		}

		if p == ParitySpace {
			c.controlOptions |= uint64(p) | parityEnable
			return nil
		}

		switch p {
		case ParityNone:
			// do nothing default is None
		case ParityOdd, ParityEven:
			c.controlOptions |= uint64(p) | parityEnable
		default:
			return fmt.Errorf("invalid parity mode: %v", p)
		}

		return nil
	}
}

// WithStopBits set stop bits for SerialPort port
// available values are {StopBitOne, StopBitOneHalf, StopBitTwo}
// default is StopBitOne
func WithStopBits(s StopBit) Option {
	return func(c *SerialPort) error {
		c.stopBits = byte(s)

		// clear flags
		c.controlOptions &= ^uint64(StopBitTwo)

		switch s {
		case StopBitOne:
			// do nothing
		case StopBitTwo:
			c.controlOptions |= uint64(s)
		default:
			return fmt.Errorf("invalid stop bits: %v", s)
		}

		return nil
	}
}

// WithSoftwareFlowControl enable software flow control
func WithSoftwareFlowControl(enable bool) Option {
	return func(c *SerialPort) error {
		// TODO: implement software flow control in windows

		// clear flags
		c.inputOptions &= ^uint64(softwareCtrlFlag)

		if enable {
			c.inputOptions |= uint64(softwareCtrlFlag)
		}

		return nil
	}
}

// WithHardwareFlowControl enable hardware flow control
func WithHardwareFlowControl(enable bool) Option {
	return func(c *SerialPort) error {
		// TODO: implement hardware flow control in windows

		// clear flags
		c.controlOptions &= ^uint64(hardwareCtrlFlag)

		if enable {
			c.controlOptions |= uint64(hardwareCtrlFlag)
		}
		return nil
	}
}
