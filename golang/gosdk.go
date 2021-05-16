package golang

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alex-held/devctl-plugins/pkg/devctlog"
	"github.com/alex-held/devctl-plugins/pkg/devctlpath"
	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	"github.com/gobuffalo/plugins"
	"github.com/gobuffalo/plugins/plugcmd"
	"github.com/gobuffalo/plugins/plugprint"
	"github.com/hashicorp/go-hclog"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/rogpeppe/go-internal/semver"
)

var _ plugcmd.SubCommander = &GoSDKCmd{}
var _ plugcmd.Commander = &GoSDKCmd{}
var _ plugins.Plugin = &GoSDKCmd{}
var _ plugins.Scoper = &GoSDKCmd{}
var _ plugprint.Describer = &GoSDKCmd{}
var _ plugprint.Describer = &GoSDKCmd{}

type GoSDKCmd struct {
	Config            *PluginConfig
	Logger            devctlog.Logger
	Pather            devctlpath.Pather
	RuntimeInfoGetter *sysutils.DefaultRuntimeInfoGetter
	Fs                vfs.VFS
	scopedPlugins     plugins.Plugins
}

func (cmd *GoSDKCmd) WithPlugins(feeder plugins.Feeder) {
	cmd.Config.Plugins = feeder()
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

func (cmd *GoSDKCmd) SubCommands() (subcommands []plugins.Plugin) {
	return cmd.ScopedPlugins()
}

func (cmd *GoSDKCmd) ScopedPlugins() []plugins.Plugin {
	var plugs []plugins.Plugin
	if cmd.scopedPlugins != nil {
		return cmd.scopedPlugins
	}
	if cmd.Config.PluginFeeder() == nil {
		return plugs
	}

	for _, p := range cmd.Config.Plugins {
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

type PluginConfig struct {
	LogLevel hclog.Level
	Vfs      vfs.VFS
	Pather   devctlpath.Pather
	Plugins  plugins.Plugins
}

func (p *PluginConfig) PluginFeeder() plugins.Feeder {
	return func() []plugins.Plugin {
		return p.Plugins
	}
}

func (p *PluginConfig) LoggerFeeder(name string) LoggerFeeder {
	return func() devctlog.Logger {
		l := NewLogger(name)
		l.SetLevel(p.LogLevel)
		return l
	}
}

func (p *PluginConfig) FsFeeder() FileSystemFeeder {
	return func() vfs.VFS {
		return p.Vfs
	}
}

func (p *PluginConfig) PatherFeeder() PatherFeeder {
	return func() devctlpath.Pather {
		return p.Pather
	}
}

type FeederGetter interface {
	LoggerFeeder(name string) LoggerFeeder
	FsFeeder() FileSystemFeeder
	PatherFeeder() PatherFeeder
	PluginFeeder() plugins.Feeder
}

func (cmd *GoSDKCmd) initializePlugins() {
	for _, plugin := range cmd.Config.Plugins {
		cmd.Logger.Trace("providing dependency feeders for dependency needers",
			"plugin name", plugin.PluginName(),
			"type", fmt.Sprintf("%T", plugin))

		if p, ok := plugin.(LoggerNeeder); ok {
			p.SetLogger(cmd.Config.LoggerFeeder(plugin.PluginName()))
		}
		if p, ok := plugin.(FileSystemNeeder); ok {
			p.SetFsFeeder(cmd.Config.FsFeeder())
		}
		if p, ok := plugin.(PatherNeeder); ok {
			p.SetPather(cmd.Config.PatherFeeder())
		}

		cmd.Config.Plugins = append(cmd.Config.Plugins, plugin)
	}

	var plugs plugins.Plugins
	for _, plugin := range cmd.Config.Plugins {
		if p, ok := plugin.(plugins.Needer); ok {
			cmd.Logger.Trace("providing plugin feeder for plugin needer",
				"plugin name", plugin.PluginName(), "type", fmt.Sprintf("%T", p),
				"is needer", ok)

			p.WithPlugins(func() []plugins.Plugin {
				return plugs
			})
		}
		plugs = append(plugs, plugin)
	}
	cmd.Config.Plugins = plugs
}

func versionFromArgs(args []string) (version string, containsVersion bool) {
	for _, arg := range args {
		semVer := "v" + strings.TrimPrefix(arg, "v")
		if semver.IsValid(semVer) {
			return arg, true
		}
	}

	return "", false
}

func (cmd *GoSDKCmd) Main(ctx context.Context, _ string, args []string) error {
	cmd.initializePlugins()

	var version string
	logger := cmd.Logger
	subcommand := FindSubcommandFromArgs(args[1:], cmd.SubCommands())

	if subcommand != nil {
		var containsVersion bool
		if version, containsVersion = versionFromArgs(args); !containsVersion {
			logger.Error("no go sdk version provided in args", "args", args)
			os.Exit(1)
		}
	}

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
	case nil:
		return nil
	default:
		logger.Error("plugin is currently not supported", "name", cmd.PluginName(), "type", fmt.Sprintf("%T", cmd))
		return fmt.Errorf("plugin '%s'  of type %T is currently not supported\n", cmd.PluginName(), cmd)
	}
}
