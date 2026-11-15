package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
	"go.uber.org/zap"
)

func NewOrderStatusTelegramNotifier(logger *zap.Logger, cfg *config.Config) (*orderStatusTelegramNotifierImpl, error) {
	token := strings.TrimSpace(cfg.TelegramNotifier.Token)
	if token == "" {
		return nil, errors.New("TELEGRAM_TOKEN must not be empty")
	}
	if cfg.TelegramNotifier.ClientID == 0 {
		return nil, errors.New("TELEGRAM_CHAT_ID must not be zero")
	}

	bot, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, &http.Client{Timeout: cfg.Clients.HTTPTimeout})

	if err != nil {
		return nil, fmt.Errorf("create telegram bot api: %w", err)
	}

	return &orderStatusTelegramNotifierImpl{
		logger: logger,
		cfg:    cfg,
		bot:    bot,
	}, nil
}

type orderStatusTelegramNotifierImpl struct {
	logger *zap.Logger
	cfg    *config.Config
	bot    *tgbotapi.BotAPI
}

func (n *orderStatusTelegramNotifierImpl) NotifyOrderStatusChanged(
	ctx context.Context,
	userID,
	orderID int64,
	status string,
) error {
	n.logger.Info("Order status changed",
		zap.Int64("user_id", userID),
		zap.Int64("order_id", orderID),
		zap.String("status", status),
	)

	text := fmt.Sprintf(
		"Статус заказа изменён\n\norder_id: %d\nuser_id: %d\nstatus: %s",
		orderID,
		userID,
		status,
	)

	err := n.sendMessage(ctx, text)

	if err != nil {
		n.logger.Error("failed to send telegram message",
			zap.Int64("order_id", orderID),
			zap.Int64("chat_id", n.cfg.TelegramNotifier.ClientID),
			zap.Error(err),
		)
		return fmt.Errorf("send telegram message: %w", err)
	}

	n.logger.Info("telegram message sent",
		zap.Int64("order_id", orderID),
		zap.Int64("chat_id", n.cfg.TelegramNotifier.ClientID),
	)

	return nil
}

func (n *orderStatusTelegramNotifierImpl) sendMessage(ctx context.Context, text string) error {
	form := url.Values{
		"chat_id": {strconv.FormatInt(n.cfg.TelegramNotifier.ClientID, 10)},
		"text":    {text},
	}
	endpoint := fmt.Sprintf(tgbotapi.APIEndpoint, n.bot.Token, "sendMessage")
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()),
	)
	if err != nil {
		return errors.New("build Telegram sendMessage request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.bot.Client.Do(req)
	if err != nil {
		// В URL запроса есть токен. Возвращаем причину ошибки без URL.
		var requestErr *url.Error
		if errors.As(err, &requestErr) {
			err = requestErr.Err
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send Telegram message: HTTP %d", resp.StatusCode)
	}
	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode Telegram response: %w", err)
	}
	if !result.OK {
		return errors.New("send Telegram message: request rejected")
	}
	return nil
}
