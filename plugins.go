package golang

import (
	"github.com/gobuffalo/plugins"
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
