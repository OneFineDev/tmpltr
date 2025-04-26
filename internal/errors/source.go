package package_errors

import (
	"strings"

	"github.com/pkg/errors"
)

var (
	OpenSourceConfigFileError  = "error opening source config file: %w"
	OpenValuesFileError        = "error opening values file: %w"
	ParseSourceConfigFileError = "error parsing source config file: %w"
	ParseSValuesFileError      = "error parsing values file: %w"
	BuildSourceConfigError     = "error building source configs: %w"
	TemplateExecutionError     = "error executing template: %w"
	TemplateFileRenameError    = "error renaming template file: %w"
)

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

func FlattenCloneErrors(errs ...error) error {
	errorsString := strings.Builder{}

	for _, e := range errs {
		errorsString.WriteString(e.Error())
		errorsString.WriteString("; ")
	}

	es := errorsString.String()

	return errors.Errorf("the following errors occured while cloning: %s", es)
}
