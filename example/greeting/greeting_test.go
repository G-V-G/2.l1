package greeting

import "testing"

func TestGreeting(t *testing.T) {
	if GetGreeting() != "Hello world" {
		t.Error("Invalid greeting")
	}
}