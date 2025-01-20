package storage_test

import (
	"testing"

	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestGitProvider_String(t *testing.T) {
	tests := []struct {
		name     string
		provider storage.GitProvider
		expected string
	}{
		{"GitHub", storage.GitProviderGitHub, "github"},
		{"GitLab", storage.GitProviderGitLab, "gitlab"},
		{"Bitbucket", storage.GitProviderBitbucket, "bitbucket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.provider.String())
		})
	}
}

func TestGitProvider_Valid(t *testing.T) {
	tests := []struct {
		name     string
		provider storage.GitProvider
		expected bool
	}{
		{"valid GitHub", storage.GitProviderGitHub, true},
		{"valid GitLab", storage.GitProviderGitLab, true},
		{"valid Bitbucket", storage.GitProviderBitbucket, true},
		{"invalid Provider", storage.GitProvider("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.provider.Valid())
		})
	}
}
