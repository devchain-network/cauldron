package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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

// PGPooler defines pgxpool behaviours.
type PGPooler interface {
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// Pinger defines ping behaviour.
type Pinger interface {
	Ping(ctx context.Context, maxRetries uint8, backoff time.Duration) error
}

// Storer defines store behaviour.
type Storer interface {
	Store(ctx context.Context, payload any) error
}

// PingStorer is a combination of pinger and storer functionality.
type PingStorer interface {
	Pinger
	Storer
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
