package slcansvc

import (
	"context"
	"errors"
)

var (
	ErrServiceInvalidID = errors.New("Service: invalid id")
)

const (
	CAN_ID_MIN = 0x0
	CAN_ID_MAX = 0x1fffffff
)

type IService interface {
	GetMessage(ctx context.Context, id int) (Message, error)
	PostMessage(ctx context.Context, m Message) error
	PutMessage(ctx context.Context, id int, m Message) error
	DeleteMessage(ctx context.Context, id int) error
	Reboot(ctx context.Context) error
	Unlock(ctx context.Context) error
}

type Service struct{}

func NewService() IService {
	return &Service{}
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
//	@Success		200	{array}	slcansvc.Message
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/slcan/{id} [get]
func (s *Service) GetMessage(ctx context.Context, id int) (Message, error) {
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
//	@Param			array	body	slcansvc.Message	false	"CAN Message"
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/slcan [post]
func (s *Service) PostMessage(ctx context.Context, m Message) error {
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
//	@Param			array	body	slcansvc.Message	false	"CAN Message"
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/slcan/{id} [put]
func (s *Service) PutMessage(ctx context.Context, id int, m Message) error {
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
func (s *Service) DeleteMessage(ctx context.Context, id int) error {
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
func (s *Service) Reboot(ctx context.Context) error {
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
func (s *Service) Unlock(ctx context.Context) error {
	return nil
}
