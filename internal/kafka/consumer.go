package kafka

import (
	"context"
	"log"
)

func (m *KafkaClientManager) ConsumeWebhookPayload() {
	ctx := context.Background()

	for {
		fetches := m.client.PollFetches(ctx)

		if errs := fetches.Errors(); len(errs) > 0 {
			log.Println("Error occured while consuming webhook payloads", errs)
		}

		iter := fetches.RecordIter()

		for !iter.Done() {
			record := iter.Next()
			StoreRecordForInterestedUsers(record)
		}
	}
}
