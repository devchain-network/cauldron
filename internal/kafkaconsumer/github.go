package kafkaconsumer

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/devchain-network/cauldron/internal/storage"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
)

// ExtractHeadersFromMessage extracts Kafka message headers.
func (Consumer) ExtractHeadersFromMessage(msg *sarama.ConsumerMessage) (*storage.GitHubWebhookData, error) {
	deliveryID, err := uuid.Parse(string(msg.Key))
	if err != nil {
		return nil, fmt.Errorf("kafkaconsumer.ExtractHeadersFromMessage deliveryID error: [%w]", err)
	}

	targetID, err := strconv.ParseUint(string(msg.Headers[2].Value), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("kafkaconsumer.ExtractHeadersFromMessage targetID error: [%w]", err)
	}

	hookID, err := strconv.ParseUint(string(msg.Headers[3].Value), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("kafkaconsumer.ExtractHeadersFromMessage hookID error: [%w]", err)
	}

	return &storage.GitHubWebhookData{
		DeliveryID: deliveryID,
		TargetID:   targetID,
		HookID:     hookID,
		Target:     string(msg.Headers[1].Value),
		Event:      github.Event(string(msg.Headers[0].Value)),
		Offset:     msg.Offset,
		Partition:  msg.Partition,
	}, nil
}

// UnmarshalPayload unmarshals incoming value to event related struct.
func (Consumer) UnmarshalPayload(event github.Event, value []byte, target string) (any, error) {
	var payload any

	switch event { //nolint:exhaustive
	case github.PingEvent:

		switch github.InstallationTargetType(target) {
		case github.InstallationTargetTypeOrganization:
			var pl github.PingOrganizationPayload
			if err := json.Unmarshal(value, &pl); err != nil {
				return nil, fmt.Errorf(
					"kafkaconsumer.UnmarshalPayload github.PingOrganizationPayload error: [%w]",
					err,
				)
			}
			payload = pl
		case github.InstallationTargetTypeRepository:
			var pl github.PingPayload
			if err := json.Unmarshal(value, &pl); err != nil {
				return nil, fmt.Errorf(
					"kafkaconsumer.UnmarshalPayload github.PingPayload error: [%w]",
					err,
				)
			}
			payload = pl
		default:
			var pl github.PingPayload
			if err := json.Unmarshal(value, &pl); err != nil {
				return nil, fmt.Errorf(
					"kafkaconsumer.UnmarshalPayload github.PingPayload error: [%w]",
					err,
				)
			}
			payload = pl
		}

	case github.CommitCommentEvent:
		var pl github.CommitCommentPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf(
				"kafkaconsumer.UnmarshalPayload github.CommitCommentPayload error: [%w]",
				err,
			)
		}
		payload = pl
	case github.CreateEvent:
		var pl github.CreatePayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.UnmarshalPayload github.CreatePayload error: [%w]", err)
		}
		payload = pl
	case github.DeleteEvent:
		var pl github.DeletePayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.UnmarshalPayload github.DeletePayload error: [%w]", err)
		}
		payload = pl
	case github.ForkEvent:
		var pl github.ForkPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf(
				"kafkaconsumer.UnmarshalPayload github.ForkPayload error: [%w]",
				err,
			)
		}
		payload = pl
	case github.GollumEvent:
		var pl github.GollumPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf(
				"kafkaconsumer.UnmarshalPayload github.GollumPayload error: [%w]",
				err,
			)
		}
		payload = pl
	case github.IssueCommentEvent:
		var pl github.IssueCommentPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.UnmarshalPayload github.IssueCommentPayload error: [%w]", err)
		}
		payload = pl
	case github.IssuesEvent:
		var pl github.IssuesPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.UnmarshalPayload github.IssuesPayload error: [%w]", err)
		}
		payload = pl
	case github.PullRequestEvent:
		var pl github.PullRequestPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.UnmarshalPayload github.PullRequestPayload error: [%w]", err)
		}
		payload = pl
	case github.PullRequestReviewCommentEvent:
		var pl github.PullRequestReviewCommentPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf(
				"kafkaconsumer.UnmarshalPayload github.PullRequestReviewCommentPayload error: [%w]",
				err,
			)
		}
		payload = pl
	case github.PullRequestReviewEvent:
		var pl github.PullRequestReviewPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.UnmarshalPayload github.PullRequestReviewPayload error: [%w]", err)
		}
		payload = pl
	case github.PushEvent:
		var pl github.PushPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.UnmarshalPayload github.PushPayload error: [%w]", err)
		}
		payload = pl
	case github.ReleaseEvent:
		var pl github.ReleasePayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.UnmarshalPayload github.ReleasePayload error: [%w]", err)
		}
		payload = pl
	case github.StarEvent:
		var pl github.StarPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf("kafkaconsumer.UnmarshalPayload github.StarPayload error: [%w]", err)
		}
		payload = pl
	case github.WatchEvent:
		var pl github.WatchPayload
		if err := json.Unmarshal(value, &pl); err != nil {
			return nil, fmt.Errorf(
				"kafkaconsumer.UnmarshalPayload github.WatchPayload error: [%w]",
				err,
			)
		}
		payload = pl
	}

	return payload, nil
}

// ExtractUserInfo extracts user information from payload.
func (Consumer) ExtractUserInfo(payload any) (int64, string) {
	switch p := payload.(type) {
	case github.PingPayload:
		return p.Sender.ID, p.Sender.Login
	case github.CommitCommentPayload:
		return p.Sender.ID, p.Sender.Login
	case github.CreatePayload:
		return p.Sender.ID, p.Sender.Login
	case github.DeletePayload:
		return p.Sender.ID, p.Sender.Login
	case github.ForkPayload:
		return p.Sender.ID, p.Sender.Login
	case github.GollumPayload:
		return p.Sender.ID, p.Sender.Login
	case github.IssueCommentPayload:
		return p.Sender.ID, p.Sender.Login
	case github.IssuesPayload:
		return p.Sender.ID, p.Sender.Login
	case github.PullRequestPayload:
		return p.Sender.ID, p.Sender.Login
	case github.PullRequestReviewCommentPayload:
		return p.Sender.ID, p.Sender.Login
	case github.PullRequestReviewPayload:
		return p.Sender.ID, p.Sender.Login
	case github.PushPayload:
		return p.Sender.ID, p.Sender.Login
	case github.ReleasePayload:
		return p.Sender.ID, p.Sender.Login
	case github.StarPayload:
		return p.Sender.ID, p.Sender.Login
	case github.WatchPayload:
		return p.Sender.ID, p.Sender.Login
	}

	return 0, "user-not-found"
}

// StoreGitHubMessage stores consumed message to database.
func (c Consumer) StoreGitHubMessage(msg *sarama.ConsumerMessage) error {
	storagePayload, err := c.ExtractHeadersFromMessage(msg)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage ExtractHeadersFromMessage error: [%w]", err)
	}

	payload, err := c.UnmarshalPayload(storagePayload.Event, msg.Value, storagePayload.Target)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage UnmarshalPayload error: [%w]", err)
	}

	userID, userLogin := c.ExtractUserInfo(payload)

	storagePayload.UserID = userID
	storagePayload.UserLogin = userLogin
	storagePayload.Payload = payload

	if err = c.Storage.GitHubStore(storagePayload); err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage Storage.GitHubStore error: [%w]", err)
	}

	return nil
}
