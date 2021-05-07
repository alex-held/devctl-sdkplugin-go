package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/alex-held/devctl-plugins/pkg/plugins/sdk"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	shared2 "github.com/alex-held/devctl-sdkplugin-go/host/shared"
)

func main() {

	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "host",
		Output: os.Stdout,
		Level:  hclog.Trace,
	})

	goDownloaderPluginPath := path.Join(os.ExpandEnv("$DEVCTL_ROOT"), "plugins", "sdk", "go", "downloader")

	// We're a host! Start by launching the plugSrc process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(goDownloaderPluginPath),
		Logger:          logger,
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Request the plugSrc
	raw, err := rpcClient.Dispense("sdk.go.downloader")
	if err != nil {
		log.Fatal(err)
	}

	// We should have a Greeter now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	ctx := context.Background()
	version := "1.16.1"
	downloader := raw.(sdk.Downloader)
	fmt.Println(downloader.Download(ctx, version))
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

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"sdk.go.downloader": &shared2.GoDownloaderPlugin{},
}
