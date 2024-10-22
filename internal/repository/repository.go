package repository

import (
	"context"
	"github.com/seshoo/bookFinder/internal/domain"
)

type Topics interface {
	GetList(context.Context, int, int) ([]domain.Topic, error)
	GetById(context.Context, string) (*domain.Topic, error)
}

type Repositories interface {
	Topic() Topics
}
