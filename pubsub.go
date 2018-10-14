package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"os"
)

const pubsubTopic = "views"

var pubsubClient *pubsub.Client = nil

func configurePubsub() (*pubsub.Client, error) {
	projectID := os.Getenv("GCLOUD_PROJECT")
	ctx := context.Background()

	var opts []option.ClientOption
	if file, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		opts = append(opts, option.WithCredentialsFile(file))
	}

	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}

	// Create the topic if it doesn't exist.
	if exists, err := client.Topic(pubsubTopic).Exists(ctx); err != nil {
		return nil, err
	} else if !exists {
		log.Info("creating pubsub topic ", pubsubTopic)
		if _, err := client.CreateTopic(ctx, pubsubTopic); err != nil {
			return nil, err
		}
	}
	return client, nil
}

var publishView = func(view View) error {
	if pubsubClient == nil {
		return errors.New("pubsub client is not set")
	}

	ctx := context.Background()

	data, err := MarshalJSON(view)
	if err != nil {
		return err
	}

	topic := pubsubClient.Topic(pubsubTopic)
	_, err = topic.Publish(ctx, &pubsub.Message{Data: data}).Get(ctx)
	return err
}
