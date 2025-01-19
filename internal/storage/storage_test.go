package storage_test

// import (
// 	"testing"
//
// 	"github.com/devchain-network/cauldron/internal/storage"
// 	"github.com/stretchr/testify/assert"
// )
//
// func TestGitProvider_String(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		provider storage.GitProvider
// 		expected string
// 	}{
// 		{
// 			name:     "github provider",
// 			provider: storage.GitProviderGitHub,
// 			expected: "github",
// 		},
// 		{
// 			name:     "gitlab provider",
// 			provider: storage.GitProviderGitLab,
// 			expected: "gitlab",
// 		},
// 		{
// 			name:     "bitbucket provider",
// 			provider: storage.GitProviderBitbucket,
// 			expected: "bitbucket",
// 		},
// 		{
// 			name:     "empty provider",
// 			provider: storage.GitProvider(""),
// 			expected: "",
// 		},
// 		{
// 			name:     "invalid provider",
// 			provider: storage.GitProvider("unknown"),
// 			expected: "unknown",
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			actual := tt.provider.String()
// 			assert.Equal(t, tt.expected, actual, "GitProvider.String() returned unexpected value")
// 		})
// 	}
// }
//
// func TestGitProvider_Valid(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		provider storage.GitProvider
// 		expected bool
// 	}{
// 		{
// 			name:     "valid github provider",
// 			provider: storage.GitProviderGitHub,
// 			expected: true,
// 		},
// 		{
// 			name:     "valid gitlab provider",
// 			provider: storage.GitProviderGitLab,
// 			expected: true,
// 		},
// 		{
// 			name:     "valid bitbucket provider",
// 			provider: storage.GitProviderBitbucket,
// 			expected: true,
// 		},
// 		{
// 			name:     "invalid empty provider",
// 			provider: storage.GitProvider(""),
// 			expected: false,
// 		},
// 		{
// 			name:     "invalid custom provider",
// 			provider: storage.GitProvider("unknown"),
// 			expected: false,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			actual := tt.provider.Valid()
// 			assert.Equal(t, tt.expected, actual, "GitProvider.Valid() returned unexpected result")
// 		})
// 	}
// }
