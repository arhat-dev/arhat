// +build !noperipheral

package arhatgopb

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/gogo/protobuf/proto"
)

func NewCmd(id, seq uint64, cmd proto.Marshaler) (*Cmd, error) {
	var kind CmdType
	switch cmd.(type) {
	case *PeripheralConnectCmd:
		kind = CMD_PERIPHERAL_CONNECT
	case *PeripheralOperateCmd:
		kind = CMD_PERIPHERAL_OPERATE
	case *PeripheralMetricsCollectCmd:
		kind = CMD_PERIPHERAL_COLLECT_METRICS
	case *PeripheralCloseCmd:
		kind = CMD_PERIPHERAL_CLOSE
	default:
		return nil, fmt.Errorf("unknown cmd: %v", cmd)
	}

	data, err := cmd.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cmd: %w", err)
	}

	return &Cmd{
		Kind:    kind,
		Id:      id,
		Seq:     seq,
		Payload: data,
	}, nil
}

func NewMsg(id, ack uint64, msg proto.Marshaler) (*Msg, error) {
	var kind MsgType
	switch msg.(type) {
	case *RegisterMsg:
		kind = MSG_REGISTER
	case *PeripheralOperationResultMsg:
		kind = MSG_PERIPHERAL_OPERATION_RESULT
	case *PeripheralMetricsMsg:
		kind = MSG_PERIPHERAL_METRICS
	case *DoneMsg:
		kind = MSG_DONE
	case *ErrorMsg:
		kind = MSG_ERROR
	case *PeripheralEventMsg:
		kind = MSG_PERIPHERAL_EVENTS
	default:
		return nil, fmt.Errorf("unknown msg: %v", msg)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg: %w", err)
	}

	return &Msg{
		Kind:    kind,
		Id:      id,
		Ack:     ack,
		Payload: data,
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
