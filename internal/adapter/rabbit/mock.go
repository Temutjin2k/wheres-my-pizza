package rabbit

import (
	"context"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
)

type FakeProd struct {
}

func NewProducerNotify() *FakeProd {
	return &FakeProd{}
}

func (FakeProd) StatusUpdate(ctx context.Context, req *models.StatusUpdate) error {
	return nil
}
