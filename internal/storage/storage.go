package storage

import (
	"context"
	"time"

	"github.com/IBM/sarama"
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

func validGitProviders() []GitProvider {
	return []GitProvider{
		GitProviderGitHub,
		GitProviderGitLab,
		GitProviderBitbucket,
	}
}

// PGPooler defines pgxpool behaviours.
type PGPooler interface {
	Close()
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// Pinger defines ping behaviour.
type Pinger interface {
	Ping(ctx context.Context, maxRetries uint8, backoff time.Duration) error
}

// MessageStorer defines kafka message store behaviour.
type MessageStorer interface {
	MessageStore(ctx context.Context, message *sarama.ConsumerMessage) error
}

// PingStorer is a combination of pinger and storer functionality.
type PingStorer interface {
	Pinger
	MessageStorer
}

// GitProvider represents git platform.
type GitProvider string

func (g GitProvider) String() string {
	return string(g)
}

// Valid checks if the GitProvider is valid.
func (g GitProvider) Valid() bool {
	for _, provider := range validGitProviders() {
		if g == provider {
			return true
		}
	}

	return false
}
