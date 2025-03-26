package kafka

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaClientManager struct {
	client *kgo.Client
	once   sync.Once
}

// NewKafkaClientManager creates a new KafkaClientManager
func NewKafkaClientManager() *KafkaClientManager {
	return &KafkaClientManager{}
}

// GetClient returns a singleton Kafka client
func (m *KafkaClientManager) GetClient() (*kgo.Client, error) {
	var err error
	m.once.Do(func() {
		// Read Kafka configuration from environment variables
		kafkaURL := os.Getenv("KAFKA_URL")
		if kafkaURL == "" {
			err = fmt.Errorf("KAFKA_URL environment variable is not set")
			return
		}

		// Configure Kafka client options
		opts := []kgo.Opt{
			kgo.SeedBrokers(kafkaURL),
			kgo.ConsumerGroup("webhook-payload-1"),
			kgo.RecordPartitioner(kgo.RoundRobinPartitioner()),
			kgo.ProducerBatchCompression(kgo.SnappyCompression()),
			kgo.ProduceRequestTimeout(10 * time.Second),
			kgo.RequiredAcks(kgo.AllISRAcks()),
		}

		// Create Kafka client
		m.client, err = kgo.NewClient(opts...)
	})

	if err != nil {
		return nil, err
	}

	return m.client, nil
}
