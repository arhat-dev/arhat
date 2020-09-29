// +build !nodev

package arhatgopb

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/gogo/protobuf/proto"
)

func NewDeviceCmd(deviceID, seq uint64, cmd proto.Marshaler) (*DeviceCmd, error) {
	var kind DeviceCmdType
	switch cmd.(type) {
	case *DeviceConnectCmd:
		kind = CMD_DEV_CONNECT
	case *DeviceOperateCmd:
		kind = CMD_DEV_OPERATE
	case *DeviceMetricsCollectCmd:
		kind = CMD_DEV_COLLECT_METRICS
	case *DeviceCloseCmd:
		kind = CMD_DEV_CLOSE
	default:
		return nil, fmt.Errorf("unknown device cmd: %v", cmd)
	}

	data, err := cmd.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cmd: %w", err)
	}

	return &DeviceCmd{
		Kind:     kind,
		DeviceId: deviceID,
		Seq:      seq,
		Payload:  data,
	}, nil
}

func NewDeviceMsg(deviceID, ack uint64, msg proto.Marshaler) (*DeviceMsg, error) {
	var kind DeviceMsgType
	switch msg.(type) {
	case *DeviceRegisterMsg:
		kind = MSG_DEV_REGISTER
	case *DeviceOperateResultMsg:
		kind = MSG_DEV_OPERATION_RESULT
	case *DeviceMetricsMsg:
		kind = MSG_DEV_METRICS
	case *DeviceDoneMsg:
		kind = MSG_DEV_DONE
	default:
		return nil, fmt.Errorf("unknown device msg: %v", msg)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg: %w", err)
	}

	return &DeviceMsg{
		Kind:     kind,
		DeviceId: deviceID,
		Ack:      ack,
		Payload:  data,
	}, nil
}

func (c *TLSConfig) GetTLSConfig() (_ *tls.Config, err error) {
	if c == nil {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		ServerName:         c.ServerName,
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	for _, c := range c.CipherSuites {
		tlsConfig.CipherSuites = append(tlsConfig.CipherSuites, uint16(c))
	}

	if caBytes := c.CaCert; len(caBytes) != 0 {
		tlsConfig.RootCAs = x509.NewCertPool()
		block, _ := pem.Decode(caBytes)
		if block == nil {
			// not encoded in pem format
			var caCerts []*x509.Certificate
			caCerts, err = x509.ParseCertificates(caBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ca certs: %w", err)
			}
			for i := range caCerts {
				tlsConfig.RootCAs.AddCert(caCerts[i])
			}
		} else if !tlsConfig.RootCAs.AppendCertsFromPEM(caBytes) {
			return nil, fmt.Errorf("failed to add pem formated ca certs")
		}
	}

	if len(c.Key) != 0 && len(c.Cert) != 0 {
		cert, err := tls.X509KeyPair(c.Cert, c.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to create x509 pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
