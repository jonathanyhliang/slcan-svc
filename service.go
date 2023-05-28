package main

import (
	"context"
	"errors"
	"strconv"
)

var (
	ErrServiceInvalidID = errors.New("Service: invalid ID")
)

type Service interface {
	GetMessage(ctx context.Context, id string) (Message, error)
	PostMessage(ctx context.Context, m Message) error
	PutMessage(ctx context.Context, id string, m Message) error
	DeleteMessage(ctx context.Context, id string) error
	Reboot(ctx context.Context) error
}

type SlcanService struct{}

func NewSlcanService() Service {
	return &SlcanService{}
}

func (s *SlcanService) GetMessage(ctx context.Context, id string) (Message, error) {
	i, err := strconv.Atoi(id)
	if err != nil {
		return Message{}, ErrServiceInvalidID
	}
	return db.GetData(uint32(i))
}

func (s *SlcanService) PostMessage(ctx context.Context, m Message) error {
	return db.PostData(m)
}

func (s *SlcanService) PutMessage(ctx context.Context, id string, m Message) error {
	i, err := strconv.Atoi(id)
	if err != nil {
		return ErrServiceInvalidID
	}
	return db.PutData(uint32(i), m)
}

func (s *SlcanService) DeleteMessage(ctx context.Context, id string) error {
	i, err := strconv.Atoi(id)
	if err != nil {
		return ErrServiceInvalidID
	}
	return db.DeleteData(uint32(i))
}

func (s *SlcanService) Reboot(ctx context.Context) error {
	return nil
}
