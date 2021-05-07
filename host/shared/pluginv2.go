package shared

import (
	"context"
	"net/rpc"

	"github.com/alex-held/devctl-plugins/pkg/plugins/sdk"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

type GOSDKPluginV2 struct {
	logger hclog.Logger
}

type DownloaderRPC struct{ client *rpc.Client }

func (g *DownloaderRPC) PluginName() string {
	return "sdk.go.downloader"
}

func (g *DownloaderRPC) CmdName() string { return "download" }

func (g *DownloaderRPC) Download(ctx context.Context, version string) error {
	var resp error
	arg := &DownloadArg{ctx, version}
	err := g.client.Call("sdk.go.downloader", arg, &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}
	return resp
}

// DownloaderRPCServer is the RPC server that DownloaderRPC talks to, conforming to
// the requirements of net/rpc
type DownloaderRPCServer struct {
	// This is the real implementation
	Impl sdk.Downloader
}

type DownloadArg struct {
	Context context.Context
	Version string
}

func (s *DownloaderRPCServer) Download(args interface{}, resp *string) error {
	dlArg := args.(*DownloadArg)
	err := s.Impl.Download(dlArg.Context, dlArg.Version)
	if err != nil {
		*resp = err.Error()
	}
	return nil
}

// GoDownloaderPlugin  is the implementation of plugSrc.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugSrc
// type. We construct a DownloaderRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return DownloaderRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugSrc connection and is a more advanced use case.
type GoDownloaderPlugin struct {
	// Impl Injection
	Impl sdk.Downloader
}

func (p *GoDownloaderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &DownloaderRPCServer{Impl: p.Impl}, nil
}

func (GoDownloaderPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &DownloaderRPC{client: c}, nil
}
