package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/twmb/franz-go/pkg/kgo"
)

func (m *KafkaClientManager) ProduceWebhookPayload(client *kgo.Client, payload []WebhookPayload, receiverName string) error {
	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		log.Println("Invalid Json Payload, err: ", err)
		return err
	}

	topic := os.Getenv("KAFKA_TOPIC")
	record := &kgo.Record {
		Topic: topic,
		Key: []byte(receiverName),
		Value: jsonPayload,
		Headers: []kgo.RecordHeader{
			{
				Key: "receiver",
				Value: []byte(receiverName),
			},
			{
				Key: "timestamp",
				Value: []byte(time.Now().Format(time.RFC3339)),
			},
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	client.Produce(ctx, record, func (r *kgo.Record, err error){
		defer wg.Done()
		if err != nil {
			log.Printf("Error Producing message, err: %v\n", err)
		}else {
			log.Printf("Message Produced to topic: %v\n", r.Topic)
		}
		
	})
	wg.Wait()

	return nil
}