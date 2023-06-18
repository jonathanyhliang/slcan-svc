package main

import (
	"context"
)

const (
	CAN_ID_MIN = 0x0
	CAN_ID_MAX = 0x1fffffff
)

type Service interface {
	GetMessage(ctx context.Context, id int) (Message, error)
	PostMessage(ctx context.Context, m Message) error
	PutMessage(ctx context.Context, id int, m Message) error
	DeleteMessage(ctx context.Context, id int) error
	Reboot(ctx context.Context) error
	Unlock(ctx context.Context) error
}

type SlcanService struct{}

func NewSlcanService() Service {
	return &SlcanService{}
}

// GetMessage godoc
//
//	@Summary	Retrieve CAN message
//	@Schemes
//	@Description	Retrieve CAN message by specifying CAN ID
//	@Tags			SLCAN
//	@Param			int	path	int	true	"CAN ID"	minimum(0)	maximum(536870911)
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	main.Message
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/slcan/{id} [get]
func (s *SlcanService) GetMessage(ctx context.Context, id int) (Message, error) {
	if id < CAN_ID_MIN || id > CAN_ID_MAX {
		return Message{}, ErrServiceInvalidID
	}
	return db.GetData(uint32(id))
}

// PostMessage godoc
//
//	@Summary	Add new CAN message
//	@Schemes
//	@Description	Add new CAN message by specifying CAN ID and data
//	@Tags			SLCAN
//	@Param			array	body	main.Message	false	"CAN Message"
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/slcan [post]
func (s *SlcanService) PostMessage(ctx context.Context, m Message) error {
	if m.ID > CAN_ID_MAX {
		return ErrServiceInvalidID
	}
	return db.PostData(m)
}

// PutMessage godoc
//
//	@Summary	Update existing CAN message
//	@Schemes
//	@Description	Update existing CAN message by specifying CAN ID and data
//	@Tags			SLCAN
//	@Param			array	body	main.Message	false	"CAN Message"
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/slcan/{id} [put]
func (s *SlcanService) PutMessage(ctx context.Context, id int, m Message) error {
	if id < CAN_ID_MIN || id > CAN_ID_MAX {
		return ErrServiceInvalidID
	}
	return db.PutData(uint32(id), m)
}

// DeleteMessage godoc
//
//	@Summary	Remove CAN message
//	@Schemes
//	@Description	Remove CAN message by specifying CAN ID
//	@Tags			SLCAN
//	@Param			int	path	int	true	"CAN ID"	minimum(0)	maximum(536870911)
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/slcan/{id} [delete]
func (s *SlcanService) DeleteMessage(ctx context.Context, id int) error {
	if id < CAN_ID_MIN || id > CAN_ID_MAX {
		return ErrServiceInvalidID
	}
	return db.DeleteData(uint32(id))
}

// Reboot godoc
//
//	@Summary	Reboot SLCAN device
//	@Schemes
//	@Description	Reboot SLCAN device for firmware update
//	@Tags			SLCAN
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/slcan/reboot [post]
func (s *SlcanService) Reboot(ctx context.Context) error {
	return nil
}

// Unlock godoc
//
//	@Summary	Unlock serial backend
//	@Schemes
//	@Description	Unlock serial backend from the success of firmware update
//	@Tags			SLCAN
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/slcan/unlock [post]
func (s *SlcanService) Unlock(ctx context.Context) error {
	return nil
}
