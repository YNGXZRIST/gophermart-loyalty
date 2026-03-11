package labelerrors

import "fmt"

type LabelError struct {
	Label string
	Err   error
}

func (e LabelError) Error() string {
	return fmt.Sprintf("[%s]: %s", e.Label, e.Err)
}
func NewLabelError(label string, err error) error {
	return LabelError{Label: label, Err: err}
}
func (e LabelError) Unwrap() error { return e.Err }
