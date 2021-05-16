package golang

import (
	"fmt"
	"os"
	"path"

	"github.com/alex-held/devctl-plugins/pkg/devctlog"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/alex-held/devctl-plugins/pkg/devctlpath"
)

type GoLinkerCmd struct {
	Pather devctlpath.Pather
	fs     vfs.VFS
	Logger devctlog.Logger
}

func (cmd *GoLinkerCmd) SetLogger(feeder LoggerFeeder) {
	cmd.Logger = feeder()
}

func (cmd *GoLinkerCmd) SetPather(feeder PatherFeeder) {
	cmd.Pather = feeder()
}

func (cmd *GoLinkerCmd) SetFsFeeder(feeder FileSystemFeeder) {
	cmd.fs = feeder()
}

func (cmd *GoLinkerCmd) RemoveBrokenSymlinks(directory string) (err error) {
	fileInfos, err := cmd.fs.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, fi := range fileInfos {
		filepath := path.Join(directory, fi.Name())
		_ = tryRemoveSymlink(cmd.fs, filepath)
	}
	return nil
}

func tryRemoveSymlink(vfs vfs.VFS, name string) error {
	fi, err := vfs.Lstat(name)
	if err != nil && os.IsNotExist(err) {
		return nil
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s: not a symlink. Inspect and remove it manually to continue", name)
	}

	if err = vfs.Remove(name); err != nil {
		return err
	}

	return nil
}

func (cmd *GoLinkerCmd) PluginName() string {
	return GoLinkCmdName
}

func (cmd *GoLinkerCmd) CmdName() string {
	return "link"
}
