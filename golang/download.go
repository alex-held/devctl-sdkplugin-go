package golang

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/alex-held/devctl-plugins/pkg/devctlog"
	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	devctlpath2 "github.com/alex-held/devctl/pkg/devctlpath"
	"github.com/gobuffalo/plugins"
	"github.com/gobuffalo/plugins/plugcmd"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/pkg/errors"

	plugins2 "github.com/alex-held/devctl/pkg/plugins"
	downloader2 "github.com/alex-held/devctl/pkg/plugins/downloader"
	"github.com/alex-held/devctl/pkg/ui/taskrunner"
)

var _ plugins.Plugin = &GoDownloadCmd{}
var _ plugcmd.Namer = &GoDownloadCmd{}
var _ plugins.Plugin = &GoDownloadCmd{}
var _ plugins2.Executor = &GoDownloadCmd{}

type GoDownloadCmd struct {
	Fs      vfs.VFS
	BaseUri string
	Pather  devctlpath2.Pather
	Runtime *sysutils.DefaultRuntimeInfoGetter
	Logger  devctlog.Logger
}

func (cmd *GoDownloadCmd) SetLogger(feeder LoggerFeeder) {
	cmd.Logger = feeder()
}

func (cmd *GoDownloadCmd) AsTasker(version string) (task taskrunner.Tasker) {
	artifactName := FormatGoArchiveArtifactName(cmd.Runtime.Get(), version)
	dlDirectory := cmd.Pather.Download("go", version)
	archivePath := path.Join(dlDirectory, artifactName)
	dlUri := cmd.Runtime.Get().Format("%s/dl/%s", cmd.BaseUri, artifactName)

	cmd.Logger.Debug("resolved download paths",
		"artifactName", artifactName,
		"dlDirectory", dlDirectory,
		"archivePath", archivePath,
		"artifactName", artifactName)

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

func (cmd *GoDownloadCmd) SetPather(feeder PatherFeeder) {
	cmd.Pather = feeder()
}

func (cmd *GoDownloadCmd) SetFsFeeder(feeder FileSystemFeeder) {
	cmd.Fs = feeder()
}

func (cmd *GoDownloadCmd) ExecuteCommand(ctx context.Context, _ string, args []string) error {
	version := args[1]
	task := cmd.AsTasker(version)
	return task.Task(ctx)
}
