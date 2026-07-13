package repository

import (
	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"go.uber.org/zap"
)

type FirebaseMessagingRepository struct {
	firebase *firebase.App
	userRepo *UserRepository
}

func NewFirebaseMessagingRepository(firebaseApp *firebase.App) *FirebaseMessagingRepository {
	return &FirebaseMessagingRepository{
		firebase: firebaseApp,
	}
}

func (r *FirebaseMessagingRepository) SendNotificationToUser(userID string, title string, body string, deeplink string) error {
	tokens := []string{}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Data: map[string]string{
			"title":    title,
			"body":     body,
			"deeplink": deeplink,
		},
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	client, err := r.firebase.Messaging(context.Background())
	if err != nil {
		return err
	}

	br, err := client.SendMulticast(context.Background(), message)
	if err != nil {
		return err
	}

	if br.FailureCount > 0 {
		for _, resp := range br.Responses {
			if !resp.Success {
				// Log the error for the failed token
				zap.S().Errorw("failed to send notification", "error", resp.Error)
			}
		}
	}

	return nil
}

func (r *FirebaseMessagingRepository) SendNotificationToUsers(userIDs []string, title string, body string, deeplink string) error {
	tokens := []string{}
	tokens := []string{}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Data: map[string]string{
			"title":    title,
			"body":     body,
			"deeplink": deeplink,
		},
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	client, err := r.firebase.Messaging(context.Background())
	if err != nil {
		return err
	}

	br, err := client.SendMulticast(context.Background(), message)
	if err != nil {
		return err
	}

	if br.FailureCount > 0 {
		for _, resp := range br.Responses {
			if !resp.Success {
				// Log the error for the failed token
				zap.S().Errorw("failed to send notification", "error", resp.Error)
			}
		}
	}

	return nil
}

func (r *FirebaseMessagingRepository) SendNotificationToTopic(topic string, title string, body string, deeplink string) error {
	message := &messaging.Message{
		Topic: topic,
		Data: map[string]string{
			"title":    title,
			"body":     body,
			"deeplink": deeplink,
		},
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	client, err := r.firebase.Messaging(context.Background())
	if err != nil {
		return err
	}

	response, err := client.Send(context.Background(), message)
	if err != nil {
		return err
	}
	if response == "" {
		zap.S().Errorw("failed to send notification to topic", "topic", topic)
		return err
	}

	zap.S().Infow("notification sent to topic", "topic", topic, "messageID", response)
	return nil
}
