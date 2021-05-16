package golang

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alex-held/devctl-plugins/pkg/devctlog"
	"github.com/alex-held/devctl-plugins/pkg/devctlpath"
	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type Renamer func(p string) string

type GoInstallCmd struct {
	Pather  devctlpath.Pather
	Runtime *sysutils.DefaultRuntimeInfoGetter
	Fs      vfs.VFS
	Logger  devctlog.Logger
}

func (cmd *GoInstallCmd) SetLogger(feeder LoggerFeeder) {
	cmd.Logger = feeder()
}

func (cmd *GoInstallCmd) SetPather(feeder PatherFeeder) {
	cmd.Pather = feeder()
}

func (cmd *GoInstallCmd) SetFsFeeder(feeder FileSystemFeeder) {
	cmd.Fs = feeder()
}

func (cmd *GoInstallCmd) PluginName() string {
	return GoInstallCmdName
}

func (cmd *GoInstallCmd) CmdName() string {
	return "install"
}

func UnTarGzip(file io.Reader, target string, renamer Renamer, v vfs.VFS) error {
	gr, _ := gzip.NewReader(file)
	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		filename := header.Name
		if renamer != nil {
			filename = renamer(filename)
		}

		p := filepath.Join(target, filename)
		fi := header.FileInfo()

		if fi.IsDir() {
			if e := v.MkdirAll(p, fi.Mode()); e != nil {
				return e
			}
			continue
		}
		file, err := v.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(file, tr)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func GoSDKUnarchiveRenamer() Renamer {
	return func(p string) string {
		parts := strings.Split(p, string(filepath.Separator))
		parts = parts[1:]
		newPath := strings.Join(parts, string(filepath.Separator))
		return newPath
	}
}
