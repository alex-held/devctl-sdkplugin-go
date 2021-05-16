package main

import (
	"context"
	"fmt"
	"os"

	devctlpath2 "github.com/alex-held/devctl-plugins/pkg/devctlpath"
	"github.com/gobuffalo/plugins"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	. "github.com/alex-held/devctl-sdkplugin-go/golang"
)

func Plugins() []plugins.Plugin {
	var feeder plugins.Feeder = func() []plugins.Plugin {
		return []plugins.Plugin{
			&GoDownloadCmd{
				BaseUri: "https://golang.org",
			},
			&GoListerCmd{},
			&GoLinkerCmd{},
			&GoInstallCmd{},
			&GoUseCmd{},
		}
	}

	return []plugins.Plugin{
		&GoSDKCmd{
			Plugins: feeder(),
			Feeder:  feeder,
			Pather:  devctlpath2.NewPather(),
			Fs:      vfs.New(osfs.New()),
		},
	}
}

func main() {
	cmd := Plugins()[0].(*GoSDKCmd)
	args := os.Args
	err := cmd.Main(context.Background(), "", args)
	
	if err != nil {
		fmt.Printf("Error executing go sdk plugin. Args=%v\n%v\n", args, err)
	}
}
