package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aerospike/tools-common-go/testutils"
	"github.com/stretchr/testify/suite"
)

var wd, _ = os.Getwd()

type CmdTestSuite struct {
	suite.Suite
	app              string
	coverFile        string
	coverFileCounter int
	host             string
	port             int
}

func TestDistanceMetricFlagSuite(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (suite *CmdTestSuite) SetupSuite() {
	suite.app = path.Join(wd, "app.test")
	suite.coverFile = path.Join(wd, "../coverage/cmd-coverage.cov")
	suite.coverFileCounter = 0
	suite.host = "127.0.0.1"
	suite.port = 10000

	err := docker_compose_up()
	if err != nil {
		suite.FailNowf("unable to start docker compose up", "%v", err)
	}

	os.Chdir("..")
	goArgs := []string{"test", "-coverpkg", "./...", "-tags=integration", "-o", suite.app}

	// Compile test binary
	compileCmd := exec.Command("go", goArgs...)
	err = compileCmd.Run()
	suite.Assert().NoError(err)
}

func (suite *CmdTestSuite) TearDownSuite() {
	docker_compose_down()
	err := os.Remove(suite.app)
	suite.Assert().NoError(err)
	time.Sleep(time.Second * 5)
	err = testutils.Stop()
	suite.Assert().NoError(err)
}

func (suite *CmdTestSuite) runCmd(asvecCmd ...string) ([]string, error) {
	strs := strings.Split(suite.coverFile, ".")
	file := strs[len(strs)-2] + "-" + strconv.Itoa(suite.coverFileCounter) + "." + strs[len(strs)-1]
	suite.coverFileCounter++
	var args []string
	args = []string{"-test.coverprofile=" + file}
	args = append(args, asvecCmd...)

	cmd := exec.Command(suite.app, args...)
	stdout, err := cmd.Output()
	// fmt.Printf("stdout: %v", string(stdout))

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return []string{string(ee.Stderr)}, err
		}
		return []string{string(stdout)}, err
	}

	lines := strings.Split(string(stdout), "\n")

	return lines, nil
}

func (suite *CmdTestSuite) TestSuccessfulCreateIndexCmd() {
	testCases := []struct {
		name string
		cmd  string
	}{
		{
			"",
			"",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)
		})
	}
}

func docker_compose_up() error {
	fmt.Println("Starting docker containers")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// docker/docker-compose.yml
	cmd := exec.CommandContext(ctx, "docker", "compose", fmt.Sprintf("-f /../../docker/docker-compose.yml"), "up", "-d")
	output, err := cmd.CombinedOutput()

	fmt.Printf("docker compose up output: %s\n", string(output))

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return err
		}
		return err
	}

	return nil
}

func docker_compose_down() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "compose", fmt.Sprintf("-f /../../docker/docker-compose.yml"), "down", "-d")
	_, err := cmd.Output()

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return err
		}
		return err
	}

	return nil
}
