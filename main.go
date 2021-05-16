package main

import (
	"context"
	"fmt"
	"os"

	devctlpath2 "github.com/alex-held/devctl-plugins/pkg/devctlpath"
	"github.com/gobuffalo/plugins"
	"github.com/gobuffalo/plugins/plugprint"
	"github.com/hashicorp/go-hclog"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"

	. "github.com/alex-held/devctl-sdkplugin-go/golang"
)

func Plugins() []plugins.Plugin {
	var feeder plugins.Feeder = func() []plugins.Plugin {
		return []plugins.Plugin{
			&GoDownloadCmd{
				BaseUri: "https://golang.org",
			},
			&GoListerCmd{},
			&GoLinkerCmd{},
			&GoInstallCmd{},
			&GoUseCmd{},
		}
	}

	return []plugins.Plugin{
		&GoSDKCmd{
			Config: &PluginConfig{
				LogLevel: hclog.Warn,
				Pather:   devctlpath2.NewPather(),
				Vfs:      vfs.New(osfs.New()),
				Plugins:  feeder(),
			},
			Logger: NewLogger("sdk/go"),
		},
	}
}

func main() {
	cmd := Plugins()[0].(*GoSDKCmd)
	args := os.Args

	var help bool
	var version bool
	var level int32
	flagSet := pflag.NewFlagSet("sdk/go", pflag.ContinueOnError)
	flagSet.BoolVarP(&help, "help", "h", false, "(--help | -h)")
	flagSet.BoolVarP(&version, "version", "v", false, "(--version | -v)")
	flagSet.Int32VarP(&level, "level", "l", 0, `(--level | -l) (
		0 <default> | allow default logging handling
		1 <trace>   | Intended to be used for tracing
		2 <debug>   | Debug information for programmer lowlevel analysis.
		3 <info>    | Info information about steady state operations.
		4 <warn>    | Warn information about rare but handled events.
		5 <error>   |  Error information about unrecoverable events.
		6 <off>     | Off disables all logging output.
	)`)

	if err := flagSet.Parse(args); err != nil {
		cmd.Logger.Error("failed to parse cli flags. exiting...", "args", args, "err", err)
		os.Exit(1)
	}

	printHelpFlagValue, err := flagSet.GetBool("help")
	if err != nil {
		cmd.Logger.Error("failed to parse cli flag. exiting...", "failed flag", "help", "args", args, "err", err)
		os.Exit(1)
	}
	if printHelpFlagValue {
		if err = plugprint.Print(os.Stdout, cmd); err != nil {
			cmd.Logger.Error("failed to print cli help. exiting...", "args", args, "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	printVersionFlagValue, err := flagSet.GetBool("version")
	if err != nil {
		cmd.Logger.Error("failed to parse cli flag. exiting...", "failed flag", "version", "args", args, "err", err)
		os.Exit(1)
	}
	if printVersionFlagValue {
		_, _ = fmt.Println(Version)
		os.Exit(0)
	}

	logLevelFlagValue, err := flagSet.GetInt32("level")
	if err != nil {
		cmd.Logger.Error("failed to parse cli flag. exiting...", "failed flag", "level", "args", args, "err", err)
		os.Exit(1)
	} else {
		cmd.Config.LogLevel = hclog.Level(logLevelFlagValue)
		cmd.Logger.SetLevel(hclog.Level(logLevelFlagValue))
	}

	cmd.Logger.Trace("parsed cli flags", "args", args, "flagSet", *flagSet)
	cmd.Logger.Trace("starting GoSDKCmd.Main plugin execution", "flagSet", *flagSet)

	if err := cmd.Main(context.Background(), "-", args); err != nil {
		cmd.Logger.Error("failed during plugin execution", "flagSet", *flagSet, "err", err)
		os.Exit(1)
	}
}
