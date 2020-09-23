package network

import (
	"bytes"
	"fmt"
	"io"
	"text/template"

	"arhat.dev/arhat/pkg/types"
)

const resolvConfTemplate = `# resolv.conf generated by arhat
{{ range .Servers -}}
nameserver {{ . }}
{{ end }}
search {{- range .Searches }} {{ . }} {{- end -}}

{{- if gt (len .Options) 0 -}}
options {{- range .Options }} {{ . }} {{- end -}}
{{ end }}
`

func NewNetworkClient(exec types.AbbotExecFunc) *Client {
	return &Client{
		execAbbot: exec,
	}
}

type Client struct {
	execAbbot func(subCmd []string, output io.Writer) error
}

func (c *Client) CreateResolvConf(nameservers, searches, options []string) ([]byte, error) {
	resolvTemplate, err := template.New("").Parse(resolvConfTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resolv.conf template")
	}

	if len(nameservers) == 0 {
		nameservers = []string{"::1", "127.0.0.1"}
	}

	if len(searches) == 0 {
		searches = []string{"."}
	}

	buf := new(bytes.Buffer)
	err = resolvTemplate.Execute(buf, map[string][]string{
		"Servers":  nameservers,
		"Searches": searches,
		"Options":  options,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute resolve.conf template")
	}

	return buf.Bytes(), nil
}
