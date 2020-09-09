// +build !mqtt,!grpc,!coap

package cmd

import (
	"github.com/spf13/pflag"
)

func keepNecessaryConnectivityFlags(flags *pflag.FlagSet) {}
