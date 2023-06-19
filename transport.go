package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

var (
	ErrServiceInvalidID = errors.New("Service: invalid ID")

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
		return nil, ErrServiceInvalidID
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
		return nil, ErrServiceInvalidID
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
		return nil, ErrServiceInvalidID
	}
	return deleteMessageRequest{ID: i}, nil
}

func DecodeRebootRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	return rebootRequest{}, nil
}

func DecodeUnlockRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	return unlockRequest{}, nil
}

type errorer interface {
	error() error
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
