package basehttphandler

// Consumer ...
type Consumer interface {
	Consume(workedID int)
}

// // Handler ...
// type Handler struct {
// 	Logger               *slog.Logger
// 	KafkaProducer        sarama.AsyncProducer
// 	ProducerMessageQueue chan *sarama.ProducerMessage
// }
