package golang

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	downloader2 "github.com/alex-held/devctl/pkg/plugins/downloader"
	"github.com/pkg/errors"

	"github.com/alex-held/devctl-plugins/pkg/devctlpath"
	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	plugins2 "github.com/gobuffalo/plugins"
	"github.com/gobuffalo/plugins/plugcmd"
	"github.com/gobuffalo/plugins/plugprint"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var _ plugcmd.SubCommander = &GoSDKCmd{}
var _ plugcmd.Commander = &GoSDKCmd{}
var _ plugins2.Plugin = &GoSDKCmd{}
var _ plugins2.Scoper = &GoSDKCmd{}
var _ plugprint.Describer = &GoSDKCmd{}

type GoSDKCmd struct {
	Plugins           plugins2.Plugins
	Feeder            plugins2.Feeder
	Pather            devctlpath.Pather
	RuntimeInfoGetter *sysutils.DefaultRuntimeInfoGetter
	Fs                vfs.VFS
	scopedPlugins     plugins2.Plugins
}

// Download downloads a tarball of wanted version
func (cmd *GoDownloadCmd) Download(ctx context.Context, version string) error {
	artifactName := FormatGoArchiveArtifactName(cmd.Runtime.Get(), version)
	dlDirectory := cmd.Pather.Download("go", version)
	archivePath := path.Join(dlDirectory, artifactName)
	dlUri := cmd.Runtime.Get().Format("%s/dl/%s", cmd.BaseUri, artifactName)

	if err := cmd.Fs.MkdirAll(dlDirectory, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed creating go sdk download Pather; version=%s", version)
	}

	if exists, _ := cmd.Fs.Exists(archivePath); exists {
		fmt.Printf("go sdk tarball already exists for version '%s' at path '%s'.. skipping redownload\n", version, archivePath)
		return nil
	}

	artifactFile, err := cmd.Fs.Create(archivePath)
	if err != nil {
		return errors.Wrapf(err, "failed creating / opening file handle for the download")
	}
	dl := downloader2.NewDownloader(dlUri, "downloading the go sdk", artifactFile, io.Discard)
	err = dl.Download(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed downloading go sdk %v from the remote server %s", version, cmd.BaseUri)
	}
	return nil
}

func (cmd *GoSDKCmd) WithPlugins(feeder plugins2.Feeder) {
	cmd.Feeder = feeder
	cmd.Plugins = feeder()
}

func (cmd *GoInstallCmd) Install(version string) error {
	archiveName := FormatGoArchiveArtifactName(cmd.Runtime.Get(), version)
	archivePath := cmd.Pather.Download("go", version, archiveName)
	installPath := cmd.Pather.SDK("go", version)

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
}

func (cmd *GoListerCmd) ListInstalled(_ string) (versions []string, err error) {
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
	panic("")
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

func (cmd *GoSDKCmd) SubCommands() (subcommands []plugins2.Plugin) {
	return cmd.ScopedPlugins()

	for _, plugin := range cmd.Plugins[0].(*GoSDKCmd).Plugins {
		subcommands = append(subcommands, plugin)
	}
	return subcommands
}

func (cmd *GoSDKCmd) ScopedPlugins() []plugins2.Plugin {
	if cmd.scopedPlugins != nil {
		return cmd.scopedPlugins
	}

	var plugs []plugins2.Plugin
	if cmd.Feeder == nil {
		return plugs
	}

	// plugs = append(plugs, cmd.Plugins...)
	for _, p := range cmd.Feeder() {
		fmt.Printf("Plugin '%s' type=%T;", p.PluginName(), p)
		switch t := p.(type) {
		case *GoDownloadCmd, *GoLinkerCmd, *GoListerCmd, *GoUseCmd, *GoInstallCmd:
			plugs = append(plugs, t)
		}
	}
	return plugs
}

func (cmd *GoSDKCmd) initializePlugins() {
	var plugs []plugins2.Plugin

	for _, plugin := range cmd.Plugins {
		fmt.Printf("plugin '%s' has type %T\n", plugin.PluginName(), plugin)

		if p, ok := plugin.(FileSystemNeeder); ok {
			p.SetFsFeeder(func() vfs.VFS {
				return cmd.Fs
			})
		}

		if p, ok := plugin.(PatherNeeder); ok {
			p.SetPather(func() devctlpath.Pather {
				return cmd.Pather
			})
		}

		plugs = append(plugs, plugin)
	}

	cmd.Plugins = plugins2.Plugins{}
	for _, plugin := range plugs {
		if p, ok := plugin.(plugins2.Needer); ok {
			p.WithPlugins(func() []plugins2.Plugin {
				return plugs
			})
		}
		cmd.Plugins = append(cmd.Plugins, plugin)
	}

	cmd.Feeder = func() []plugins2.Plugin {
		return cmd.Plugins
	}
}

func (cmd *GoSDKCmd) Main(ctx context.Context, _ string, args []string) error {

	cmd.initializePlugins()

	subcommand := FindSubcommandFromArgs(args[1:], cmd.SubCommands())

	version := args[2]
	switch cmd := subcommand.(type) {
	case *GoUseCmd:
		return cmd.Use(ctx, version)
	case *GoDownloadCmd:
		return cmd.Download(ctx, version)
	case *GoInstallCmd:
		return cmd.Install(version)
	case *GoLinkerCmd:
		return cmd.Link(ctx, version)
	case *GoListerCmd:
		// TODO: multiplex console output, so that return values are not necessary
		versions, err := cmd.ListInstalled(version)
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", versions)
		return err
	default:
		return fmt.Errorf("plugin '%s'  of type %T is currently not supported\n", cmd.PluginName(), cmd)
	}
	return fmt.Errorf("plugin %s has a unsupported api\n", cmd.PluginName())
}
