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

func MakeGetMessageEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetMessageRequest)
		m, e := s.GetMessage(ctx, req.ID)
		return GetMessageResponse{Msg: m, Err: e}, nil
	}
}

func MakePostMessageEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(PostMessageRequest)
		e := s.PostMessage(ctx, req.Msg)
		return PostMessageResponse{Err: e}, nil
	}
}

func MakePutMessageEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(PutMessageRequest)
		e := s.PutMessage(ctx, req.ID, req.Msg)
		return PutMessageResponse{Err: e}, nil
	}
}

func MakeDeleteMessageEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DeleteMessageRequest)
		e := s.DeleteMessage(ctx, req.ID)
		return DeleteMessageResponse{Err: e}, nil
	}
}

func MakeRebootEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(RebootRequest)
		e := s.Reboot(ctx)
		return RebootResponse{Err: e}, nil
	}
}

func MakeUnlockEndpoint(s IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(UnlockRequest)
		e := s.Unlock(ctx)
		return UnlockResponse{Err: e}, nil
	}
}

type GetMessageRequest struct {
	ID int
}

type GetMessageResponse struct {
	Msg Message `json:"message,omitempty"`
	Err error   `json:"err,omitempty"`
}

type PostMessageRequest struct {
	Msg Message `json:"message,omitempty"`
}

type PostMessageResponse struct {
	Err error `json:"err,omitempty"`
}

type PutMessageRequest struct {
	ID  int
	Msg Message `json:"message,omitempty"`
}

type PutMessageResponse struct {
	Err error `json:"err,omitempty"`
}

type DeleteMessageRequest struct {
	ID int
}

type DeleteMessageResponse struct {
	Err error `json:"err,omitempty"`
}

type RebootRequest struct{}

type RebootResponse struct {
	Err error `json:"err,omitempty"`
}

type UnlockRequest struct{}

type UnlockResponse struct {
	Err error `json:"err,omitempty"`
}
