package bridgr_test

import (
	"net/url"
	"os"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/docker/docker/api/types/registry"
	"github.com/google/go-cmp/cmp"
)

func TestConjoined(t *testing.T) {
	tests := []struct {
		name   string
		cred   bridgr.Credential
		expect string
	}{
		{"user and password", bridgr.Credential{Username: "me", Password: "myself"}, "me:myself"},
		{"empty credential", bridgr.Credential{}, ":"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.cred.Conjoin()
			if result != test.expect {
				t.Errorf("Expected %s from Conjoin() but got %s", test.expect, result)
			}
		})
	}
}

func TestBase64(t *testing.T) {
	tests := []struct {
		name   string
		cred   bridgr.Credential
		expect string
	}{
		{"user and password", bridgr.Credential{Username: "me", Password: "myself"}, "bWU6bXlzZWxm"},
		{"empty creds", bridgr.Credential{}, ""},
		{"only username", bridgr.Credential{Username: "michael"}, "bWljaGFlbDo="},
		{"only password", bridgr.Credential{Password: "bluth"}, "OmJsdXRo"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.cred.Base64()
			if result != test.expect {
				t.Errorf("Expected %s but got %s", test.expect, result)
			}
		})
	}
}

func TestCredentialRead(t *testing.T) {
	tests := []struct {
		name   string
		envs   map[string]string
		expect bridgr.Credential
	}{
		{"user and password", map[string]string{"BRIDGR_BLUTH_COM_USER": "michael", "BRIDGR_BLUTH_COM_PASS": "boss"}, bridgr.Credential{Username: "michael", Password: "boss"}},
		{"empty creds", map[string]string{}, bridgr.Credential{}},
		{"only username", map[string]string{"BRIDGR_BLUTH_COM_USER": "michael"}, bridgr.Credential{Username: "michael"}},
		{"only password", map[string]string{"BRIDGR_BLUTH_COM_TOKEN": "bossman"}, bridgr.Credential{Password: "bossman"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := bridgr.WorkerCredentialReader{}
			src, _ := url.Parse("https://bluth.com/")
			for env, val := range test.envs {
				os.Setenv(env, val)
			}
			result, _ := reader.Read(src)
			if !cmp.Equal(test.expect, result) {
				t.Error(cmp.Diff(test.expect, result))
			}
			for env := range test.envs {
				os.Unsetenv(env)
			}
		})
	}
}

func TestDockerCredsWrite(t *testing.T) {
	docker := bridgr.DockerCredential{AuthConfig: registry.AuthConfig{}}
	expect := bridgr.DockerCredential{AuthConfig: registry.AuthConfig{Username: "tobias", Password: "themaninsideme"}}
	cred := bridgr.Credential{Username: "tobias", Password: "themaninsideme"}
	err := docker.Write(cred)

	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(expect, docker) {
		t.Error(cmp.Diff(expect, docker))
	}
}

func TestDockerCredsString(t *testing.T) {
	tests := []struct {
		name   string
		cred   bridgr.DockerCredential
		expect string
	}{
		{"just username", bridgr.DockerCredential{AuthConfig: registry.AuthConfig{Username: "buster"}}, "eyJ1c2VybmFtZSI6ImJ1c3RlciJ9"},
		{"user and password", bridgr.DockerCredential{AuthConfig: registry.AuthConfig{Username: "buster", Password: "monster!!"}}, "eyJ1c2VybmFtZSI6ImJ1c3RlciIsInBhc3N3b3JkIjoibW9uc3RlciEhIn0="},
		{"empty", bridgr.DockerCredential{AuthConfig: registry.AuthConfig{}}, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.cred.String()
			if !cmp.Equal(test.expect, result) {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name   string
		cred   bridgr.Credential
		expect bool
	}{
		{"user and password", bridgr.Credential{Username: "me", Password: "myself"}, true},
		{"empty creds", bridgr.Credential{}, false},
		{"only username", bridgr.Credential{Username: "michael"}, true},
		{"only password", bridgr.Credential{Password: "bluth"}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.cred.IsValid()
			if result != test.expect {
				t.Errorf("Expected %t but got %t", test.expect, result)
			}
		})
	}
}
