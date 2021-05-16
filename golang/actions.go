package golang

import (
	"context"
	"io"
	"os"
	"path"

	downloader2 "github.com/alex-held/devctl/pkg/plugins/downloader"
	"github.com/pkg/errors"
)

// Download downloads a tarball of wanted version
func (cmd *GoDownloadCmd) Download(ctx context.Context, version string) error {
	artifactName := FormatGoArchiveArtifactName(cmd.Runtime.Get(), version)
	dlDirectory := cmd.Pather.Download("go", version)
	archivePath := path.Join(dlDirectory, artifactName)
	dlUri := cmd.Runtime.Get().Format("%s/dl/%s", cmd.BaseUri, artifactName)

	cmd.Logger.Debug("resolved download paths",
		"artifactName", artifactName,
		"dlDirectory", dlDirectory,
		"archivePath", archivePath,
		"artifactName", artifactName)

	if err := cmd.Fs.MkdirAll(dlDirectory, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed creating go sdk download PatherFeeder; version=%s", version)
	}

	if exists, _ := cmd.Fs.Exists(archivePath); exists {
		cmd.Logger.Info("go sdk tarball already exists. skipping re-download...",
			"sdk version", version,
			"archivePath", archivePath)
		return nil
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
}

func (cmd *GoInstallCmd) Install(version string) error {
	archiveName := FormatGoArchiveArtifactName(cmd.Runtime.Get(), version)
	archivePath := cmd.Pather.Download("go", version, archiveName)
	installPath := cmd.Pather.SDK("go", version)
	cmd.Logger.Debug("resolved install paths", "archiveName", archiveName, "archivePath", archivePath, "installPath", installPath)

	archive, err := cmd.Fs.OpenFile(archivePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "failed to open go sdk archive=%s\n", archivePath)
	}
	err = cmd.Fs.MkdirAll(installPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "failed to Extract go sdk %s; dest=%s; archive=%s\n", version, installPath, archivePath)
	}
	err = UnTarGzip(archive, installPath, GoSDKUnarchiveRenamer(), cmd.Fs)
	if err != nil {
		return errors.Wrapf(err, "failed to Extract go sdk %s; dest=%s; archive=%s\n", version, installPath, archivePath)
	}
	return nil
}

func (cmd *GoListerCmd) ListInstalled(_ string) (versions []string, err error) {
	dir := cmd.Pather.SDK("go")
	cmd.Logger.Debug("resolved go sdk root", "go sdk root path", dir)

	fileInfos, err := cmd.Fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var current string
	for _, fi := range fileInfos {
		if sdkVersion := fi.Name(); sdkVersion != "current" {
			versions = append(versions, sdkVersion)
		}
		current = fi.Name()
	}
	cmd.Logger.Info("local go sdk versions", "versions", versions, "current", current)

	return versions, nil
}

func (cmd *GoLinkerCmd) Link(ctx context.Context, version string) (err error) {
	sdkPath := cmd.Pather.SDK("go", version)
	current := cmd.Pather.SDK("go", "current")

	cmd.Logger.Debug("resolved linking paths",
		"target sdk version", version,
		"target sdk path", sdkPath,
		"current sdk path", current)

	err = cmd.fs.RemoveAll(current)
	if err != nil {
		return errors.Wrapf(err, "failed to remove current symlink")
	}

	if e, err := cmd.fs.DirExists(sdkPath); !e || err != nil {
		return err
	}

	err = cmd.fs.Symlink(sdkPath, current)
	cmd.Logger.Info("linked go sdk to current", "version", version, "current sdk path", current)
	return err
}
