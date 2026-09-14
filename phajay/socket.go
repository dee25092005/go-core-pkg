package phajay

import (
	"context"
	"fmt"
	"log"
	"time"

	socketio "github.com/maldikhan/go.socket.io/socket.io/v5/client"
)

type PaymentCallback func(txID string, rawData map[string]interface{})

type SocketClient struct {
	secretKey string
}

func NewSokcetClient(secretKey string) *SocketClient {
	return &SocketClient{
		secretKey: secretKey,
	}
}

func (s *SocketClient) StartListener(ctx context.Context, onComplete PaymentCallback) error {
	client, err := socketio.NewClient(
		socketio.WithRawURL("https://payment-gateway.phajay.co"),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize phajay socket client: %w", err)
	}

	client.On("connect", func() {
		log.Println("Connected to PhaJay Socket.io Gateway")

		eventName := fmt.Sprintf("join::%s", s.secretKey)
		client.On(eventName, func(data map[string]interface{}) {
			status, _ := data["status"].(string)
			txID, _ := data["transactionId"].(string)

			if status == "PAYMENT_COMPLETED" && txID != "" {
				onComplete(txID, data)
			}
		})
	})

	client.On("disconnect", func() {
		log.Println(" Disconnected from PhaJay Socket.io Gateway. Retrying...")
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := client.Connect(ctx); err != nil {
					time.Sleep(5 * time.Second)
				} else {
					return
				}
			}
		}
	}()

	return nil
}
