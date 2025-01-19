package storage

import (
	"time"
)

// constants.
const (
	DefaultDBPingBackoff    = 2 * time.Second
	DefaultDBPingMaxRetries = 10
	DefaultDBPingTimeout    = 5 * time.Second

	GitProviderGitHub    GitProvider = "github"
	GitProviderGitLab    GitProvider = "gitlab"
	GitProviderBitbucket GitProvider = "bitbucket"
)

var validGitProviders = []GitProvider{
	GitProviderGitHub,
	GitProviderGitLab,
	GitProviderBitbucket,
}

// Pinger ...
type Pinger interface {
	Ping(maxRetries uint8, backoff time.Duration) error
}

// GitProvider represents git platform.
type GitProvider string

func (g GitProvider) String() string {
	return string(g)
}

// Valid checks if the GitProvider is valid.
func (g GitProvider) Valid() bool {
	for _, provider := range validGitProviders {
		if g == provider {
			return true
		}
	}

	return false
}
