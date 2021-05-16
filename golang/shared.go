package golang

import (
	"os"

	"github.com/alex-held/devctl-plugins/pkg/devctlog"
	devctlpath2 "github.com/alex-held/devctl-plugins/pkg/devctlpath"
	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	"github.com/alex-held/devctl/pkg/plugins"
	"github.com/hashicorp/go-hclog"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type FileSystemNeeder interface {
	SetFsFeeder(feeder FileSystemFeeder)
}
type FileSystemFeeder func() vfs.VFS

type PatherNeeder interface {
	SetPather(feeder PatherFeeder)
}
type PatherFeeder func() devctlpath2.Pather

type LoggerNeeder interface {
	plugins.Plugin
	SetLogger(feeder LoggerFeeder)
}
type LoggerFeeder func() devctlog.Logger

const (
	GoSDKPluginVersion = "v0.0.1"
	GoCmdName          = "sdk/go"
	GoListCmdName      = "sdk/go/list"
	GoDownloadCmdName  = "sdk/go/download"
	GoInstallCmdName   = "sdk/go/install"
	GoLinkCmdName      = "sdk/go/link"
	GoUseCmdName       = "sdk/go/use"
)

func NewLogger(prefix string) (logger devctlog.Logger) {
	opt := devctlog.DefaultLoggerOptions(prefix)
	opt.Name = prefix
	opt.Color = hclog.AutoColor
	opt.Output = os.Stderr
	opt.IncludeLocation = false
	opt.TimeFormat = "15:04:05"
	opt.DisableTime = false
	opt.Level = hclog.Debug
	opt.JSONFormat = false

	return hclog.New(opt)
}

func FormatGoArchiveArtifactName(ri sysutils.RuntimeInfo, version string) string {
	return ri.Format("go%s.[os]-[arch].tar.gz", version)
}
