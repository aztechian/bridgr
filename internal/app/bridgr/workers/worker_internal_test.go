package workers

import (
	"os"
)

func resetTestEnvCredentials() {
	os.Unsetenv("BRIDGR_TEST_DOCKER_ORG_USER")
	os.Unsetenv("BRIDGR_TEST_DOCKER_ORG_PASS")
	os.Unsetenv("BRIDGR_TEST_DOCKER_ORG_TOKEN")
}
