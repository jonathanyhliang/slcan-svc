package main

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	GetMessageEndpoint    endpoint.Endpoint
	PostMessageEndpoint   endpoint.Endpoint
	PutMessageEndpoint    endpoint.Endpoint
	DeleteMessageEndpoint endpoint.Endpoint
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		GetMessageEndpoint:    MakeGetMessageEndpoint(s),
		PostMessageEndpoint:   MakePostMessageEndpoint(s),
		PutMessageEndpoint:    MakePutMessageEndpoint(s),
		DeleteMessageEndpoint: MakeDeleteMessageEndpoint(s),
	}
}

func MakeGetMessageEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getMessageRequest)
		m, e := s.GetMessage(ctx, req.ID)
		return getMessageResponse{Msg: m, Err: e}, nil
	}
}

func MakePostMessageEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postMessageRequest)
		e := s.PostMessage(ctx, req.Msg)
		return postMessageResponse{Err: e}, nil
	}
}

func MakePutMessageEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(putMessageRequest)
		e := s.PutMessage(ctx, req.ID, req.Msg)
		return putMessageResponse{Err: e}, nil
	}
}

func MakeDeleteMessageEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteMessageRequest)
		e := s.DeleteMessage(ctx, req.ID)
		return deleteMessageResponse{Err: e}, nil
	}
}

type getMessageRequest struct {
	ID string
}

type getMessageResponse struct {
	Msg Message `json:"message,omitempty"`
	Err error   `json:"err,omitempty"`
}

type postMessageRequest struct {
	Msg Message `json:"message,omitempty"`
}

type postMessageResponse struct {
	Err error `json:"err,omitempty"`
}

type putMessageRequest struct {
	ID  string
	Msg Message `json:"message,omitempty"`
}

type putMessageResponse struct {
	Err error `json:"err,omitempty"`
}

type deleteMessageRequest struct {
	ID string
}

type deleteMessageResponse struct {
	Err error `json:"err,omitempty"`
}
