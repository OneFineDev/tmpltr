package storage

import "fmt"

type TransportAuthMismatchError struct {
	Url                string
	ExpectedAuthMethod string
}

func (e *TransportAuthMismatchError) Error() string {
	return fmt.Sprintf("mismatched or no auth method (%s) for url: %s", e.ExpectedAuthMethod, e.Url)
}

type SshKeyError struct {
	SshKeyPath string
	OpErr      error
}

func (e *SshKeyError) Error() string {
	return fmt.Sprintf("failed to create ssh auth for %s: %s", e.SshKeyPath, e.OpErr)
}
