package conf

import "time"

type StorageConfig struct {
	// Name of backend used to mount remote volumes
	//
	// available backend option:
	//	- "" (disabled)
	//  - sshfs
	Backend             string        `json:"backend" yaml:"backend"`
	StdoutFile          string        `json:"stdoutFile" yaml:"stdoutFile"`
	StderrFile          string        `json:"stderrFile" yaml:"stderrFile"`
	ProcessCheckTimeout time.Duration `json:"processCheckTimeout" yaml:"processCheckTimeout"`

	// LookupPaths to lookup executables required by the backend
	// will default to $PATH if not set
	LookupPaths []string `json:"lookupPaths" yaml:"lookupPaths"`

	// Args with env var references for backend mount operation
	//
	// valid env var references are
	//	 - ${ARHAT_STORAGE_REMOTE_PATH}
	//   - ${ARHAT_STORAGE_LOCAL_PATH}
	Args map[string][]string `json:"args" yaml:"args"`
}
