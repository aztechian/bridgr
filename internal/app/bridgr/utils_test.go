package bridgr_test

import "testing"
import "bridgr/internal/app/bridgr"

func TestDockerImage(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"bare image", "myimg", "library/myimg:latest"},
		{"img with ver", "myimg:1.2", "library/myimg:1.2"},
		{"repo underscore", "_/myimg", "library/myimg:latest"},
		{"canonical", "library/myimg:1.3", "library/myimg:1.3"},
		{"private repo", "myreg.com/project/myimg", "myreg.com/project/myimg:latest"},
		{"private repo with version", "myreg.com/project/otherimage:4.21", "myreg.com/project/otherimage:4.21"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := bridgr.DockerImage(test.input)
			if got != test.expect {
				t.Errorf("Got docker image of %s, was expecting %s", got, test.expect)
			}
		})
	}
}
