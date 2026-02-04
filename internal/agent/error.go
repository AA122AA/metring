package agent

type ReqError struct {
	err error
}

func (re *ReqError) Unwrap() error {
	return re.err
}

func (re *ReqError) Error() string {
	return re.err.Error()
}

func (re *ReqError) Is(target error) bool {
	_, ok := target.(*ReqError)
	return ok
}

func NewReqError(err error) *ReqError {
	return &ReqError{
		err: err,
	}
}
