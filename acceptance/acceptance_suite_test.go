package acceptance_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var volmanPath string
var volmanServerPort int
var debugServerAddress string
var volmanRunner *ginkgomon.Runner

var driverPath string
var driverServerPort int
var debugServerAddress2 string
var unixDriverRunner *ginkgomon.Runner
var tmpDriversPath string

var keyringFileContents string
var clusterIp string

func TestVolman(t *testing.T) {
	// these integration tests can take a bit, especially under load;
	// 1 second is too harsh
	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "RexRay Driver Cmd Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	volmanPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/cmd/volman", "-race")
	Expect(err).NotTo(HaveOccurred())

	driverPath = "/home/ubuntu/funnyrexray" //todo: inject via ENV varible

	return []byte(strings.Join([]string{volmanPath, driverPath}, ","))
}, func(pathsByte []byte) {
	path := string(pathsByte)
	volmanPath = strings.Split(path, ",")[0]
	driverPath = strings.Split(path, ",")[1]

	// read config files
	// keyringFileContents = os.Getenv("rexray_KEYRING")
	// clusterIp = os.Getenv("rexray_CLUSTER_IP")
})

var _ = BeforeEach(func() {
	var err error
	tmpDriversPath = "/etc/docker/plugins/"
	Expect(err).NotTo(HaveOccurred())

	commandWithEnv := exec.Command(driverPath)
	currentEnv := os.Environ()
	newEnv := append(currentEnv, "SCALEIO_PROTECTIONDOMAINNAME=default")
	newEnv = append(newEnv, "SCALEIO_USERID=fc5118f500000000")
	newEnv = append(newEnv, "SCALEIO_STORAGEPOOLNAME=default")
	newEnv = append(newEnv, "SCALEIO_SYSTEMNAME=")
	newEnv = append(newEnv, "SCALEIO_PROTECTIONDOMAINID=4c9e0aa100000000")
	newEnv = append(newEnv, "SCALEIO_THINORTHICK=ThinProvisioned")
	newEnv = append(newEnv, "SCALEIO_SYSTEMID=5c4c45f02c98551c")
	newEnv = append(newEnv, "REXRAY_STORAGEDRIVERS=scaleio")
	newEnv = append(newEnv, "SCALEIO_PASSWORD=Some1Example")
	newEnv = append(newEnv, "SCALEIO_USERNAME=admin")
	newEnv = append(newEnv, "SCALEIO_USECERTS=true")
	newEnv = append(newEnv, "SCALEIO_ENDPOINT=https://52.87.123.252/api")
	newEnv = append(newEnv, "SCALEIO_STORAGEPOOLID=8b3b0c5300000000")
	newEnv = append(newEnv, "SCALEIO_INSECURE=true")
	newEnv = append(newEnv, "SCALEIO_VERSION=2.0")
	commandWithEnv.Env = newEnv

	unixDriverRunner = ginkgomon.New(ginkgomon.Config{
		Name:       "RexRayServer",
		Command:    commandWithEnv,
		StartCheck: "started",
	})

	volmanServerPort = 8750 + GinkgoParallelNode()
	debugServerAddress = fmt.Sprintf("0.0.0.0:%d", 8850+GinkgoParallelNode())
	volmanRunner = ginkgomon.New(ginkgomon.Config{
		Name: "volman",
		Command: exec.Command(
			volmanPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", volmanServerPort),
			"-debugAddr", debugServerAddress,
			"-driversPath", tmpDriversPath,
		),
		StartCheck: "volman.started",
	})
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
