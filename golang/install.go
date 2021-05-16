package golang

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	devctlpath2 "github.com/alex-held/devctl/pkg/devctlpath"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/pkg/errors"

	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	"github.com/alex-held/devctl/pkg/ui/taskrunner"
)

type Renamer func(p string) string

type GoInstallCmd struct {
	Pather  devctlpath2.Pather
	Runtime *sysutils.DefaultRuntimeInfoGetter
	Fs      vfs.VFS
}

func (cmd *GoInstallCmd) SetPather(feeder PatherFeeder) {
	cmd.Pather = feeder()
}

func (cmd *GoInstallCmd) SetFsFeeder(feeder FileSystemFeeder) {
	cmd.Fs = feeder()
}

func (cmd *GoInstallCmd) AsTasker(version string) taskrunner.Tasker {

	archiveName := FormatGoArchiveArtifactName(cmd.Runtime.Get(), version)
	archivePath := cmd.Pather.Download("go", version, archiveName)
	installPath := cmd.Pather.SDK("go", version)

	return &taskrunner.ConditionalTask{
		Description: "installing go sdk %s into the go sdk directory",
		Action: func(ctx context.Context) error {
			archive, err := cmd.Fs.OpenFile(archivePath, os.O_RDWR, os.ModePerm)

			if err != nil {
				return errors.Wrapf(err, "failed to open go sdk archive=%s\n", archivePath)
			}
			err = cmd.Fs.MkdirAll(installPath, os.ModePerm)
			if err != nil {
				return errors.Wrapf(err, "failed to Extract go sdk %s; dest=%s; archive=%s\n", version, installPath, archivePath)
			}
			err = UnTarGzip(archive, installPath, GoSDKUnarchiveRenamer(), cmd.Fs)
			if err != nil {
				return errors.Wrapf(err, "failed to Extract go sdk %s; dest=%s; archive=%s\n", version, installPath, archivePath)
			}
			return nil
		},
		ShouldExecute: func() bool {
			// dont run installer if the version is already installed
			exists, _ := cmd.Fs.Exists(installPath)
			return !exists
		},
	}
}

func (cmd *GoInstallCmd) PluginName() string {
	return GoInstallCmdName
}

func (cmd *GoInstallCmd) CmdName() string {
	return "install"
}

func (cmd *GoInstallCmd) ExecuteCommand(ctx context.Context, root string, args []string) (err error) {
	version := args[1]
	task := cmd.AsTasker(version)
	return task.Task(ctx)
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
