package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"arhat.dev/aranya-proto/aranyagopb/aranyagoconst"
	"github.com/dgrijalva/jwt-go"

	"arhat.dev/arhat/pkg/client/clientutil"
	"arhat.dev/arhat/pkg/conf"
)

type ConnectivityMQTT struct {
	clientutil.ConnectivityCommonConfig `json:",inline" yaml:",inline"`

	Version            string             `json:"version" yaml:"version"`
	Variant            string             `json:"variant" yaml:"variant"`
	Transport          string             `json:"transport" yaml:"transport"`
	TopicNamespaceFrom conf.ValueFromSpec `json:"topicNamespaceFrom" yaml:"topicNamespaceFrom"`
	ClientID           string             `json:"clientID" yaml:"clientID"`
	Username           string             `json:"username" yaml:"username"`
	Password           string             `json:"password" yaml:"password"`
	Keepalive          int32              `json:"keepalive" yaml:"keepalive"`
}

type ConnectivityMQTTConnectInfo struct {
	Username string
	Password string
	ClientID string

	MsgPubTopic       string
	CmdSubTopic       string
	CmdSubTopicHandle string
	WillPubTopic      string
	SupportRetain     bool

	MaxPayloadSize int
	TLSConfig      *tls.Config
}

func (c *ConnectivityMQTT) GetConnectInfo() (*ConnectivityMQTTConnectInfo, error) {
	result := new(ConnectivityMQTTConnectInfo)

	result.MaxPayloadSize = c.MaxPayloadSize

	topicNamespace, err := c.TopicNamespaceFrom.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get topic namespace value: %w", err)
	}

	variant := strings.ToLower(c.Variant)
	switch variant {
	case aranyagoconst.VariantAzureIoTHub:
		if result.MaxPayloadSize <= 0 {
			result.MaxPayloadSize = aranyagoconst.MaxAzureIoTHubD2CDataSize
		}

		deviceID := c.ClientID
		result.ClientID = deviceID

		propertyBag, err2 := url.ParseQuery(topicNamespace)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse property bag: %w", err2)
		}
		propertyBag["dev"] = []string{deviceID}
		propertyBag["arhat"] = []string{""}
		// azure iot-hub topics
		result.MsgPubTopic = fmt.Sprintf("devices/%s/messages/events/%s", deviceID, propertyBag.Encode())
		result.WillPubTopic = result.MsgPubTopic
		result.CmdSubTopic = fmt.Sprintf("devices/%s/messages/devicebound/#", deviceID)
		result.CmdSubTopicHandle = fmt.Sprintf("devices/%s/messages/devicebound/.*", deviceID)

		result.Username = fmt.Sprintf("%s/%s/?api-version=2018-06-30", c.Endpoint, deviceID)
		// Password is set to SAS token if not using mTLS
		result.Password = c.Password
	case aranyagoconst.VariantGCPIoTCore:
		if !c.TLS.Enabled || c.TLS.Key == "" {
			return nil, fmt.Errorf("no private key found")
		}

		if c.TLS.Cert != "" {
			return nil, fmt.Errorf("cert file must be empty")
		}

		if result.MaxPayloadSize <= 0 {
			result.MaxPayloadSize = aranyagoconst.MaxGCPIoTCoreD2CDataSize
		}

		result.ClientID = c.ClientID
		parts := strings.Split(c.ClientID, "/")
		if len(parts) != 8 {
			return nil, fmt.Errorf("expect 8 sections in client id but found %d", len(parts))
		}

		// second section is project id
		projectID := parts[1]
		claims := jwt.StandardClaims{
			Audience: projectID,
			IssuedAt: time.Now().Unix(),
			// valid for half a day (max value is 24 hr)
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		}

		keyBytes, err2 := ioutil.ReadFile(c.TLS.Key)
		if err2 != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err2)
		}

		var (
			key        interface{}
			signMethod jwt.SigningMethod
		)

		block, _ := pem.Decode(keyBytes)
		switch block.Type {
		case "EC PRIVATE KEY":
			signMethod = jwt.SigningMethodES256
			key, err = x509.ParseECPrivateKey(block.Bytes)
		case "RSA PRIVATE KEY":
			signMethod = jwt.SigningMethodRS256
			key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		default:
			return nil, fmt.Errorf("unsupported private key algorithm")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		token := jwt.NewWithClaims(signMethod, claims)
		jwtToken, err2 := token.SignedString(key)
		if err2 != nil {
			return nil, fmt.Errorf("failed to sign jwt token: %w", err2)
		}

		// last section is the device id
		deviceID := parts[7]
		result.MsgPubTopic = fmt.Sprintf("/devices/%s/events", deviceID)
		if topicNamespace != "" {
			result.MsgPubTopic = fmt.Sprintf("/devices/%s/events/%s", deviceID, topicNamespace)
		}
		result.CmdSubTopic = fmt.Sprintf("/devices/%s/commands/#", deviceID)
		result.CmdSubTopicHandle = fmt.Sprintf("/devices/%s/commands.*", deviceID)
		result.WillPubTopic = fmt.Sprintf("/devices/%s/state", deviceID)
		result.Password = jwtToken
	case aranyagoconst.VariantAWSIoTCore:
		if !c.TLS.Enabled || c.TLS.Cert == "" || c.TLS.Key == "" {
			return nil, fmt.Errorf("tls cert key pair must be provided for aws-iot-core")
		}

		if result.MaxPayloadSize <= 0 {
			result.MaxPayloadSize = aranyagoconst.MaxAwsIoTCoreD2CDataSize
		}

		result.ClientID = c.ClientID
		result.CmdSubTopic, result.MsgPubTopic, result.WillPubTopic = aranyagoconst.MQTTTopics(topicNamespace)
		result.CmdSubTopicHandle = result.CmdSubTopic
	case "", "standard":
		if result.MaxPayloadSize <= 0 {
			result.MaxPayloadSize = aranyagoconst.MaxMQTTDataSize
		}

		result.Username = c.Username
		result.Password = c.Password
		result.ClientID = c.ClientID
		result.SupportRetain = true

		result.CmdSubTopic, result.MsgPubTopic, result.WillPubTopic = aranyagoconst.MQTTTopics(topicNamespace)
		result.CmdSubTopicHandle = result.CmdSubTopic
	default:
		return nil, fmt.Errorf("unsupported variant type")
	}

	result.TLSConfig, err = c.TLS.GetTLSConfig(false)
	if err != nil {
		return nil, fmt.Errorf("failed to create client tls config: %w", err)
	}

	if variant == aranyagoconst.VariantAWSIoTCore {
		result.TLSConfig.NextProtos = []string{"x-amzn-mqtt-ca"}
	}

	return result, nil
}
