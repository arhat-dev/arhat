package network

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"arhat.dev/abbot-proto/abbotgopb"
)

func encodeRequest(req *abbotgopb.Request) (string, error) {
	data, err := req.Marshal()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// nolint:unparam
func (c *Client) doRequest(req *abbotgopb.Request) (*abbotgopb.Response, error) {
	encodedReq, err := encodeRequest(req)
	if err != nil {
		return nil, err
	}

	output := new(bytes.Buffer)
	err = c.execAbbot([]string{"request", encodedReq}, output)
	result := output.String()

	if err != nil {
		if result != "" {
			return nil, fmt.Errorf("%s: %w", result, err)
		}

		return nil, err
	}

	respBytes, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		// not base64 encoded, error happened
		return nil, fmt.Errorf(result)
	}

	resp := new(abbotgopb.Response)
	err = resp.Unmarshal(respBytes)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
