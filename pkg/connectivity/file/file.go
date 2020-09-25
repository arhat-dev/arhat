package file

import (
	"fmt"
	"os"
	"strings"

	"arhat.dev/aranya-proto/aranyagopb"
	"github.com/fsnotify/fsnotify"

	"arhat.dev/arhat/pkg/connectivity"
	"arhat.dev/arhat/pkg/types"
)

func init() {
	connectivity.Register("file",
		aranyagopb.CONNECTIVITY_MODE_CLIENT, NewFileDeviceConnectivity)
}

func NewFileDeviceConnectivity(
	target string,
	params map[string]string,
	tlsConfig *aranyagopb.TLSConfig,
) (types.Connectivity, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("directory not supported")
	}

	var (
		writeFile string
		writeMode int
		op        fsnotify.Op
	)
	for k, v := range params {
		switch k {
		case "write_file":
			writeFile = v
		case "write_mode":
			modes := strings.Split(v, ",")
			for _, m := range modes {
				writeMode |= map[string]int{
					"rw":     os.O_RDWR,
					"w":      os.O_WRONLY,
					"append": os.O_APPEND,
					"trunc":  os.O_TRUNC,
					"sync":   os.O_SYNC,
					"excl":   os.O_EXCL,
					"create": os.O_CREATE,
				}[m]
			}
		case "watch_events":
			events := strings.Split(v, ",")
			for _, e := range events {
				op |= map[string]fsnotify.Op{
					"create": fsnotify.Create,
					"write":  fsnotify.Write,
					"remove": fsnotify.Remove,
					"rename": fsnotify.Rename,
					"chmod":  fsnotify.Chmod,
				}[e]
			}
		}
	}

	if writeFile == "" {
		return nil, fmt.Errorf("no write file specified")
	}

	return &fileDeviceConnectivity{
		watchFile: target,
		watcher:   w,
		writeMode: writeMode,
		reactOp:   op,
	}, nil
}

var _ types.Connectivity = &fileDeviceConnectivity{}

type fileDeviceConnectivity struct {
	watchFile string
	// writeFile string
	watcher *fsnotify.Watcher

	writeMode int
	reactOp   fsnotify.Op
}

func (s *fileDeviceConnectivity) Connect() error {
	return nil
}

func (s *fileDeviceConnectivity) doReadAfterWrite(params map[string]string) ([][]byte, error) {
	return nil, nil
}

func (s *fileDeviceConnectivity) Operate(params map[string]string, data []byte) ([][]byte, error) {
	return s.doReadAfterWrite(params)
}

func (s *fileDeviceConnectivity) CollectMetrics(params map[string]string) ([]*types.DeviceMetricValue, error) {
	//return s.doReadAfterWrite(params)
	return nil, nil
}

func (s *fileDeviceConnectivity) Close() error {
	return nil
}
