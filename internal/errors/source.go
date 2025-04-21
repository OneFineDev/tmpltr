package package_errors

type SourceError struct {
	Message string
	Err     error
}

func (s SourceError) Error() string {
	return s.Message
}
func (s SourceError) Unwrap() error {
	return s.Err
}
