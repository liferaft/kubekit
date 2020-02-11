package cli

import "fmt"

// UserError this is an error caused by a user input, not by the system
type UserError struct {
	msg string
}

func (u UserError) Error() string { return u.msg }

// UserErrorf create an User Error similar to fmt.Errorf()
func UserErrorf(format string, a ...interface{}) UserError {
	return UserError{fmt.Sprintf(format, a...)}
}
