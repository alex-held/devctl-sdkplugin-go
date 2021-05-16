package golang

import (
	"github.com/alex-held/devctl-plugins/pkg/devctlog"
	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	devctlpath2 "github.com/alex-held/devctl/pkg/devctlpath"
	"github.com/gobuffalo/plugins"
	"github.com/gobuffalo/plugins/plugcmd"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var _ plugins.Plugin = &GoDownloadCmd{}
var _ plugcmd.Namer = &GoDownloadCmd{}
var _ plugins.Plugin = &GoDownloadCmd{}

type GoDownloadCmd struct {
	Fs      vfs.VFS
	BaseUri string
	Pather  devctlpath2.Pather
	Runtime *sysutils.DefaultRuntimeInfoGetter
	Logger  devctlog.Logger
}

func (cmd *GoDownloadCmd) SetLogger(feeder LoggerFeeder) {
	cmd.Logger = feeder()
}

func (cmd *GoDownloadCmd) CmdName() string {
	return "download"
}

func (cmd *GoDownloadCmd) PluginName() string {
	return GoDownloadCmdName
}

func (cmd *GoDownloadCmd) SetPather(feeder PatherFeeder) {
	cmd.Pather = feeder()
}

func (cmd *GoDownloadCmd) SetFsFeeder(feeder FileSystemFeeder) {
	cmd.Fs = feeder()
}
