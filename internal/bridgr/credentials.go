package bridgr

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
)

// CredentialReader is the interface that wraps a Credential Read method
// The Reader should get credential information from somewhere (usually the runtime environment), and returns
// a Credential struct as well as a boolean that indicates whether the reading of credential pieces (ie, Username, Password and/or Token) was successful.
// Success is not strictly whether the pieces of information were found, but whether they existed in the environment to be read. A Credential struct with
// empty strings for its fields is still a valid Credential
type CredentialReader interface {
	Read(*url.URL) (Credential, bool)
}

// CredentialWriter is the interface wrapping the Write method for Credentials
// Write may write a credential to any form (string, another struct), and any medium (memory, file, etc).
type CredentialWriter interface {
	Write(Credential) error
}

// CredentialReaderWriter is an interface that combines both the CredentialReader and CredentialWriter
type CredentialReaderWriter interface {
	CredentialReader
	CredentialWriter
}

// Credential encapsulates a username/password pair
type Credential struct {
	Username string
	Password string
}

// IsValid returns a boolean to indicate that the Credential has at least a Username or a Password
func (c *Credential) IsValid() bool {
	return len(c.Username) > 0 || len(c.Password) > 0
}

// Conjoin returns a string with the Credential content joined by a ':' (colon character)
func (c *Credential) Conjoin() string {
	return c.Username + ":" + c.Password
}

// Base64 returns a string with the provided Credential content base64 encoded after being joined by ':' (colon character)
func (c *Credential) Base64() string {
	value := c.Conjoin()
	if value == ":" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(value))
}

// WorkerCredentialReader reads a credential for a URL from the environment variables
type WorkerCredentialReader struct{}

func (w *WorkerCredentialReader) Read(url *url.URL) (Credential, bool) {
	basename := "BRIDGR_" + strings.ToUpper(strings.ReplaceAll(url.Hostname(), ".", "_"))
	Debugf("Looking up credentials for: %s", basename)
	found := false
	userVal, ok := os.LookupEnv(basename + "_USER")
	found = found || ok
	passwdVal := ""
	if pw, ok := os.LookupEnv(basename + "_PASS"); ok {
		passwdVal = pw
		found = found || ok
	} else {
		token, tok := os.LookupEnv(basename + "_TOKEN")
		passwdVal = token
		found = found || tok
	}
	return Credential{
		Username: userVal,
		Password: passwdVal,
	}, found
}

// DockerCredential implements the CredentialReader and CredentialWriter interface for the Docker "login" format
type DockerCredential struct {
	types.AuthConfig
	WorkerCredentialReader
}

func (credWriter *DockerCredential) Write(c Credential) error {
	credWriter.Username = c.Username
	credWriter.Password = c.Password
	return nil
}

func (credWriter *DockerCredential) String() string {
	if credWriter.Username == "" && credWriter.Password == "" {
		return ""
	}
	jsonAuth, _ := json.Marshal(credWriter)
	return base64.URLEncoding.EncodeToString(jsonAuth)
}
