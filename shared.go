package golang

import (
	"github.com/alex-held/devctl-plugins/pkg/sysutils"
)

func FormatGoArchiveArtifactName(ri sysutils.RuntimeInfo, version string) string {
	return ri.Format("go%s.[os]-[arch].tar.gz", version)
}
