package mapenv

import "fmt"

type DecodeError struct {
	description string
	field       string
	err         error
}

func newDecodeError(description string, field string, err error) DecodeError {
	return DecodeError{description: description, field: field, err: err}
}

func (d DecodeError) Description() string {
	return d.description
}

func (d DecodeError) Field() string {
	return d.field
}

func (d DecodeError) Err() error {
	return d.err
}

func (d DecodeError) Error() string {
	if d.err == nil {
		return d.description
	}
	return fmt.Sprintf("%s: err %v", d.description, d.err)
}
