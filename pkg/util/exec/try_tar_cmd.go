// +build !noexectry,!noexectry_tar,!js

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

package exec

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/pkg/wellknownerrors"
	"github.com/mholt/archiver/v3"
	"github.com/spf13/cobra"
)

func init() {
	tryCommands["tar"] = tryTarCmd
}

type untarOpts struct {
	targetDir        string
	updateMTime      bool
	noSameOwner      bool
	noSamePermission bool
}

func (o *untarOpts) buildArgs(src string) []string {
	args := []string{
		"-x", "-f", src,
	}

	if o.targetDir != "" {
		args = append(args, "-C", o.targetDir)
	}

	if o.updateMTime {
		args = append(args, "-m")
	}

	if o.noSameOwner {
		args = append(args, "--no-same-owner")
	}

	if o.noSamePermission {
		args = append(args, "--no-same-permissions")
	}

	return args
}

// tryTarCmd handle tar command execution, this is meant to help user copy
// files to systems without a full tar implementation (e.g. busyBox).
//
// Its will ignore following tar command flags if not using host tar
// executable:
//	- `-m`
//	- `--no-same-owner`
//	- `--no-same-permissions`
//
// returned error will be wellknownerrors.ErrNotSupported if the tar command is
// not doing any of following operation:
// 	- accepting stdin as input (-)
//	- designate target dir explicitly (-C)
//	- extract (-x)
func tryTarCmd(
	stdin io.Reader,
	stdout, stderr io.Writer,
	_ <-chan *aranyagopb.ContainerTerminalResizeCmd,
	command []string, _ bool,
) error {
	var (
		extract    bool
		sourceFile string
		opts       = new(untarOpts)
	)

	cmd := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := stdout
			if stderr != nil {
				w = stderr
			}

			if w == nil {
				w = ioutil.Discard
			}

			return runTarCmd(stdin, w, opts)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&extract, "extract", "x", false, "")
	flags.StringVarP(&sourceFile, "file", "f", "", "")
	flags.StringVarP(&opts.targetDir, "change-dir", "C", "", "")
	flags.BoolVarP(&opts.updateMTime, "update-mtime", "m", false, "")
	flags.BoolVar(&opts.noSameOwner, "no-same-owner", false, "")
	flags.BoolVar(&opts.noSamePermission, "no-same-permissions", false, "")

	cmd.SetArgs(command[1:])
	if err := cmd.ParseFlags(command[1:]); err != nil {
		return wellknownerrors.ErrNotSupported
	}

	if !extract || sourceFile != "-" || opts.targetDir == "" {
		return wellknownerrors.ErrNotSupported
	}

	return cmd.Execute()
}

func runTarCmd(stdin io.Reader, w io.Writer, opts *untarOpts) error {
	// write data to the temp file
	srcFile, err := func() (string, error) {
		tempFile, err := ioutil.TempFile(os.TempDir(), "*.tar")
		if err != nil {
			return "", err
		}
		defer func() { _ = tempFile.Close() }()

		n, err := io.Copy(tempFile, stdin)
		if n == 0 && err != nil {
			return "", err
		}
		_ = tempFile.Sync()

		return tempFile.Name(), nil
	}()
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(srcFile) }()

	tarExec, err := exec.LookPath("tar")
	if err == nil {
		cmd := exec.Command(tarExec, opts.buildArgs(srcFile)...)
		data, err2 := cmd.CombinedOutput()

		if len(data) > 0 {
			_, _ = fmt.Fprintln(w, string(data))
		}

		if err2 != nil {
			_, _ = fmt.Fprintf(w, "host tar failed: %v\n", err2)
		}

		return nil
	}

	// host tar failed, use incomplete embedded tar handler
	_, _ = w.Write(
		[]byte("using embedded tar handler: option -m, --no-same-owner and --no-same-permissions will be ignored\n"))

	f, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	u, err := func() (archiver.Unarchiver, error) {
		defer func() { _ = f.Close() }()
		u, err2 := archiver.ByHeader(f)
		if err2 != nil {
			return nil, err2
		}

		return u, nil
	}()
	if err != nil {
		return err
	}

	// TODO: implement custom archiver.WalkFunc to support `-m`, `--no-same-permissions`, `--no-same-owner`
	switch un := u.(type) {
	case *archiver.Tar:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	case *archiver.TarBrotli:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	case *archiver.TarBz2:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	case *archiver.TarGz:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	case *archiver.TarLz4:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	case *archiver.TarSz:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	case *archiver.TarXz:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	case *archiver.TarZstd:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	case *archiver.Zip:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	case *archiver.Rar:
		un.OverwriteExisting = true
		defer func() { _ = un.Close() }()
		//return un.Walk(srcFile, doUnarchiveFile(opts))
	default:
		return fmt.Errorf("unexpected tarball format: %w", wellknownerrors.ErrNotSupported)
	}

	err = u.Unarchive(srcFile, opts.targetDir)
	if err != nil {
		_, _ = fmt.Fprintf(w, "embedded tar failed: %v\n", err)
	}
	return err
}
