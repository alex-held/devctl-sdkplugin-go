package golang

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoSDKPluginSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "go-plugin USE")
}
