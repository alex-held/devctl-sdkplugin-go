package golang

import (
	"context"
	"fmt"
	"os"

	"github.com/alex-held/devctl-plugins/pkg/devctlog"
	"github.com/alex-held/devctl-plugins/pkg/devctlpath"
	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	. "github.com/gobuffalo/plugins"
	"github.com/gobuffalo/plugins/plugcmd"
	"github.com/gobuffalo/plugins/plugprint"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var _ plugcmd.SubCommander = &GoSDKCmd{}
var _ plugcmd.Commander = &GoSDKCmd{}
var _ Plugin = &GoSDKCmd{}
var _ Scoper = &GoSDKCmd{}
var _ plugprint.Describer = &GoSDKCmd{}

type GoSDKCmd struct {
	Logger            devctlog.Logger
	Plugins           Plugins
	Feeder            Feeder
	Pather            devctlpath.Pather
	RuntimeInfoGetter *sysutils.DefaultRuntimeInfoGetter
	Fs                vfs.VFS
	scopedPlugins     Plugins
}

func (cmd *GoSDKCmd) WithPlugins(feeder Feeder) {
	cmd.Feeder = feeder
	cmd.Plugins = feeder()
}

func (cmd *GoSDKCmd) Link(version string) error {
	panic("")
}

func (cmd *GoSDKCmd) ExecuteCommand(ctx context.Context, root string, args []string) error {
	return cmd.Main(ctx, root, args)
}

func (cmd *GoSDKCmd) CmdName() string {
	return "go"
}

func (cmd *GoSDKCmd) PluginName() string {
	return "sdk/go"
}

func (cmd *GoSDKCmd) Description() string {
	return "manages the installations of the go sdk"
}

func (cmd *GoSDKCmd) SubCommands() (subcommands []Plugin) {
	return cmd.ScopedPlugins()
}

func (cmd *GoSDKCmd) ScopedPlugins() []Plugin {
	var plugs []Plugin

	if cmd.scopedPlugins != nil {
		return cmd.scopedPlugins
	}
	if cmd.Feeder == nil {
		return plugs
	}

	for _, p := range cmd.Feeder() {
		fmt.Printf("Plugin '%s' type=%T;", p.PluginName(), p)
		switch t := p.(type) {
		case *GoDownloadCmd, *GoLinkerCmd, *GoListerCmd, *GoUseCmd, *GoInstallCmd:
			cmd.Logger.Trace("adding scoped plugin",
				"name", p.PluginName(),
				"type", fmt.Sprintf("%T", p),
				"parent plugin", cmd.PluginName(),
				"parent type", fmt.Sprintf("%T", cmd),
			)
			plugs = append(plugs, t)
		}
	}
	return plugs
}

func (cmd *GoSDKCmd) initializePlugins() {
	var plugs []Plugin
	for _, plugin := range cmd.Plugins {
		cmd.Logger.Trace("providing dependency feeders for dependency needers",
			"plugin name", plugin.PluginName(),
			"type", fmt.Sprintf("%T", plugin))

		if p, ok := plugin.(LoggerNeeder); ok {
			p.SetLogger(func() devctlog.Logger {
				return NewLogger(plugin.PluginName())
			})
		}
		if p, ok := plugin.(FileSystemNeeder); ok {
			p.SetFsFeeder(func() vfs.VFS {
				return cmd.Fs
			})
		}
		if p, ok := plugin.(PatherNeeder); ok {
			p.SetPather(func() devctlpath.Pather {
				return cmd.Pather
			})
		}
		plugs = append(plugs, plugin)
	}

	cmd.Plugins = Plugins{}
	for _, plugin := range plugs {
		if p, ok := plugin.(Needer); ok {
			cmd.Logger.Trace("providing plugin feeder for plugin needer",
				"plugin name", plugin.PluginName(), "type", fmt.Sprintf("%T", p),
				"is needer", ok)

			p.WithPlugins(func() []Plugin {
				return plugs
			})
		}
		cmd.Plugins = append(cmd.Plugins, plugin)
	}
	cmd.Feeder = func() []Plugin {
		return cmd.Plugins
	}
}

func (cmd *GoSDKCmd) Main(ctx context.Context, _ string, args []string) error {
	cmd.initializePlugins()
	logger := cmd.Logger
	subcommand := FindSubcommandFromArgs(args[1:], cmd.SubCommands())
	version := args[2]

	switch cmd := subcommand.(type) {
	case *GoUseCmd:
		return cmd.Use(ctx, version)
	case *GoDownloadCmd:
		return cmd.Download(ctx, version)
	case *GoInstallCmd:
		return cmd.Install(version)
	case *GoLinkerCmd:
		return cmd.Link(ctx, version)
	case *GoListerCmd:
		// TODO: multiplex console output, so that return values are not necessary
		versions, err := cmd.ListInstalled(version)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(os.Stdout, "%v\n", versions)
		return err
	default:
		logger.Error("plugin is currently not supported", "name", cmd.PluginName(), "type", fmt.Sprintf("%T", cmd))
		return fmt.Errorf("plugin '%s'  of type %T is currently not supported\n", cmd.PluginName(), cmd)
	}
}
