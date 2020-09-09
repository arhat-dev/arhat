// +build grpc

package cmd

import (
	"github.com/spf13/pflag"
	"strings"
)

func keepNecessaryConnectivityFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(f *pflag.Flag) {
		switch {
		case strings.HasPrefix(f.Name, "conn.coap"), strings.HasPrefix(f.Name, "conn.mqtt"):
			f.Hidden = true
		}
	})
}
