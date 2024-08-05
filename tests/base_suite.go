//go:build unit || integration

package tests

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	avs "github.com/aerospike/avs-client-go"
	"github.com/stretchr/testify/suite"
)

type CmdBaseTestSuite struct {
	suite.Suite
	App          string
	ComposeFile  string
	CoverFile    string
	AvsHostPort  *avs.HostPort
	AvsTLSConfig *tls.Config
	AvsCreds     *avs.UserPassCredentials
	AvsClient    *avs.AdminClient
	SuiteFlags   []string
	Logger       *slog.Logger
}

var wd, _ = os.Getwd()

func (suite *CmdBaseTestSuite) SetupSuite() {
	suite.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	suite.App = path.Join(wd, "app.test")
	suite.CoverFile = path.Join(wd, "../coverage/cmd-coverage.cov")

	err := DockerComposeUp(suite.ComposeFile)

	time.Sleep(time.Second * 10)

	if err != nil {
		suite.FailNowf("unable to start docker compose up", "%v", err)
	}

	goArgs := []string{"build", "-cover", "-coverpkg", "./...", "-o", suite.App}

	// Compile test binary
	compileCmd := exec.Command("go", goArgs...)
	out, err := compileCmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Couldn't compile test bin stdout/err: %v\n", string(out))
	}

	suite.Assert().NoError(err)

	suite.AvsClient, err = GetAdminClient(suite.AvsHostPort, suite.AvsCreds, suite.AvsTLSConfig, suite.Logger)
	if err != nil {
		suite.FailNowf("unable to create admin client", "%v", err)
	}
}

func (suite *CmdBaseTestSuite) TearDownSuite() {
	err := os.Remove(suite.App)
	suite.Assert().NoError(err)
	time.Sleep(time.Second * 5)
	suite.Assert().NoError(err)
	suite.AvsClient.Close()

	err = DockerComposeDown(suite.ComposeFile)
	if err != nil {
		fmt.Println("unable to stop docker compose down")
	}
}

// All this does is append the suite flags to args because certain runs (e.g.
// flag parse error tests) should not append this flags
func (suite *CmdBaseTestSuite) RunSuiteCmd(asvecCmd ...string) (string, string, error) {
	suiteFlags := strings.Split(strings.Join(suite.SuiteFlags, " "), " ")
	asvecCmd = append(suiteFlags, asvecCmd...)
	return suite.RunCmd(asvecCmd...)
}

func (suite *CmdBaseTestSuite) RunCmd(asvecCmd ...string) (string, string, error) {
	suite.Logger.Info("running command", slog.String("cmd", strings.Join(asvecCmd, " ")))
	cmd := exec.Command(suite.App, asvecCmd...)
	cmd.Env = []string{"GOCOVERDIR=" + os.Getenv("COVERAGE_DIR")}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}
