package errors

import "errors"

var (
	ErrEmaiAlreadyTaken         = errors.New("email already taken")
	ErrSomethingWentWrong       = errors.New("Something went wrong")
	ErrInvalidCredentials       = errors.New("Invalid Credentials")
	ErrEmailNotVerified         = errors.New("Email Not Verified")
	ErrCantSendVerificationMail = errors.New("Cant Resend Verification Mail")
)
