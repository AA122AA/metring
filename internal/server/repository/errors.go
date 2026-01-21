package repository

import "fmt"

type EmptyRepoError struct {
	err error
}

func (er *EmptyRepoError) Error() string {
	return er.err.Error()
}

func (er *EmptyRepoError) Unwrap() error {
	return er.err
}

func (re *EmptyRepoError) Is(target error) bool {
	_, ok := target.(*EmptyRepoError)
	return ok
}

func NewEmptyRepoError(err error) *EmptyRepoError {
	if err == nil {
		err = fmt.Errorf("no data")
	}
	return &EmptyRepoError{
		err: fmt.Errorf("%w", err),
	}
}
