package golang

import (
	"context"
	_ "embed"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alex-held/devctl-plugins/pkg/sysutils"
	"github.com/gobuffalo/plugins"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alex-held/devctl-plugins/pkg/devctlpath"
	pkgPlugins "github.com/alex-held/devctl/pkg/plugins"
	. "github.com/alex-held/devctl/pkg/testutils"
)

func TestGoSDKPluginSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "go-plugSrc USE")
}

const (
	rootPath = "/root"
	version  = "1.51"
)

type NamedNoOpPlugin struct {
	Name string
	pkgPlugins.NoOpPlugin
}

func init() {
	ArchiveBytes, _ = ioutil.ReadFile("testdata/go1.16.3.darwin-amd64.tar.gz")
}

var ArchiveBytes []byte

func (p *NamedNoOpPlugin) PluginName() string { return p.Name }

var _ = Describe("go-plugSrc USE", func() {
	var (
		vs              vfs.VFS
		versionSdkDir   string
		expectedCurrent string
		pp              devctlpath.Pather
		sut             *GoUseCmd
		dlCmd           *GoDownloadCmd
		linkerCmd       *GoLinkerCmd
		installerCmd    *GoInstallCmd
		srvr            *httptest.Server
		mux             *http.ServeMux
		runtime         *sysutils.DefaultRuntimeInfoGetter
	)

	AfterSuite(func() {
		srvr.Close()
	})

	BeforeEach(func() {
		runtime = &sysutils.DefaultRuntimeInfoGetter{
			GOOS:   "darwin",
			GOARCH: "amd64",
		}
		mux = http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(ArchiveBytes)))
			_, e := w.Write(ArchiveBytes)
			if e != nil {
				panic(e)
			}
		})

		srvr = httptest.NewServer(mux)
		vs = vfs.New(memoryfs.New())
		pp = devctlpath.NewPather(devctlpath.WithConfigRootFn(func() string {
			return rootPath
		}))
		linkerCmd = &GoLinkerCmd{
			Pather: pp,
			fs:     vs,
		}
		installerCmd = &GoInstallCmd{
			Runtime: runtime,
			Pather:  pp,
			Fs:      vs,
		}
		dlCmd = &GoDownloadCmd{
			Fs:      vs,
			BaseUri: srvr.URL,
			Pather:  pp,
			Runtime: runtime,
		}
		sut = &GoUseCmd{
			Plugins: []plugins.Plugin{
				dlCmd,
				installerCmd,
				linkerCmd,
			},
			Pather: pp,
			Fs:     vs,
		}
		versionSdkDir = pp.SDK("go", version)
		expectedCurrent = pp.SDK("go", "current")
	})

	Context("USE <version>", func() {
		When("using the TaskRunner", func() {
			BeforeEach(func() {
				_ = vs.MkdirAll(versionSdkDir, os.ModePerm)
				_ = vs.MkdirAll(pp.SDK("go", version, "src"), os.ModePerm)
				_ = vs.MkdirAll(pp.SDK("go", version, "doc"), os.ModePerm)
				_ = vs.MkdirAll(pp.SDK("go", version, "bin"), os.ModePerm)

				feeder := plugins.Feeder(func() []plugins.Plugin {
					return []plugins.Plugin{
						&NamedNoOpPlugin{
							Name: GoDownloadCmdName,
						},
					}
				})
				sut.WithPlugins(feeder)
			})

			It("resolves the correct Plugins", func() {
				runner := sut.CreateTaskRunner("1.16.3")
				desc := runner.Describe()
				Expect(desc).Should(ContainSubstring("1.16.3"))
			})

		})

		When("no @current version has been installed", func() {
			BeforeEach(func() {
				_ = vs.MkdirAll(versionSdkDir, os.ModePerm)
				_ = vs.MkdirAll(pp.SDK("go", version, "src"), os.ModePerm)
				_ = vs.MkdirAll(pp.SDK("go", version, "doc"), os.ModePerm)
				_ = vs.MkdirAll(pp.SDK("go", version, "bin"), os.ModePerm)
			})

			It("The new SDK Version is symlinked to @current Version ", func() {
				Expect(sut.ExecuteCommand(context.Background(), "devctl", []string{"use", version})).To(Succeed())
				linkDest, _ := vs.Readlink(expectedCurrent)
				Expect(linkDest).Should(Equal(versionSdkDir))
			})
		})

		When("@current has already a sdk configured", func() {

			BeforeEach(func() {
				_ = vs.MkdirAll(pp.SDK("go"), os.ModePerm)
				_ = vs.MkdirAll(pp.Download("go"), os.ModePerm)
				_ = vs.MkdirAll(pp.SDK("go", "19.5"), os.ModePerm)
				_ = vs.Symlink(pp.SDK("go", "19.5"), expectedCurrent)
			})

			It("replaces @current symlink with which links to <version>", func() {
				Expect(sut.ExecuteCommand(context.Background(), "devctl", []string{"use", "1.16.3"})).To(Succeed())
				Expect(expectedCurrent).Should(BeASymlink(vs))
				actual, err := vs.Readlink(expectedCurrent)
				Expect(err).Should(Succeed())
				Expect(actual).ShouldNot(Equal("1.16.3"))
			})

			It("replaces @current symlink with which links to <version>", func() {
				Expect(sut.ExecuteCommand(context.Background(), "devctl", []string{"use", "1.16.3"})).To(Succeed())
				currentFi, err := vs.Lstat(expectedCurrent)
				statCurrentFi, err := vs.Stat(expectedCurrent)
				_ = statCurrentFi

				Expect(expectedCurrent).Should(BeASymlink(vs))
				Expect(currentFi).ShouldNot(BeNil())
				Expect(err).Should(BeNil())

				expectedNewVersion := pp.SDK("go", "1.16.3")
				newVersionFi, err := vs.Lstat(expectedNewVersion)
				Expect(err).Should(BeNil())
				Expect(newVersionFi).ShouldNot(BeNil())
				Expect(expectedNewVersion).Should(BeADirectoryFs(vs))

				expectedOldVersion := pp.SDK("go", "19.5")
				oldFi, err := vs.Lstat(expectedOldVersion)
				Expect(oldFi).ShouldNot(BeNil())
				Expect(err).Should(BeNil())
				Expect(expectedOldVersion).Should(Or(BeADirectoryFs(vs), BeASymlink(vs)))
			})

			It("removes symlink from old to current", func() {
				Expect(sut.ExecuteCommand(context.Background(), "devctl", []string{"use", "1.16.3"})).To(Succeed())

				Expect(pp.SDK("go", "19.5")).Should(BeADirectoryFs(vs))
				_, err := vs.Readlink(pp.SDK("go", "19.5"))
				Expect(err).ShouldNot(Succeed())
			})
		})
	})
})
