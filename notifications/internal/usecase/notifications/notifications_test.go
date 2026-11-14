package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/entity"
	"go.uber.org/zap"
)

func TestSendOrderStatusChangedNotification(t *testing.T) {
	var got struct {
		UserID  int64  `json:"user_id"`
		OrderID int64  `json:"order_id"`
		Status  string `json:"status"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected callback request: method=%s content-type=%s", r.Method, r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode callback: %v", err)
		}
	}))
	defer server.Close()

	service := NewNotificationsServer(zap.NewNop(), strings.TrimPrefix(server.URL, "http://"))
	err := service.SendOrderStatusChangedNotification(context.Background(), entity.OrderStatusChangedNotification{
		UserID:  7,
		OrderID: 42,
		Status:  entity.OrderStatusPaid,
	})
	if err != nil {
		t.Fatalf("send notification: %v", err)
	}
	if got.UserID != 7 || got.OrderID != 42 || got.Status != "paid" {
		t.Fatalf("unexpected callback payload: %+v", got)
	}
}

func TestSendOrderStatusChangedNotificationCallbackFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	service := NewNotificationsServer(zap.NewNop(), server.URL)
	err := service.SendOrderStatusChangedNotification(context.Background(), entity.OrderStatusChangedNotification{
		UserID:  7,
		OrderID: 42,
		Status:  entity.OrderStatusPaid,
	})
	if err == nil {
		t.Fatal("expected callback error for HTTP 503")
	}
}
