package golang

import (
	"context"
	"fmt"
	"os"

	"github.com/alex-held/devctl-plugins/pkg/devctlpath"
	"github.com/alex-held/devctl-plugins/pkg/plugins/sdk"
	"github.com/gobuffalo/plugins"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const (
	GoSDKPluginVersion = "v0.0.1"
	GoCmdName          = "sdk/go"
	GoListCmdName      = "sdk/go/list"
	GoDownloadCmdName  = "sdk/go/download"
	GoInstallCmdName   = "sdk/go/install"
	GoLinkCmdName      = "sdk/go/link"
	GoUseCmdName       = "sdk/go/use"
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

func main() {
	cmd := &GoSDKCmd{}
	args := os.Args
	err := cmd.Main(context.Background(), "", args)
	if err != nil {
		fmt.Printf("Error executing go sdk plugin. Args=%v\n%v\n", args, err)
	}
}

func (cmd *GoSDKCmd) Main(ctx context.Context, _ string, args []string) error {

	cmd.Fs = vfs.New(osfs.New())
	cmd.Pather = devctlpath.NewPather()

	plugs := cmd.ScopedPlugins()
	subcommand := FindSubcommandFromArgs(args, plugs)

	version := args[0]

	switch cmd := subcommand.(type) {
	case *GoUseCmd:
		return cmd.Use(ctx, version)
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
