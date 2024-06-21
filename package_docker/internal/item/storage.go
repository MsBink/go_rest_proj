package item

import "context"

type Storage interface {
	Create(ctx context.Context, item Item) (string, error)
	FindOne(ctx context.Context, id string) (Item, error)
	FindAllByUser(ctx context.Context, userID string) ([]Item, error)
	FindAll(ctx context.Context) ([]Item, error)
	Update(ctx context.Context, item Item) error
	Delete(ctx context.Context, id string) error
}
