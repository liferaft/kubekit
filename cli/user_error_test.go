package cli

import (
	"fmt"
	"testing"
)

func TestUserError(t *testing.T) {
	errFactory := func() error {
		return UserErrorf("this is a user error")
	}

	err := errFactory()
	if _, ok := err.(UserError); !ok {
		t.Errorf("UserErrorf() is not returning a User Error")
	}

	want := "error: this is a user error"
	if got := fmt.Sprintf("error: %s", err); got != want {
		t.Errorf("UserError.Error() = %v; want %v", got, want)
	}

	err2 := fmt.Errorf("this is a system error")
	if _, ok := err2.(UserError); ok {
		t.Errorf("Error is considered an UserError")
	}
}
