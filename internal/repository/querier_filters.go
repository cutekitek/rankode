package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type DynamicQuerier interface {
	Querier
	WithTx(tx pgx.Tx) *Queries
	GetTaskListByFilter(ctx context.Context, filter TaskListFilter) ([]Task, error)
}