package kafkaconsumer_test

// import (
// 	"testing"
//
// 	"github.com/IBM/sarama"
// 	"github.com/devchain-network/cauldron/internal/kafkaconsumer"
// 	"github.com/devchain-network/cauldron/internal/storage"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// )
//
// func TestExtractHeadersFromMessage(t *testing.T) {
// 	deliveryID := uuid.New()
// 	headers := []*sarama.RecordHeader{
// 		{Key: []byte("event"), Value: []byte("push")},
// 		{Key: []byte("target"), Value: []byte("repo-name")},
// 		{Key: []byte("targetID"), Value: []byte("123456")},
// 		{Key: []byte("hookID"), Value: []byte("654321")},
// 	}
// 	mockMsg := &sarama.ConsumerMessage{
// 		Key:       []byte(deliveryID.String()),
// 		Headers:   headers,
// 		Offset:    42,
// 		Partition: 1,
// 	}
//
// 	consumer := kafkaconsumer.Consumer{}
// 	result, err := consumer.ExtractHeadersFromMessage(mockMsg)
//
// 	assert.NoError(t, err)
//
// 	assert.Equal(t, deliveryID, result.DeliveryID)
// 	assert.Equal(t, uint64(123456), result.TargetID)
// 	assert.Equal(t, uint64(654321), result.HookID)
// 	assert.Equal(t, "repo-name", result.Target)
// 	assert.Equal(t, storage.GitHubWebhookData{
// 		DeliveryID: deliveryID,
// 		TargetID:   123456,
// 		HookID:     654321,
// 		Target:     "repo-name",
// 		Event:      "push",
// 		Offset:     42,
// 		Partition:  1,
// 	}, *result)
// }
//
// func TestExtractHeadersFromMessage_InvalidDeliveryID(t *testing.T) {
// 	mockMsg := &sarama.ConsumerMessage{
// 		Key: []byte("invalid-uuid"),
// 	}
//
// 	consumer := kafkaconsumer.Consumer{}
// 	result, err := consumer.ExtractHeadersFromMessage(mockMsg)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, result)
// 	assert.Contains(t, err.Error(), "deliveryID error")
// }
//
// func TestExtractHeadersFromMessage_InvalidTargetID(t *testing.T) {
// 	headers := []*sarama.RecordHeader{
// 		{Key: []byte("event"), Value: []byte("push")},
// 		{Key: []byte("target"), Value: []byte("repo-name")},
// 		{Key: []byte("targetID"), Value: []byte("invalid")},
// 		{Key: []byte("hookID"), Value: []byte("654321")},
// 	}
// 	mockMsg := &sarama.ConsumerMessage{
// 		Key:     []byte(uuid.New().String()),
// 		Headers: headers,
// 	}
//
// 	consumer := kafkaconsumer.Consumer{}
// 	result, err := consumer.ExtractHeadersFromMessage(mockMsg)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, result)
// 	assert.Contains(t, err.Error(), "targetID error")
// }
//
// func TestExtractHeadersFromMessage_InvalidHookID(t *testing.T) {
// 	headers := []*sarama.RecordHeader{
// 		{Key: []byte("event"), Value: []byte("push")},
// 		{Key: []byte("target"), Value: []byte("repo-name")},
// 		{Key: []byte("targetID"), Value: []byte("123456")},
// 		{Key: []byte("hookID"), Value: []byte("invalid")},
// 	}
// 	mockMsg := &sarama.ConsumerMessage{
// 		Key:     []byte(uuid.New().String()),
// 		Headers: headers,
// 	}
//
// 	consumer := kafkaconsumer.Consumer{}
// 	result, err := consumer.ExtractHeadersFromMessage(mockMsg)
//
// 	assert.Error(t, err)
// 	assert.Nil(t, result)
// 	assert.Contains(t, err.Error(), "hookID error")
// }
