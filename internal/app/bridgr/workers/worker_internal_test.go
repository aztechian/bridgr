package workers

import "net/url"
import "testing"
import "os"

func TestCredentials(t *testing.T) {
	url, _ := url.Parse("https://test.docker.org")
	u, p := credentials(url)
	if u != "" {
		t.Errorf("Expected user to be blank for %s", url)
	}

	os.Setenv("BRIDGR_TEST_DOCKER_ORG_USER", "myuser")
	os.Setenv("BRIDGR_TEST_DOCKER_ORG_PASS", "mypassword")
	u, p = credentials(url)
	if u == "" {
		t.Errorf("Expected a value for BRIDGR_TEST_DOCKER_ORG_USER")
	}
	t.Logf("User: %s, Pass: %s", u, p)
}
