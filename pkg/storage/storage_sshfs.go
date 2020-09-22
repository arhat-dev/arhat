/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/log"

	"arhat.dev/arhat/pkg/conf"
	"arhat.dev/arhat/pkg/constant"
	"arhat.dev/arhat/pkg/storage/storageutil"
	"arhat.dev/arhat/pkg/types"
)

func NewSSHFSStorage(appCtx context.Context, config *conf.StorageConfig) (types.Storage, error) {
	sshfsPath, err := storageutil.Lookup("sshfs", config.LookupPaths)
	if err != nil {
		return nil, err
	}

	fuseUnmountPath, err := storageutil.LookupUnmountUtil(config.LookupPaths, true)
	if err != nil {
		return nil, err
	}

	// validate args
	args := config.Args["sshfs"]
	if len(args) < 2 {
		return nil, fmt.Errorf("expect at least 2 args")
	}

	valid := true
	count := 0
	// first arg MUST include env ref REMOTE_PATH
	os.Expand(args[0], func(s string) string {
		count++
		if s != constant.StorageArgEnvRemotePath {
			valid = false
		}
		return ""
	})
	if !valid || count > 1 {
		return nil, fmt.Errorf("first arg invalid, must include one $%s", constant.StorageArgEnvRemotePath)
	}

	count = 0
	// second arg MUST include env ref LOCAL_PATH
	localPath := os.Expand(args[1], func(s string) string {
		count++
		if s != constant.StorageArgEnvLocalPath {
			valid = false
		}
		return ""
	})
	if !valid || count > 1 || localPath != "" {
		return nil, fmt.Errorf("second arg invalid, must be $%s", constant.StorageArgEnvLocalPath)
	}

	blacklistOptionsPrefix := []string{
		"password_stdin",
		"IdentityFile",
		"-F",
		"slave",
	}

	// other args should not contain any env ref

	for _, arg := range args[2:] {
		os.Expand(arg, func(s string) string {
			valid = false
			return ""
		})

		if !valid {
			return nil, fmt.Errorf("invalid arg %s: should not contain env ref", arg)
		}

		for _, prefix := range blacklistOptionsPrefix {
			if strings.HasPrefix(arg, prefix) {
				return nil, fmt.Errorf("option %s is not allowed", arg)
			}
		}
	}

	return &sshfsStorage{
		ctx: appCtx,
		log: log.Log.WithName("sshfs"),

		stdoutFile: config.StdoutFile,
		stderrFile: config.StderrFile,
		wait:       config.ProcessCheckTimeout,

		fuseUnmountPath: fuseUnmountPath,
		sshfsPath:       sshfsPath,
		args:            args,

		sshfsInstances: make(map[string]*os.Process),
		mu:             new(sync.Mutex),

		sshIdentityStore: new(atomic.Value),
	}, nil
}

type sshfsStorage struct {
	ctx context.Context
	log log.Interface

	stdoutFile string
	stderrFile string
	wait       time.Duration

	fuseUnmountPath string
	sshfsPath       string
	args            []string

	sshfsInstances map[string]*os.Process
	mu             *sync.Mutex

	sshIdentityStore *atomic.Value
}

func (s *sshfsStorage) Name() string {
	return constant.StorageBackendSSHFS
}

func (s *sshfsStorage) getIdentityBytes() []byte {
	if d := s.sshIdentityStore.Load(); d != nil {
		return d.([]byte)
	}
	return nil
}

func (s *sshfsStorage) Mount(remotePath, mountPoint string, onExited types.StorageFailureHandleFunc) error {
	var identityFile string
	p := s.getIdentityBytes()
	if p == nil {
		return fmt.Errorf("identity not provided")
	}

	f, err := ioutil.TempFile(os.TempDir(), "*.id")
	if err != nil {
		return fmt.Errorf("failed to create temporary identity file: %w", err)
	}

	identityFile = f.Name()

	err = func() error {
		defer func() { _ = f.Close() }()

		n, err2 := f.Write(p)
		if err2 != nil || n != len(p) {
			return fmt.Errorf("failed to write identity content: %w", err2)
		}

		return f.Chmod(0400)
	}()

	if err != nil {
		return err
	}

	command, err := storageutil.ResolveCommand(s.sshfsPath, s.args, remotePath, mountPoint)
	if err != nil {
		return err
	}

	command = append(command,
		"-o", fmt.Sprintf("IdentityFile=%s", identityFile), // ssh identity
		"-o", "reconnect", // auto reconnect
		"-f", // foreground
	)

	// make sure no fuse volume has been mounted
	_ = s.Unmount(mountPoint)
	process, err := storageutil.Execute(command, s.stdoutFile, s.stderrFile, s.wait, func(err error) {
		// best effort
		_ = os.Remove(identityFile)
		_ = s.Unmount(mountPoint)

		onExited(remotePath, mountPoint, err)
	})

	logger := s.log.WithFields(log.String("mountPoint", mountPoint))
	if err == nil {
		s.mu.Lock()
		s.sshfsInstances[mountPoint] = process
		s.mu.Unlock()

		logger.V("mounted")
	} else {
		logger.I("failed to mount")
	}

	return err
}

func (s *sshfsStorage) Unmount(mountPoint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	baseLogger := s.log.WithFields(log.String("mountPoint", mountPoint))
	baseLogger.V("unmounting fuse volume")
	err := storageutil.Unmount(s.fuseUnmountPath, mountPoint)
	if err != nil {
		baseLogger.I("failed to unmount fuse volume", log.Error(err))
	}

	if process, ok := s.sshfsInstances[mountPoint]; ok {
		logger := baseLogger.WithFields(log.Int("pid", process.Pid))
		logger.V("killing process")
		e := process.Kill()
		if e != nil {
			logger.I("failed to kill process", log.Error(e))
		}

		delete(s.sshfsInstances, mountPoint)
		if err == nil {
			return e
		}
	}

	return err
}

func (s *sshfsStorage) SetCredentials(options *aranyagopb.CredentialEnsureCmd) {
	s.sshIdentityStore.Store(options.SshPrivateKey)
}
