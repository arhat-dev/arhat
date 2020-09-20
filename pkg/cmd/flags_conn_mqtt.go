// +build mqtt

package cmd

import (
	"strings"

	"github.com/spf13/pflag"
)

func keepNecessaryConnectivityFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(f *pflag.Flag) {
		switch {
		case strings.HasPrefix(f.Name, "conn.grpc"), strings.HasPrefix(f.Name, "conn.coap"):
			f.Hidden = true
		}
	})
}
