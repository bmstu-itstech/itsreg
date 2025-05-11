package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestMessage_String(t *testing.T) {
	msg := bots.NewMessage("abc")
	require.Equal(t, "abc", msg.String())
}

func TestMessage_Text(t *testing.T) {
	msg := bots.NewMessage("abc")
	require.Equal(t, "abc", msg.Text())
}

func TestMessage_Merge(t *testing.T) {
	ab := bots.NewMessage("ab")
	cd := bots.NewMessage("cd")
	require.Equal(t, bots.NewMessage("ab\ncd"), ab.Merge(cd))
}
