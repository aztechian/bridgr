package workers_test

import (
	"bridgr/internal/app/bridgr/workers"
	"testing"
)

func TestConjoined(t *testing.T) {
	tests := []struct {
		name   string
		cred   workers.Credential
		expect string
	}{
		{"user and password", workers.Credential{Username: "me", Password: "myself"}, "me:myself"},
		{"empty credential", workers.Credential{}, ":"},
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
		cred   workers.Credential
		expect string
	}{
		{"user and password", workers.Credential{Username: "me", Password: "myself"}, "bWU6bXlzZWxm"},
		{"empty creds", workers.Credential{}, ""},
		{"only username", workers.Credential{Username: "michael"}, "bWljaGFlbDo="},
		{"only password", workers.Credential{Password: "bluth"}, "OmJsdXRo"},
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
