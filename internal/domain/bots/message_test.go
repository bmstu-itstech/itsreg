package bots_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestNewMessage(t *testing.T) {
	_, err := bots.NewMessage("abc")
	require.NoError(t, err)

	_, err = bots.NewMessage("")
	require.Error(t, err)
	require.EqualError(t, err, "expected not empty text in message")
}

func TestMessage_String(t *testing.T) {
	msg := bots.MustNewMessage("abc")
	require.Equal(t, "abc", msg.String())
}

func TestMessage_Text(t *testing.T) {
	msg := bots.MustNewMessage("abc")
	require.Equal(t, "abc", msg.Text())
}

func TestMessage_Merge(t *testing.T) {
	ab := bots.MustNewMessage("ab")
	cd := bots.MustNewMessage("cd")
	require.Equal(t, bots.MustNewMessage("ab\ncd"), ab.Merge(cd))
}

func TestMessage_PromoteToBotMessage(t *testing.T) {
	t.Run("promote with nil options", func(t *testing.T) {
		msg := bots.MustNewMessage("abc")
		got := msg.PromoteToBotMessage(nil)
		require.Equal(t, msg.Text(), got.Text())
		require.Equal(t, msg.String(), got.String())
		require.Empty(t, got.Options())
	})

	t.Run("promote with empty options", func(t *testing.T) {
		msg := bots.MustNewMessage("abc")
		got := msg.PromoteToBotMessage(make([]bots.Option, 0))
		require.Equal(t, msg.Text(), got.Text())
		require.Equal(t, msg.String(), got.String())
		require.Empty(t, got.Options())
	})

	t.Run("promote with not nil options", func(t *testing.T) {
		msg := bots.MustNewMessage("abc")
		opts := []bots.Option{
			"a", "b", "c",
		}
		got := msg.PromoteToBotMessage(opts)
		require.Equal(t, msg.Text(), got.Text())
		require.Equal(t, msg.String(), got.String())
		require.Equal(t, opts, got.Options())
	})
}
