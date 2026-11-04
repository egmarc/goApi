package api

import "testing"

func TestNewHandlerRejectsInvalidToken(t *testing.T) {
	for _, token := range []string{"", " ", "two words", "token\n", "\ttoken"} {
		t.Run(token, func(t *testing.T) {
			if _, err := NewHandler(token); err == nil {
				t.Fatal("expected an invalid token to be rejected")
			}
		})
	}
}
