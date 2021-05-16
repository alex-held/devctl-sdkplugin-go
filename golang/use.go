package golang

import (
	"context"
	"fmt"

	"github.com/alex-held/devctl-plugins/pkg/devctlog"
	"github.com/gobuffalo/plugins"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/pkg/errors"

	"github.com/alex-held/devctl/pkg/devctlpath"
)

type GoUseCmd struct {
	Plugins plugins.Plugins
	Pather  devctlpath.Pather
	Fs      vfs.VFS
	Logger  devctlog.Logger
}

func (cmd *GoUseCmd) SetLogger(feeder LoggerFeeder) {
	cmd.Logger = feeder()
}

func (cmd *GoUseCmd) SetFsFeeder(feeder FileSystemFeeder) {
	cmd.Fs = feeder()
}

func (cmd *GoUseCmd) SetPather(feeder PatherFeeder) {
	cmd.Pather = feeder()
}

func (cmd *GoUseCmd) Use(ctx context.Context, version string) error {
	var downloadCmd *GoDownloadCmd
	var installCmd *GoInstallCmd
	var linkerCmd *GoLinkerCmd

	for _, plugin := range cmd.Plugins {
		if p, ok := plugin.(*GoDownloadCmd); ok {
			downloadCmd = p
		}
		if p, ok := plugin.(*GoInstallCmd); ok {
			installCmd = p
		}
		if p, ok := plugin.(*GoLinkerCmd); ok {
			linkerCmd = p
		}
		cmd.Logger.Trace("GoUserCmd does not support plugin", "name", plugin.PluginName(), "type", fmt.Sprintf("%T", plugin))
	}
	err := downloadCmd.Download(ctx, version)
	if err != nil {
		return errors.Wrapf(err, "failed downloading go sdk %s\n", version)
	}
	err = installCmd.Install(version)
	if err != nil {
		return errors.Wrapf(err, "failed installing go sdk %s\n", version)
	}
	err = linkerCmd.Link(ctx, version)
	if err != nil {
		return errors.Wrapf(err, "failed installing go sdk %s\n", version)
	}
	return nil
}

func (cmd *GoUseCmd) WithPlugins(feeder plugins.Feeder) {
	cmd.Plugins = feeder()
}

func (cmd *GoUseCmd) PluginName() string {
	return GoUseCmdName
}

func (cmd *GoUseCmd) CmdName() string {
	return "use"
}
