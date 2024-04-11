package utils_test

import (
	"encoder/framework/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsJson(t *testing.T) {
	t.Parallel()

	t.Run("Should return nil when JSON is valid", func(t *testing.T) {
		t.Parallel()
		json := `{"test": "test"}`
		err := utils.IsJson(json)
		require.Nil(t, err)
	})

	t.Run("Should return error when JSON is invalid", func(t *testing.T) {
		t.Parallel()
		json := `{"test": "test"`
		err := utils.IsJson(json)
		require.NotNil(t, err)
	})
}
