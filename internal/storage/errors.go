package storage

import "fmt"

type TransportAuthMismatchError struct {
	URL                string
	ExpectedAuthMethod string
}

func (e *TransportAuthMismatchError) Error() string {
	return fmt.Sprintf("mismatched or no auth method (%s) for url: %s", e.ExpectedAuthMethod, e.URL)
}

type SSHKeyError struct {
	SSHKeyPath string
	OpErr      error
}

func (e *SSHKeyError) Error() string {
	return fmt.Sprintf("failed to create ssh auth for %s: %s", e.SSHKeyPath, e.OpErr)
}
