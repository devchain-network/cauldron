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

// StoreGitHubMessage stores consumed message to database.
func (c Consumer) StoreGitHubMessage(msg *sarama.ConsumerMessage) error {
	deliveryID, err := uuid.Parse(string(msg.Key))
	if err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage deliveryID error: [%w]", err)
	}

	targetID, err := strconv.ParseUint(string(msg.Headers[2].Value), 10, 64)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage targetID error: [%w]", err)
	}

	hookID, err := strconv.ParseUint(string(msg.Headers[3].Value), 10, 64)
	if err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage hookID error: [%w]", err)
	}

	target := string(msg.Headers[1].Value)
	event := github.Event(string(msg.Headers[0].Value))
	offset := msg.Offset
	partition := msg.Partition

	var payload any

	switch event { //nolint:exhaustive
	case github.CommitCommentEvent:
		var pl github.CommitCommentPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf(
				"kafkaconsumer.storeGitHubMessage github.CommitCommentPayload error: [%w]",
				err,
			)
		}
		payload = pl
	case github.CreateEvent:
		var pl github.CreatePayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.CreatePayload error: [%w]", err)
		}
		payload = pl
	case github.DeleteEvent:
		var pl github.DeletePayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.DeletePayload error: [%w]", err)
		}
		payload = pl
	case github.ForkEvent:
		var pl github.ForkPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf(
				"kafkaconsumer.storeGitHubMessage github.ForkPayload error: [%w]",
				err,
			)
		}
		payload = pl
	case github.GollumEvent:
		var pl github.GollumPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf(
				"kafkaconsumer.storeGitHubMessage github.GollumPayload error: [%w]",
				err,
			)
		}
		payload = pl
	case github.IssueCommentEvent:
		var pl github.IssueCommentPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.IssueCommentPayload error: [%w]", err)
		}
		payload = pl
	case github.IssuesEvent:
		var pl github.IssuesPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.IssuesPayload error: [%w]", err)
		}
		payload = pl
	case github.PullRequestEvent:
		var pl github.PullRequestPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.PullRequestPayload error: [%w]", err)
		}
		payload = pl
	case github.PullRequestReviewCommentEvent:
		var pl github.PullRequestReviewCommentPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf(
				"kafkaconsumer.storeGitHubMessage github.PullRequestReviewCommentPayload error: [%w]",
				err,
			)
		}
		payload = pl
	case github.PullRequestReviewEvent:
		var pl github.PullRequestReviewPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.PullRequestReviewPayload error: [%w]", err)
		}
		payload = pl
	case github.PushEvent:
		var pl github.PushPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.PushPayload error: [%w]", err)
		}
		payload = pl
	case github.ReleaseEvent:
		var pl github.ReleasePayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.ReleasePayload error: [%w]", err)
		}
		payload = pl
	case github.StarEvent:
		var pl github.StarPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf("kafkaconsumer.storeGitHubMessage github.StarPayload error: [%w]", err)
		}
		payload = pl
	case github.WatchEvent:
		var pl github.WatchPayload
		if err = json.Unmarshal(msg.Value, &pl); err != nil {
			return fmt.Errorf(
				"kafkaconsumer.storeGitHubMessage github.WatchPayload error: [%w]",
				err,
			)
		}
		payload = pl
	}

	var userID int64
	var userLogin string

	switch payload := payload.(type) {
	case github.CommitCommentPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.CreatePayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.DeletePayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.ForkPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.GollumPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.IssueCommentPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.IssuesPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.PullRequestPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.PullRequestReviewCommentPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.PullRequestReviewPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.PushPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.ReleasePayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.StarPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	case github.WatchPayload:
		userID = payload.Sender.ID
		userLogin = payload.Sender.Login
	}

	storagePayload := storage.GitHubWebhookData{
		DeliveryID: deliveryID,
		Event:      event,
		Target:     target,
		TargetID:   targetID,
		HookID:     hookID,
		Offset:     offset,
		Partition:  partition,
		UserID:     userID,
		UserLogin:  userLogin,
		Payload:    payload,
	}

	if err = c.Storage.GitHubStore(&storagePayload); err != nil {
		return fmt.Errorf("kafkaconsumer.storeGitHubMessage Storage.GitHubStore error: [%w]", err)
	}

	return nil
}
