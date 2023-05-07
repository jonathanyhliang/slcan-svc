package main

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) GetMessage(ctx context.Context, id string) (m Message, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetMessage", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetMessage(ctx, id)
}

func (mw loggingMiddleware) PostMessage(ctx context.Context, m Message) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "PostMessage", "id", m.ID, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.PostMessage(ctx, m)
}

func (mw loggingMiddleware) PutMessage(ctx context.Context, id string, m Message) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "PutMessage", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.PutMessage(ctx, id, m)
}

func (mw loggingMiddleware) DeleteMessage(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "DeleteMessage", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.DeleteMessage(ctx, id)
}

func BackendMiddleware(backend Backend) Middleware {
	return func(next Service) Service {
		return &backendMiddleware{
			next:    next,
			backend: backend,
		}
	}
}

type backendMiddleware struct {
	next    Service
	backend Backend
}

func (mw backendMiddleware) GetMessage(ctx context.Context, id string) (m Message, err error) {
	d, e := mw.next.GetMessage(ctx, id)
	if e != nil {
		e = mw.backend.GetMessage(id)
	}
	return d, e
}

func (mw backendMiddleware) PostMessage(ctx context.Context, m Message) (err error) {
	e := mw.next.PostMessage(ctx, m)
	if e == nil {
		e = mw.backend.PostMessage(m)
	}
	return e
}

func (mw backendMiddleware) PutMessage(ctx context.Context, id string, m Message) (err error) {
	e := mw.next.PutMessage(ctx, id, m)
	if e == nil {
		e = mw.backend.PostMessage(m)
	}
	return e
}

func (mw backendMiddleware) DeleteMessage(ctx context.Context, id string) (err error) {
	return mw.next.DeleteMessage(ctx, id)
}
