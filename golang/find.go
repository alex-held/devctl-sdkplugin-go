package golang

import (
	"path"
	"strings"

	devctlPlug "github.com/alex-held/devctl/pkg/plugins"
	"github.com/gobuffalo/plugins"
	"github.com/gobuffalo/plugins/plugcmd"
)

func FindSubcommandFromArgs(args []string, plugs []plugins.Plugin) plugins.Plugin {
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		return FindSubcommand(a, plugs)
	}
	return nil
}

func FindSubcommand(name string, plugs []plugins.Plugin) plugins.Plugin {
	for _, p := range plugs {
		c, ok := p.(devctlPlug.Plugin)
		if !ok {
			continue
		}
		if n, ok := c.(plugcmd.Namer); ok {
			if n.CmdName() == name {
				return n
			}
		}
		if name == path.Base(p.PluginName()) {
			return p
		}
	}
	return nil
}
