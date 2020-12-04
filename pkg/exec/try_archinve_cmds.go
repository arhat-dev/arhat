// +build !noexectry,!noexectry_archive
// +build !js

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
	"path/filepath"
	"strconv"
	"strings"

	"arhat.dev/pkg/exechelper"
	"arhat.dev/pkg/wellknownerrors"
	"github.com/mholt/archiver/v3"
	"github.com/spf13/pflag"
)

const (
	binTar   = "tar"
	binZip   = "zip"
	binUnzip = "unzip"
	binUnrar = "unrar"
)

func init() {
	for _, bin := range []string{binTar, binZip, binUnzip, binUnrar} {
		tryCommands[bin] = tryArchiveCmd
	}
}

// tryArchiveCmd handle tar/zip/unzip/rar/unrar command execution
// this is meant to help user copy
// files to systems without a full tar implementation (e.g. busybox, windows).
//
// tar: It will ignore following tar command flags if not using host tar
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
func tryArchiveCmd(
	stdin io.Reader,
	stdout, stderr io.Writer,
	command []string, _ bool,
) (Cmd, error) {
	var opts = new(archiveOpts)

	opts.rawArgs = command[1:]

	bin := filepath.Base(command[0])
	switch bin {
	case binTar:
		if len(opts.rawArgs) > 0 && !strings.HasPrefix(opts.rawArgs[0], "-") {
			opts.rawArgs = append([]string{"-" + opts.rawArgs[0]}, opts.rawArgs[1:]...)
		}
	default:
	}

	flags := opts.flags(bin)
	if flags == nil {
		return nil, wellknownerrors.ErrNotSupported
	}

	err := flags.Parse(opts.rawArgs)
	if err != nil {
		return nil, err
	}

	opts.newBundleSourcePaths = flags.Args()

	if err = opts.resolveAndValidate(bin); err != nil {
		return nil, err
	}

	if stdout == nil {
		stdout = ioutil.Discard
	}

	if stderr == nil {
		stderr = ioutil.Discard
	}

	return &flexCmd{
		do: func() error {
			return runArchiveCmd(command, stdin, stdout, stderr, opts)
		},
	}, nil
}

// nolint:gocyclo
func runArchiveCmd(
	command []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	opts *archiveOpts,
) error {
	var (
		err error
		bin = filepath.Base(command[0])

		// extract
		extractFile = opts.extractBundlePath

		// create bundle
		arc         archiver.Archiver
		dataOut     = stdout
		tmpDestR    io.ReadCloser
		tmpDestFile = ""
		resetDest   = func(erase bool) {}
	)

	if opts.extract {
		if stdin != nil && extractFile == "-" {
			// write data to a temporary file, do not read directly from stdin
			// in case host tar failed and we will only have partial data remain
			extractFile, err = func() (string, error) {
				tempFile, err2 := ioutil.TempFile(opts.extractToDir, "*."+bin)
				if err2 != nil {
					return "", err2
				}
				defer func() { _ = tempFile.Close() }()

				n, err2 := tempFile.ReadFrom(stdin)
				if n == 0 && err2 != nil {
					return tempFile.Name(), err2
				}
				_ = tempFile.Sync()

				return tempFile.Name(), nil
			}()

			if extractFile != "" {
				defer func() {
					if err == nil {
						// this file was handled by host bin, not useful anymore
						_ = os.Remove(extractFile)
					}
				}()
			}
			if err != nil {
				return err
			}
		}
	} else {
		raw, fileExt := opts.getHandler(bin)
		var ok bool
		arc, ok = raw.(archiver.Archiver)
		if !ok {
			_, _ = fmt.Fprintf(stderr, "invalid archive options\n")
			return fmt.Errorf("invalid opts")
		}

		if opts.newBundleDest == "-" {
			// no destination, use stdout as destionation
			var f *os.File
			f, err = ioutil.TempFile(".", "*"+fileExt)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "failed to create temp archive file: %v\n", err)
				return err
			}

			tmpDestFile = f.Name()
			defer func() {
				_ = f.Close()
				_ = os.Remove(tmpDestFile)
			}()

			dataOut = f
			tmpDestR = f
			resetDest = func(erase bool) {
				if erase {
					_ = f.Truncate(0)
				}
				_, _ = f.Seek(0, io.SeekStart)
			}
		}
	}

	// try host executable first

	// maybe the first arg is an absolute path outside PATH
	// find it first
	var hostBin, thisBin string
	hostBin, err = exec.LookPath(command[0])
	if err == nil {
		thisBin, err = filepath.Abs(os.Args[0])
	}
	err = func() error {
		if err != nil {
			return err
		}

		if hostBin == thisBin {
			return fmt.Errorf("same bin")
		}

		args, err2 := opts.buildArgs(bin, extractFile)
		if err2 != nil {
			return err2
		}

		startedCmd, err2 := exechelper.Do(exechelper.Spec{
			Command: append([]string{hostBin}, args...),
			Stdin:   nil,
			Stdout:  dataOut, // in case creating bundle to stdout
			Stderr:  stderr,
		})
		if err2 != nil {
			resetDest(true)
			return err2
		}

		_, err2 = startedCmd.Wait()
		if err2 != nil {
			resetDest(true)
			_, _ = fmt.Fprintf(stderr, "host %s failed: %v\n", bin, err2)
			return err2
		}

		if tmpDestR != nil {
			// rewrite data to stdout when successful
			resetDest(false)
			_, err2 = io.Copy(stdout, tmpDestR)
			if err2 != nil && err2 != io.EOF {
				resetDest(true)
				return err2
			}

			return nil
		}

		return nil
	}()
	if err == nil {
		// handled by host command
		return nil
	}

	// not handled by host command, try archiver

	// show notice for unsupported options
	switch bin {
	case binTar:
		// host tar failed, use incomplete embedded tar handler
		_, _ = stderr.Write([]byte("using embedded tar handler:\n"))
		_, _ = stderr.Write([]byte("    option --no-same-owner and --no-same-permissions will be ignored\n"))
	case binUnrar:
	case binZip:
	case binUnzip:
	}

	if opts.extract {
		f, err2 := os.Open(extractFile)
		if err2 != nil {
			return err2
		}

		u, err2 := func() (archiver.Unarchiver, error) {
			defer func() { _ = f.Close() }()
			u, err3 := archiver.ByHeader(f)
			if err3 != nil {
				// cannot create unarchiver, use opts to determine
				var (
					raw interface{}
					ok  bool
				)

				raw, _ = opts.getHandler(bin)
				u, ok = raw.(archiver.Unarchiver)
				if !ok {
					return nil, fmt.Errorf("failed to detect format: %w", err3)
				}
			}

			return u, nil
		}()
		if err2 != nil {
			return err2
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

		err = u.Unarchive(extractFile, opts.extractToDir)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "embedded tar failed: %v\n", err)
		}
		return err
	}

	// create bundle and write to destination
	if tmpDestFile == "" {
		// has destination
		return arc.Archive(opts.newBundleSourcePaths, opts.newBundleDest)
	}

	_ = tmpDestR.Close()

	err = arc.Archive(opts.newBundleSourcePaths, tmpDestFile)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to archive files: %v\n", err)
		return err
	}

	f, err := os.Open(tmpDestFile)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to open temp archive file: %v\n", err)
		return err
	}

	_, err = io.Copy(stdout, f)
	if err != nil && err != io.EOF {
		_, _ = fmt.Fprintf(stderr, "failed to copy result to stdout: %v\n", err)
		return err
	}

	return nil
}

// nolint:maligned
type archiveOpts struct {
	// create bundle
	create               bool
	newBundleSourcePaths []string
	newBundleDest        string

	// extract bundle
	extract           bool
	extractBundlePath string
	extractToDir      string
	updateMTime       bool
	noSameOwner       bool
	noSamePermission  bool

	customCompression string

	noCompression bool
	// standard tar compression and supported by archiver
	xz    bool
	gzip  bool
	bzip2 bool
	zstd  bool

	// standard tar compression but not supported by archiver
	lzma bool
	lzip bool
	lzop bool

	compressionLevel  int
	compressionLevels []bool

	rawArgs []string
}

func (o *archiveOpts) flags(bin string) *pflag.FlagSet {
	flags := pflag.NewFlagSet("many-archive-bin", pflag.ContinueOnError)
	switch bin {
	case binTar:
		flags.BoolVarP(&o.extract, "extract", "x", false, "")
		flags.BoolVarP(&o.create, "create", "c", false, "")
		flags.StringVarP(&o.extractBundlePath, "file", "f", "", "")
		flags.StringVarP(&o.extractToDir, "directory", "C", "", "")
		{
			// compression methods
			flags.StringVarP(&o.customCompression, "use-compress-program", "I", "", "")
			flags.BoolVarP(&o.bzip2, "bzip2", "j", false, "")
			flags.BoolVarP(&o.xz, "xz", "J", false, "")
			flags.BoolVarP(&o.gzip, "gzip", "z", false, "")
			flags.BoolVar(&o.gzip, "gnuzip", false, "")
			flags.BoolVar(&o.gzip, "ungzip", false, "")
			flags.BoolVar(&o.zstd, "zstd", false, "")
		}
		{
			// options unsuppoted but required by kubectl cp
			flags.BoolVarP(&o.updateMTime, "touch", "m", false, "")
			flags.BoolVar(&o.noSameOwner, "no-same-owner", false, "")
			flags.BoolVar(&o.noSamePermission, "no-same-permissions", false, "")
		}
	case binZip:
		o.create = true
		o.extract = false

		flags.StringVarP(&o.newBundleDest, "output-file", "O", "", "")
		{
			// compression
			flags.StringVarP(&o.customCompression, "compression-method", "Z", "", "")
			o.compressionLevels = make([]bool, 10)
			for i := 0; i < len(o.compressionLevels); i++ {
				flags.BoolVarP(&o.compressionLevels[i],
					"compression-level-"+strconv.Itoa(i), strconv.Itoa(i), false, "",
				)
			}
		}
	case binUnzip:
		o.create = false
		o.extract = true

		flags.StringVarP(&o.extractToDir, "directory", "d", "", "")
	case binUnrar:
		o.create = false
		o.extract = true
	}

	return flags
}

func (o *archiveOpts) resolveAndValidate(bin string) error {
	if o.create == o.extract {
		return fmt.Errorf("create/extrea not set")
	}

	for i, set := range o.compressionLevels {
		if set {
			o.compressionLevel = i
		}
	}

	switch bin {
	case binTar:
		if o.extract {
			switch {
			case o.extractBundlePath == "":
				// no file to extract
				return fmt.Errorf("extract target not set")
			default:
			}
		} else {
			o.newBundleDest = o.extractBundlePath
			switch {
			case len(o.newBundleSourcePaths) == 0:
				// no source to create bundle
				return fmt.Errorf("create target sources not set")
			default:
			}
		}

		return nil
	case binZip:
		switch o.customCompression {
		case "store":
			o.noCompression = true
		case "deflate", "":
			// default option in zip
		case "bzip2":
			o.bzip2 = true
		case "zstd":
			o.zstd = true
		case "lzma":
			o.lzma = true
		case "xz":
			o.xz = true
		default:
			return fmt.Errorf("unsupported compression method: %s", o.customCompression)
		}

		return nil
	case binUnzip:
		return nil
	case binUnrar:
		return nil
	}

	return wellknownerrors.ErrNotSupported
}

func (o *archiveOpts) buildArgs(bin, bundlePath string) ([]string, error) {
	var args []string

	var (
		tarCompressionArgs []string
	)

	switch o.customCompression {
	case "snappy":
		tarCompressionArgs = []string{"-I", "snappy"}
	case "brotli":
		tarCompressionArgs = []string{"-I", "brotli"}
	case "lz4":
		tarCompressionArgs = []string{"-I"}
		if o.compressionLevel >= 0 {
			tarCompressionArgs = append(
				tarCompressionArgs,
				"lz4 -"+strconv.Itoa(o.compressionLevel),
			)
		} else {
			// use default compression level
			tarCompressionArgs = append(
				tarCompressionArgs,
				"lz4",
			)
		}
	case "":
		switch {
		case o.lzma:
			tarCompressionArgs = []string{"--lzma"}
		case o.lzip:
			tarCompressionArgs = []string{"--lzip"}
		case o.lzop:
			tarCompressionArgs = []string{"--lzop"}
		case o.gzip:
			tarCompressionArgs = []string{"-z"}
		case o.xz:
			tarCompressionArgs = []string{"-J"}
		case o.bzip2:
			tarCompressionArgs = []string{"-j"}
		case o.zstd:
			tarCompressionArgs = []string{"--zstd"}
		default:
			// no compression
		}
	default:
		// unexpected, should have been filtered when validating
		return nil, wellknownerrors.ErrInvalidOperation
	}

	if o.extract {
		switch bin {
		case binTar:
			args = []string{"-x"}
			args = append(args, tarCompressionArgs...)
			args = append(args, "-f", bundlePath)

			if o.extractToDir != "" {
				args = append(args, "-C", o.extractToDir)
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
		case binUnzip:
			args = []string{bundlePath}
			if o.extractToDir != "" {
				args = append(args, "-d", o.extractToDir)
			}

			return args, nil
		case binUnrar:
			return []string{bundlePath}, nil
		}
	} else {
		return o.rawArgs, nil
	}

	return args, nil
}

func (o *archiveOpts) getHandler(bin string) (io.Closer, string) {
	switch bin {
	case binTar:
		switch o.customCompression {
		case "snappy":
			tarSz := archiver.NewTarSz()

			tarSz.OverwriteExisting = true
			return tarSz, ".tar.sz"
		case "brotli":
			tarBrotli := archiver.NewTarBrotli()

			tarBrotli.OverwriteExisting = true
			return tarBrotli, ".tar.br"
		case "lz4":
			tarLz4 := archiver.NewTarLz4()

			tarLz4.OverwriteExisting = true
			if o.compressionLevel >= 0 {
				tarLz4.CompressionLevel = o.compressionLevel
			}
			return tarLz4, ".tar.lz4"
		case "":
			switch {
			case o.lzma, o.lzip, o.lzop:
				// need to try host bin
				return nil, ""
			case o.gzip:
				tarGz := archiver.NewTarGz()

				tarGz.OverwriteExisting = true
				if o.compressionLevel >= 0 {
					tarGz.CompressionLevel = o.compressionLevel
				}
				return tarGz, ".tar.gz"
			case o.xz:
				tarXz := archiver.NewTarXz()

				tarXz.OverwriteExisting = true
				return tarXz, ".tar.xz"
			case o.bzip2:
				tarBz2 := archiver.NewTarBz2()

				tarBz2.OverwriteExisting = true
				if o.compressionLevel >= 0 {
					tarBz2.CompressionLevel = o.compressionLevel
				}
				return tarBz2, ".tar.bz2"
			case o.zstd:
				tarZstd := archiver.NewTarZstd()

				tarZstd.OverwriteExisting = true
				return tarZstd, ".tar.zst"
			default:
				tar := archiver.NewTar()

				tar.OverwriteExisting = true
				return tar, ".tar"
			}
		default:
			// unexpected, should have been filtered when validating
			return nil, ""
		}
	case binUnrar:
		rar := archiver.NewRar()

		rar.OverwriteExisting = true
		return rar, ".rar"
	case binZip, binUnzip:
		zip := archiver.NewZip()

		zip.OverwriteExisting = true
		if o.compressionLevel >= 0 {
			zip.CompressionLevel = o.compressionLevel
		}

		switch {
		case o.noCompression:
			zip.FileMethod = archiver.Store
		case o.bzip2:
			zip.FileMethod = archiver.BZIP2
		case o.lzma:
			zip.FileMethod = archiver.LZMA
		case o.zstd:
			zip.FileMethod = archiver.ZSTD
		case o.xz:
			zip.FileMethod = archiver.XZ
		default:
			zip.FileMethod = archiver.Deflate
		}

		return zip, ".zip"
	default:
		return nil, ""
	}
}
