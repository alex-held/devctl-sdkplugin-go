package golang

import (
	devctlpath2 "github.com/alex-held/devctl-plugins/pkg/devctlpath"
	"github.com/alex-held/devctl-plugins/pkg/sysutils"
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

const (
	GoSDKPluginVersion = "v0.0.1"
	GoCmdName          = "sdk/go"
	GoListCmdName      = "sdk/go/list"
	GoDownloadCmdName  = "sdk/go/download"
	GoInstallCmdName   = "sdk/go/install"
	GoLinkCmdName      = "sdk/go/link"
	GoUseCmdName       = "sdk/go/use"
)

func FormatGoArchiveArtifactName(ri sysutils.RuntimeInfo, version string) string {
	return ri.Format("go%s.[os]-[arch].tar.gz", version)
}
