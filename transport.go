package slcansvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

var (
	// ErrTransportBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrTransportBadRouting = errors.New("Transport: bad routing")
)

func MakeHTTPHandler(s IService, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("GET").Path("/slcan/{id}").Handler(httptransport.NewServer(
		e.GetMessageEndpoint,
		DecodeGetMessageRequest,
		EncodeResponse,
		options...,
	))
	r.Methods("POST").Path("/slcan").Handler(httptransport.NewServer(
		e.PostMessageEndpoint,
		DecodePostMessageRequest,
		EncodeResponse,
		options...,
	))
	r.Methods("PUT").Path("/slcan/{id}").Handler(httptransport.NewServer(
		e.PutMessageEndpoint,
		DecodePutMessageRequest,
		EncodeResponse,
		options...,
	))
	r.Methods("DELETE").Path("/slcan/{id}").Handler(httptransport.NewServer(
		e.DeleteMessageEndpoint,
		DecodeDeleteMessageRequest,
		EncodeResponse,
		options...,
	))
	r.Methods("POST").Path("/slcan/reboot").Handler(httptransport.NewServer(
		e.RebootEndpoint,
		DecodeRebootRequest,
		EncodeResponse,
		options...,
	))
	r.Methods("POST").Path("/slcan/unlock").Handler(httptransport.NewServer(
		e.UnlockEndpoint,
		DecodeUnlockRequest,
		EncodeResponse,
		options...,
	))
	r.PathPrefix("/slcan/docs").Handler(httpSwagger.WrapHandler)

	return r
}

func DecodeGetMessageRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrTransportBadRouting
	}
	i, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrTransportBadRouting
	}
	return getMessageRequest{ID: i}, nil
}

func DecodePostMessageRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req postMessageRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Msg); e != nil {
		return nil, e
	}
	return req, nil
}

func DecodePutMessageRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrTransportBadRouting
	}
	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		return nil, err
	}
	i, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrTransportBadRouting
	}
	return putMessageRequest{ID: i, Msg: msg}, nil
}

func DecodeDeleteMessageRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrTransportBadRouting
	}
	i, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrTransportBadRouting
	}
	return deleteMessageRequest{ID: i}, nil
}

func DecodeRebootRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	return rebootRequest{}, nil
}

func DecodeUnlockRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	return unlockRequest{}, nil
}

func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func EncodeGetMessageRequest(ctx context.Context, req *http.Request, request interface{}) error {
	// r.Methods("GET").Path("/slcan/{id}")
	r := request.(getMessageRequest)
	id := strconv.Itoa(r.ID)
	req.URL.Path = "/slcan/" + id
	return encodeRequest(ctx, req, nil)
}

func EncodePostMessageRequest(ctx context.Context, req *http.Request, request interface{}) error {
	// r.Methods("POST").Path("/slcan")
	req.URL.Path = "/slcan"
	return encodeRequest(ctx, req, request)
}

func EncodePutMessageRequest(ctx context.Context, req *http.Request, request interface{}) error {
	// r.Methods("PUT").Path("/slcan/{id}")
	r := request.(putMessageRequest)
	id := strconv.Itoa(r.ID)
	req.URL.Path = "/slcan/" + id
	return encodeRequest(ctx, req, request)
}

func EncodeDeleteMessageRequest(ctx context.Context, req *http.Request, request interface{}) error {
	// r.Methods("DELETE").Path("/slcan/{id}")
	r := request.(deleteMessageRequest)
	id := strconv.Itoa(r.ID)
	req.URL.Path = "/slcan/" + id
	return encodeRequest(ctx, req, request)
}

func EncodeRebootRequest(ctx context.Context, req *http.Request, request interface{}) error {
	// r.Methods("POST").Path("/slcan/reboot")
	req.URL.Path = "/slcan/reboot"
	return encodeRequest(ctx, req, request)
}

func EncodeUnlockRequest(ctx context.Context, req *http.Request, request interface{}) error {
	// r.Methods("POST").Path("/slcan/unlock")
	req.URL.Path = "/slcan/unlock"
	return encodeRequest(ctx, req, request)
}

func DecodeGetMessageResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp getMessageResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func DecodePostMessageResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp postMessageResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func DecodePutMessageResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp putMessageResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func DecodeDeleteMessageResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp deleteMessageResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func DecodeRebootResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp rebootResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func DecodeUnlockResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp unlockResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

type errorer interface {
	error() error
}

// encodeRequest likewise JSON-encodes the request to the HTTP request body.
// Don't use it directly as a transport/http.Client EncodeRequestFunc:
// Endpoints require mutating the HTTP method and request path.
func encodeRequest(_ context.Context, req *http.Request, request interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrDatabaseNotFound:
		return http.StatusNotFound
	case ErrDatabaseAlreadyExists, ErrTransportBadRouting:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
