package service

import (
	"context"
	"errors"
)

type Tracking struct {
}

func NewTracking() *Tracking {
	return &Tracking{}
}

func (s *Tracking) Start(ctx context.Context) error {
	return errors.ErrUnsupported
}
