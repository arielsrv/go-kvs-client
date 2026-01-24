package kvs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arielsrv/go-kvs-client/kvs"
)

func TestIsValidKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "valid key",
			key:      "valid-key",
			expected: true,
		},
		{
			name:     "empty key",
			key:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kvs.IsValidKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
