package slcansvc

import (
	"context"
	"net/url"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type Endpoints struct {
	GetMessageEndpoint    endpoint.Endpoint
	PostMessageEndpoint   endpoint.Endpoint
	PutMessageEndpoint    endpoint.Endpoint
	DeleteMessageEndpoint endpoint.Endpoint
	RebootEndpoint        endpoint.Endpoint
	UnlockEndpoint        endpoint.Endpoint
}

func MakeServerEndpoints(s IService) Endpoints {
	return Endpoints{
		GetMessageEndpoint:    MakeGetMessageEndpoint(s),
		PostMessageEndpoint:   MakePostMessageEndpoint(s),
		PutMessageEndpoint:    MakePutMessageEndpoint(s),
		DeleteMessageEndpoint: MakeDeleteMessageEndpoint(s),
		RebootEndpoint:        MakeRebootEndpoint(s),
		UnlockEndpoint:        MakeUnlockEndpoint(s),
	}
}

// MakeClientEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the remote instance, via a transport/http.Client.
func MakeClientEndpoints(instance string) (Endpoints, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	tgt, err := url.Parse(instance)
	if err != nil {
		return Endpoints{}, err
	}
	tgt.Path = ""

	options := []httptransport.ClientOption{}

	// Note that the request encoders need to modify the request URL, changing
	// the path. That's fine: we simply need to provide specific encoders for
	// each endpoint.

	return Endpoints{
		GetMessageEndpoint: httptransport.NewClient("GET", tgt,
			EncodeGetMessageRequest, DecodeGetMessageResponse, options...).Endpoint(),
		PostMessageEndpoint: httptransport.NewClient("POST", tgt,
			EncodePostMessageRequest, DecodePostMessageResponse, options...).Endpoint(),
		PutMessageEndpoint: httptransport.NewClient("PUT", tgt,
			EncodePutMessageRequest, DecodePutMessageResponse, options...).Endpoint(),
		DeleteMessageEndpoint: httptransport.NewClient("DELETE", tgt,
			EncodeDeleteMessageRequest, DecodeDeleteMessageResponse, options...).Endpoint(),
		RebootEndpoint: httptransport.NewClient("POST", tgt,
			EncodeRebootRequest, DecodeRebootResponse, options...).Endpoint(),
		UnlockEndpoint: httptransport.NewClient("POST", tgt,
			EncodeUnlockRequest, DecodeUnlockResponse, options...).Endpoint(),
	}, nil
}

// struct Endpoints implements interface IService. Primarily useful in a client.
func (e Endpoints) GetMessage(ctx context.Context, id int) (Message, error) {
	request := getMessageRequest{ID: id}
	response, err := e.GetMessageEndpoint(ctx, request)
	if err != nil {
		return Message{}, err
	}
	resp := response.(getMessageResponse)
	return resp.Msg, resp.Err
}

func (e Endpoints) PostMessage(ctx context.Context, m Message) error {
	request := postMessageRequest{Msg: m}
	response, err := e.PostMessageEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(postMessageResponse)
	return resp.Err
}

func (e Endpoints) PutMessage(ctx context.Context, id int, m Message) error {
	request := putMessageRequest{ID: id, Msg: m}
	response, err := e.PutMessageEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(putMessageResponse)
	return resp.Err
}

func (e Endpoints) DeleteMessage(ctx context.Context, id int) error {
	request := deleteMessageRequest{ID: id}
	response, err := e.DeleteMessageEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(deleteMessageResponse)
	return resp.Err
}

func (e Endpoints) Reboot(ctx context.Context) error {
	response, err := e.DeleteMessageEndpoint(ctx, rebootRequest{})
	if err != nil {
		return err
	}
	resp := response.(rebootResponse)
	return resp.Err
}

func (e Endpoints) Unlock(ctx context.Context) error {
	response, err := e.DeleteMessageEndpoint(ctx, unlockRequest{})
	if err != nil {
		return err
	}
	resp := response.(unlockResponse)
	return resp.Err
}

func MakeGetMessageEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getMessageRequest)
		m, e := s.GetMessage(ctx, req.ID)
		return getMessageResponse{Msg: m, Err: e}, nil
	}
}

func MakePostMessageEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postMessageRequest)
		e := s.PostMessage(ctx, req.Msg)
		return postMessageResponse{Err: e}, nil
	}
}

func MakePutMessageEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(putMessageRequest)
		e := s.PutMessage(ctx, req.ID, req.Msg)
		return putMessageResponse{Err: e}, nil
	}
}

func MakeDeleteMessageEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteMessageRequest)
		e := s.DeleteMessage(ctx, req.ID)
		return deleteMessageResponse{Err: e}, nil
	}
}

func MakeRebootEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(rebootRequest)
		e := s.Reboot(ctx)
		return rebootResponse{Err: e}, nil
	}
}

func MakeUnlockEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(unlockRequest)
		e := s.Unlock(ctx)
		return unlockResponse{Err: e}, nil
	}
}

type getMessageRequest struct {
	ID int
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
	ID  int
	Msg Message `json:"message,omitempty"`
}

type putMessageResponse struct {
	Err error `json:"err,omitempty"`
}

type deleteMessageRequest struct {
	ID int
}

type deleteMessageResponse struct {
	Err error `json:"err,omitempty"`
}

type rebootRequest struct{}

type rebootResponse struct {
	Err error `json:"err,omitempty"`
}

type unlockRequest struct{}

type unlockResponse struct {
	Err error `json:"err,omitempty"`
}
