package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
)

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrTransportBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	// GET     /slcan/:id	retrieves the given SLCAN message by id
	// POST    /slcan/		adds another SLCAN message
	// PUT     /slcan/:id	post updated SLCAN message
	// DELETE  /slcan/:id 	remove the given SLCAN message
	r.Methods("GET").Path("/slcan/{id}").Handler(httptransport.NewServer(
		e.GetMessageEndpoint,
		decodeGetMessageRequest,
		encodeResponse,
		options...,
	))
	r.Methods("POST").Path("/slcan/").Handler(httptransport.NewServer(
		e.PostMessageEndpoint,
		decodePostMessageRequest,
		encodeResponse,
		options...,
	))
	r.Methods("PUT").Path("/slcan/{id}").Handler(httptransport.NewServer(
		e.PutMessageEndpoint,
		decodePutMessageRequest,
		encodeResponse,
		options...,
	))
	r.Methods("DELETE").Path("/slcan/{id}").Handler(httptransport.NewServer(
		e.DeleteMessageEndpoint,
		decodeDeleteMessageRequest,
		encodeResponse,
		options...,
	))
	r.Methods("POST").Path("/slcan/reboot").Handler(httptransport.NewServer(
		e.RebootEndpoint,
		decodeRebootRequest,
		encodeResponse,
		options...,
	))
	return r
}

func decodeGetMessageRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrTransportBadRouting
	}
	return getMessageRequest{ID: id}, nil
}

func decodePostMessageRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req postMessageRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Msg); e != nil {
		return nil, e
	}
	return req, nil
}

func decodePutMessageRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrTransportBadRouting
	}
	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		return nil, err
	}
	return putMessageRequest{ID: id, Msg: msg}, nil
}

func decodeDeleteMessageRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrTransportBadRouting
	}
	return deleteMessageRequest{ID: id}, nil
}

func decodeRebootRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	return rebootRequest{}, nil
}

type errorer interface {
	error() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

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
	case ErrDatabaseAlreadyExists, ErrServiceInvalidID:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
