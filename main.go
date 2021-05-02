package golang

import (
	"context"
	"fmt"

	"github.com/alex-held/devctl/pkg/devctlpath"
	"github.com/gobuffalo/plugins"
	_ "github.com/gobuffalo/plugins"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/alex-held/devctl-plugins/pkg/plugins/sdk"
)

const (
	GoCmdName         = "sdk/go"
	GoListCmdName     = "sdk/go/list"
	GoDownloadCmdName = "sdk/go/download"
	GoInstallCmdName  = "sdk/go/install"
	GoLinkCmdName     = "sdk/go/link"
	GoUseCmdName      = "sdk/go/use"
)

func Plugins() []plugins.Plugin {
	return []plugins.Plugin{
		&GoSDKCmd{
			Plugins: []plugins.Plugin{
				&GoDownloadCmd{},
				&GoListerCmd{},
				&GoLinkerCmd{},
				&GoInstallCmd{},
				&GoUseCmd{},
			},
		},
	}
}

func (cmd *GoSDKCmd) Main(ctx context.Context, root string, args []string) error {
	cmd.Fs = vfs.New(osfs.New())
	cmd.Pather = devctlpath.NewPather()

	plugs := cmd.ScopedPlugins()
	subcommand := FindSubcommandFromArgs(args, plugs)

	version := args[0]

	switch cmd := subcommand.(type) {
	case *GoUseCmd:
		return cmd.ExecuteCommand(ctx, "", []string{version})
	case *GoDownloadCmd:
		return cmd.Download(ctx, version)
	case sdk.Installer:
		return cmd.Install(version)
	case sdk.Linker:
		return cmd.Link(version)
	case sdk.Lister:
		// TODO: multiplex console output, so that return values are not necessary
		versions, err := cmd.ListInstalled(version)
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", versions)
		return err
	}
	return fmt.Errorf("plugin %s has a unsupported api", cmd.PluginName())
}
