package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAction_InheritedMethods(t *testing.T) {
	action := Action{
		Human: Human{name: "Alice", age: 30},
	}

	t.Run("Set and Get Name", func(t *testing.T) {
		action.SetName("Bob")
		assert.Equal(t, "Bob", action.GetName())
	})

	t.Run("Set and Get Age", func(t *testing.T) {
		action.SetAge(25)
		assert.Equal(t, 25, action.GetAge())
	})
}
