package workers

import (
	"encoding/base64"
	"net/url"
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
