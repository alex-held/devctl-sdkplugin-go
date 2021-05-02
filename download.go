package golang

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	"github.com/gobuffalo/plugins"
	"github.com/gobuffalo/plugins/plugcmd"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/pkg/errors"

	"github.com/alex-held/devctl/pkg/devctlpath"
	plugins2 "github.com/alex-held/devctl/pkg/plugins"
	downloader2 "github.com/alex-held/devctl/pkg/plugins/downloader"
	"github.com/alex-held/devctl/pkg/ui/taskrunner"
)

var _ plugcmd.Namer = &GoDownloadCmd{}
var _ plugins.Plugin = &GoDownloadCmd{}
var _ plugins2.Executor = &GoDownloadCmd{}

type GoDownloadCmd struct {
	Fs      vfs.VFS
	BaseUri string
	Pather  devctlpath.Pather
	Runtime *sysutils.DefaultRuntimeInfoGetter
}

func (cmd *GoDownloadCmd) AsTasker(version string) (task taskrunner.Tasker) {
	artifactName := FormatGoArchiveArtifactName(cmd.Runtime.Get(), version)
	dlDirectory := cmd.Pather.Download("go", version)
	archivePath := path.Join(dlDirectory, artifactName)
	dlUri := cmd.Runtime.Get().Format("%s/dl/%s", cmd.BaseUri, artifactName)

	task = &taskrunner.ConditionalTask{
		Description: fmt.Sprintf("downloading the go sdk %s archive to the local storage\n", version),
		Action: func(ctx context.Context) error {
			if err := cmd.Fs.MkdirAll(dlDirectory, os.ModePerm); err != nil {
				return errors.Wrapf(err, "failed creating go sdk download Pather; version=%s", version)
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
		},
		ShouldExecute: func() bool {
			exists, _ := cmd.Fs.Exists(archivePath)
			return !exists
		},
	}
	return task
}

func (cmd *GoDownloadCmd) CmdName() string {
	return "download"
}

func (cmd *GoDownloadCmd) PluginName() string {
	return GoDownloadCmdName
}

func (cmd *GoDownloadCmd) ExecuteCommand(ctx context.Context, _ string, args []string) error {
	cmd.Init()
	version := args[1]
	task := cmd.AsTasker(version)
	return task.Task(ctx)
}

func (cmd *GoDownloadCmd) Init() {
	cmd.Fs = vfs.New(osfs.New())
	// TODO: just checking if I can skip init, using better types
	//	cmd.Runtime = system.OSRuntimeInfoGetter{}
	cmd.Pather = devctlpath.DefaultPather()
	if cmd.BaseUri == "" {
		cmd.BaseUri = "https://golang.org"
	}
}
