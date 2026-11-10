package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/sqlc"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/transactor"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type outboxRepository struct {
	queries *sqlc.Queries
}

func NewOutboxRepository(pool *pgxpool.Pool) *outboxRepository {
	return &outboxRepository{queries: sqlc.New(pool)}
}

func (o *outboxRepository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return sqlc.New(tx)
	}

	return o.queries
}

func (o *outboxRepository) SendMessage(
	ctx context.Context,
	idempotencyKey string,
	kind Kind,
	message []byte,
) error {
	err := o.getQueries(ctx).SendOutboxMessage(ctx, sqlc.SendOutboxMessageParams{
		IdempotencyKey: idempotencyKey,
		Data:           message,
		Kind:           int32(kind),
	})

	if err != nil {
		return fmt.Errorf("send outbox message: %w", err)
	}

	return nil
}

// status == CREATED || (status == IN_PROGRESS && time.Now() - updated_at > TTL)

func (o *outboxRepository) GetMessages(
	ctx context.Context,
	batchSize int,
	inProgressTTL time.Duration,
) ([]Data, error) {
	rows, err := o.getQueries(ctx).GetOutboxMessages(ctx, sqlc.GetOutboxMessagesParams{
		InProgressTtl: pgtype.Interval{
			Microseconds: inProgressTTL.Microseconds(),
			Valid:        true,
		},
		BatchSize: int32(batchSize),
	})

	if err != nil {
		return nil, fmt.Errorf("get outbox messages: %w", err)
	}

	result := make([]Data, 0, len(rows))
	for _, row := range rows {
		result = append(result, Data{
			IdempotencyKey: row.IdempotencyKey,
			Data:           row.Data,
			Kind:           Kind(row.Kind),
		})
	}

	return result, nil
}

func (o *outboxRepository) MarkAsProcessed(ctx context.Context, idempotencyKeys []string) error {
	if len(idempotencyKeys) == 0 {
		return nil
	}

	err := o.getQueries(ctx).MarkOutboxMessagesAsProcessed(ctx, idempotencyKeys)
	if err != nil {
		return fmt.Errorf("mark outbox messages as processed: %w", err)
	}

	return nil
}

func (o *outboxRepository) MarkAsRetryable(ctx context.Context, idempotencyKeys []string) error {
	if len(idempotencyKeys) == 0 {
		return nil
	}

	err := o.getQueries(ctx).MarkOutboxMessagesAsRetryable(ctx, idempotencyKeys)
	if err != nil {
		return fmt.Errorf("mark outbox messages as retryable: %w", err)
	}

	return nil
}
