package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alex-held/devctl/pkg/devctlpath"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	golang "github.com/alex-held/devctl-sdkplugin-go"
	shared2 "github.com/alex-held/devctl-sdkplugin-go/host/shared"
)

// GoDownloader is a real implementation of sdk.Downloader
type GoDownloader struct {
	logger hclog.Logger
}

func (g *GoDownloader) PluginName() string {
	return "sdk.go.Download"
}

func (g *GoDownloader) CmdName() string {
	return "download"
}

func (g *GoDownloader) Download(ctx context.Context, version string) (err error) {
	g.logger.Info(fmt.Sprintf("GoDownloader handles sdk  sdk %s download request\n", version))

	cmd := golang.GoDownloadCmd{
		Fs:      vfs.New(osfs.New()),
		BaseUri: "https://golang.dl/cl",
		Pather:  devctlpath.NewPather(),
		Runtime: nil,
	}

	g.logger.Info(fmt.Sprintf("GoDownloader created *GoDownloadCmd to download go sdk\nGoDownloadCmd=%v\n", cmd))

	err = cmd.Download(ctx, version)
	if err != nil {
		g.logger.Error(fmt.Sprintf("GoDownloader failed to download go sdk %s\n", version))
	}
	g.logger.Info(fmt.Sprintf("GoDownloader downloaded go sdk %s successfully!\n", version))
	return err
}

// handshakeConfigs are used to just do a basic handshake between
// a plugSrc and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugSrc
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "sdk.go.downloader",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	downloader := &GoDownloader{
		logger: logger,
	}

	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"sdk.go.downloader": &shared2.GoDownloaderPlugin{Impl: downloader},
	}

	logger.Debug("message from GoDownloaderPlugin")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
