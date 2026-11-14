package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/entity"
	"go.uber.org/zap"
)

type NotificationsService struct {
	logger       *zap.Logger
	callbackAddr string
	client       *http.Client
}

func NewNotificationsServer(logger *zap.Logger, callbackAddr string) *NotificationsService {
	callbackAddr = strings.TrimSpace(callbackAddr)
	if callbackAddr == "" {
		return nil
	}
	if !strings.Contains(callbackAddr, "://") {
		callbackAddr = "http://" + callbackAddr
	}

	return &NotificationsService{
		logger:       logger,
		callbackAddr: callbackAddr,
		client:       &http.Client{Timeout: 3 * time.Second},
	}
}
func (n *NotificationsService) SendOrderStatusChangedNotification(ctx context.Context, body entity.OrderStatusChangedNotification) error {
	payload := struct {
		UserID  int64  `json:"user_id"`
		OrderID int64  `json:"order_id"`
		Status  string `json:"status"`
	}{
		UserID:  body.UserID,
		OrderID: body.OrderID,
		Status:  string(body.Status),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.callbackAddr, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create callback request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("send callback: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("callback returned %s", resp.Status)
	}

	n.logger.Info("Order status changed",
		zap.String("user_id", strconv.FormatInt(body.UserID, 10)),
		zap.String("order_id", strconv.FormatInt(body.OrderID, 10)),
		zap.String("status", string(body.Status)))
	return nil
}
