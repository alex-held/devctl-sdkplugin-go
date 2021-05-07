package golang

import (
	"context"
	"os"

	"github.com/pkg/errors"

	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	"github.com/alex-held/devctl/cli/cmds/sdk"
	"github.com/alex-held/devctl/pkg/devctlpath"
	plugins2 "github.com/gobuffalo/plugins"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type GoSDKCmd struct {
	Plugins   plugins2.Plugins
	pluginsFn plugins2.Feeder
	Pather    devctlpath.Pather

	RuntimeInfoGetter *sysutils.DefaultRuntimeInfoGetter
	Fs                vfs.VFS
}

func (cmd *GoSDKCmd) WithPlugins(feeder plugins2.Feeder) {
	cmd.Plugins = feeder()
}

func (cmd *GoInstallCmd) Install(version string) error {
	archiveName := FormatGoArchiveArtifactName(cmd.Runtime.Get(), version)
	archivePath := cmd.Pather.Download("go", version, archiveName)
	installPath := cmd.Pather.SDK("go", version)

	archive, err := cmd.Fs.OpenFile(archivePath, os.O_WRONLY, os.ModePerm)
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
}

func (cmd *GoListerCmd) ListInstalled(_ context.Context) (versions []string, err error) {
	dir := cmd.Pather.SDK("go")
	fileInfos, err := cmd.Fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, fi := range fileInfos {
		if sdkVersion := fi.Name(); sdkVersion != "current" {
			versions = append(versions, sdkVersion)
		}
	}
	return versions, nil
}

func (cmd *GoSDKCmd) Link(version string) error {
	panic("implement me")
}

func (cmd *GoSDKCmd) ExecuteCommand(ctx context.Context, root string, args []string) error {
	return cmd.Main(ctx, root, args)
}

func (cmd *GoSDKCmd) CmdName() string {
	return "go"
}

func (cmd *GoSDKCmd) PluginName() string {
	return "sdk/go"
}

func (cmd *GoSDKCmd) Description() string {
	return "manages the installations of the go sdk"
}

func (cmd *GoSDKCmd) SubCommands() []plugins2.Plugin {
	return plugins2.Plugins{
		&GoListerCmd{},
		&GoDownloadCmd{},
		&GoUseCmd{},
		&GoInstallCmd{},
	}
}

func (cmd *GoSDKCmd) ScopedPlugins() []plugins2.Plugin {
	var plugs []plugins2.Plugin
	if cmd.pluginsFn == nil {
		return plugs
	}

	plugs = append(plugs, cmd.Plugins...)
	for _, p := range cmd.pluginsFn() {
		if _, ok := p.(sdk.Sdker); ok {
			plugs = append(plugs, p)
		}
	}
	return plugs
}
