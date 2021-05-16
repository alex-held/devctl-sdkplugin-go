package golang

import (
	"github.com/alex-held/devctl-plugins/pkg/devctlog"
	"github.com/gobuffalo/plugins"
	"github.com/gobuffalo/plugins/plugcmd"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/afero"

	"github.com/alex-held/devctl-plugins/pkg/devctlpath"
)

var _ plugcmd.Namer = &GoListerCmd{}
var _ plugins.Plugin = &GoListerCmd{}

type GoListerCmd struct {
	fs     afero.Fs
	Fs     vfs.VFS
	Pather devctlpath.Pather
	Logger devctlog.Logger
}

func (cmd *GoListerCmd) SetLogger(feeder LoggerFeeder) {
	cmd.Logger = feeder()
}

func (cmd *GoListerCmd) SetPather(feeder PatherFeeder) {
	cmd.Pather = feeder()
}

func (cmd *GoListerCmd) SetFsFeeder(feeder FileSystemFeeder) {
	cmd.Fs = feeder()
}

func (cmd *GoListerCmd) CmdName() string {
	return "list"
}

func (cmd *GoListerCmd) PluginName() string {
	return "sdk/go/list"
}
