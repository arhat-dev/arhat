package types

import "io"

type AbbotExecFunc func(subCmd []string, output io.Writer) error

type NetworkClient interface {
	CreateResolvConf(nameservers, searches, options []string) ([]byte, error)
}
