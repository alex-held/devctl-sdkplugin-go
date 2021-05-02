package golang

import (
	"context"
	"fmt"
	"time"

	"github.com/gobuffalo/plugins"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/pkg/errors"

	"github.com/alex-held/devctl/pkg/ui/taskrunner"

	"github.com/alex-held/devctl/pkg/devctlpath"
)

type GoUseCmd struct {
	plugins plugins.Plugins
	path    devctlpath.Pather
	fs      vfs.VFS
}

func (cmd *GoUseCmd) WithPlugins(feeder plugins.Feeder) { cmd.plugins = feeder() }

func (cmd *GoUseCmd) CreateTaskRunner(version string) (runner taskrunner.Runner) {
	var tasks taskrunner.Tasks
	for _, plug := range cmd.plugins {
		switch t := plug.(type) {
		case *GoDownloadCmd:
			tasks = append(tasks, t.AsTasker(version))
		case *GoInstallCmd:
			tasks = append(tasks, t.AsTasker(version))
		case *GoLinkerCmd:
			tasks = append(tasks, t.AsTasker(version))
		default: // no-op
		}
	}

	runner = taskrunner.NewTaskRunner(
		taskrunner.WithTasks(tasks...),
		taskrunner.WithTitle(fmt.Sprintf("use go sdk %s", version)),
		taskrunner.WithTimeout(50*time.Millisecond),
	)
	return runner
}

func (cmd *GoUseCmd) PluginName() string {
	return GoUseCmdName
}

func (cmd *GoUseCmd) CmdName() string {
	return "use"
}

func (cmd *GoUseCmd) ExecuteCommand(ctx context.Context, _ string, args []string) (err error) {
	version := args[1]
	runner := cmd.CreateTaskRunner(version)
	err = runner.Run(ctx)
	if err != nil {
		return errors.Wrapf(err, "GoUse-TaskRunner execution failed.. ERROR=%v, GoSDKVersion=%s, Tasks=%s", err, version, runner.Describe())
	}

	return nil

	/*
		sdkPath, _ := cmd.fs.EvalSymlinks(cmd.Pather.SDK("go", version))
		current, _ := cmd.fs.EvalSymlinks(cmd.Pather.SDK("go", "current"))

		// 1. Clean up existing @current
		fi, err := cmd.fs.Stat(current)
		if err == nil {
			_ = cmd.fs.Remove(current)
			fi.PluginName()
		}

		// 2. Make sure directories exists
		_ = cmd.fs.MkdirAll(cmd.Pather.SDK("go"), os.ModePerm)

		// 4. Is the go sdk version installed?
		if exists, _ := cmd.fs.DirExists(sdkPath); !exists {

			// 4 -> Start different plugin do install
			// todo: search plugin and start it
			// todo: ask for user input, if the sdk should be installed

			if err = cmd.fs.MkdirAll(sdkPath, os.ModePerm); err != nil {
				runner := cmd.CreateTaskRunner(version)
				err = runner.Run(ctx)
				return errors.Wrapf(err, "GoUse-TaskRunner execution failed.. ERROR=%v, GoSDKVersion=%s, Tasks=%s", err, version, runner.Describe())
			}
		}

		// 5. Use the go sdk version to @current
		// ln -s -v -F  /root/sdks/go/1.16.3  /root/sdks/go/current
		// err = cmd.fs.Symlink(sdkPath, current)
		return err*/
}
