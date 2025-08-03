package services

import (
	"context"
	"errors"
)

type KitchenService struct {
}

func NewKitchen() *KitchenService {
	return &KitchenService{}
}

func (s *KitchenService) Start(ctx context.Context) error {
	return errors.ErrUnsupported
}
