package usecase

import "context"

type Transactor interface {
	Do(ctx context.Context, fn func(txCtx context.Context) error) error
}
