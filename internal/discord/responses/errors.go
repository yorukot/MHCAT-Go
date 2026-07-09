package responses

import "errors"

var (
	ErrAlreadyResponded  = errors.New("interaction already has an initial response")
	ErrNoInitialResponse = errors.New("interaction has no initial response")
	ErrEphemeralChanged  = errors.New("interaction ephemeral state cannot be changed after defer")
	ErrInvalidModal      = errors.New("interaction modal is invalid")
)

const safeInternalErrorMessage = "Something went wrong while handling this interaction."

func SafeErrorMessage(err error) Message {
	if err == nil {
		return Message{Content: safeInternalErrorMessage, Ephemeral: true}
	}
	return Message{Content: safeInternalErrorMessage, Ephemeral: true}
}
