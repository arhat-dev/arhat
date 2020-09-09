// +build !nodev_serial,!nodev

package serial

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/goiiot/libserial"

	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/devices/deviceutils"
	"arhat.dev/arhat/pkg/types"
)

func init() {
	types.RegisterDeviceConnectivityFactory("serial",
		aranyagopb.DEVICE_CONNECTIVITY_MODE_CLIENT, NewSerialDeviceConnectivity)
}

func NewSerialDeviceConnectivity(
	target string,
	params map[string]string,
	tlsConfig *aranyagopb.DeviceConnectivityTLSConfig,
) (types.DeviceConnectivity, error) {
	var opts []libserial.Option
	for k, v := range params {
		switch k {
		case "hw_flow_ctrl", "rts/cts":
			switch v {
			case "enabled":
				opts = append(opts, libserial.WithHardwareFlowControl(true))
			case "disabled":
				opts = append(opts, libserial.WithHardwareFlowControl(false))
			}
		case "sw_flow_ctrl":
			switch v {
			case "enabled":
				opts = append(opts, libserial.WithSoftwareFlowControl(true))
			case "disabled":
				opts = append(opts, libserial.WithSoftwareFlowControl(false))
			}
		case "parity":
			parity, ok := map[string]libserial.Parity{
				"none":  libserial.ParityNone,
				"even":  libserial.ParityEven,
				"odd":   libserial.ParityOdd,
				"mark":  libserial.ParityMark,
				"space": libserial.ParitySpace,
			}[v]
			if !ok {
				return nil, fmt.Errorf("invalid parity param %q", v)
			}

			opts = append(opts, libserial.WithParity(parity))
		case "baud_rate":
			baudRate, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("incalid baud_rate param %q: %w", v, err)
			}

			opts = append(opts, libserial.WithBaudRate(baudRate))
		case "data_bits":
			dataBits, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("incalid data_bits param %q: %w", v, err)
			}

			opts = append(opts, libserial.WithDataBits(dataBits))
		case "stop_bits":
			var stopBits libserial.StopBit
			switch v {
			case "1":
				stopBits = libserial.StopBitOne
			case "2":
				stopBits = libserial.StopBitTwo
			default:
				return nil, fmt.Errorf("invalid stop_bits param %s", v)
			}

			opts = append(opts, libserial.WithStopBits(stopBits))
		default:
			return nil, fmt.Errorf("invalid param %q: %q", k, v)
		}
	}

	return &serialDeviceConnectivity{
		device: target,
		opts:   opts,

		mu: new(sync.RWMutex),
	}, nil
}

var _ types.DeviceConnectivity = &serialDeviceConnectivity{}

type serialDeviceConnectivity struct {
	device string
	opts   []libserial.Option

	openedPort *libserial.SerialPort
	mu         *sync.RWMutex
}

func (s *serialDeviceConnectivity) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	port, err := libserial.Open(s.device, s.opts...)
	if err != nil {
		return fmt.Errorf("failed to open serial port %q: %w", s.device, err)
	}

	s.openedPort = port

	return nil
}

func (s *serialDeviceConnectivity) doReadAfterWrite(params map[string]string) ([][]byte, error) {
	err := func() error {
		s.mu.RLock()
		defer s.mu.RUnlock()

		if s.openedPort == nil {
			return fmt.Errorf("serial port not opened")
		}

		return nil
	}()

	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return deviceutils.ReadAfterWrite(s.openedPort, s.openedPort, params)
}

func (s *serialDeviceConnectivity) Operate(params map[string]string) ([][]byte, error) {
	return s.doReadAfterWrite(params)
}

func (s *serialDeviceConnectivity) CollectMetrics(params map[string]string) ([][]byte, error) {
	return s.doReadAfterWrite(params)
}

func (s *serialDeviceConnectivity) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.openedPort.Close()
	if err != nil {
		return err
	}

	s.openedPort = nil
	return nil
}
