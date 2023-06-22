package slcansvc

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(IService) IService

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next IService) IService {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   IService
	logger log.Logger
}

func (mw loggingMiddleware) GetMessage(ctx context.Context, id int) (m Message, err error) {
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

func (mw loggingMiddleware) PutMessage(ctx context.Context, id int, m Message) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "PutMessage", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.PutMessage(ctx, id, m)
}

func (mw loggingMiddleware) DeleteMessage(ctx context.Context, id int) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "DeleteMessage", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.DeleteMessage(ctx, id)
}

func (mw loggingMiddleware) Reboot(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "Reboot", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Reboot(ctx)
}

func (mw loggingMiddleware) Unlock(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "Unlock", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Unlock(ctx)
}

func BackendMiddleware(backend IBackend) Middleware {
	return func(next IService) IService {
		return &backendMiddleware{
			next:    next,
			backend: backend,
		}
	}
}

type backendMiddleware struct {
	next    IService
	backend IBackend
}

func (mw backendMiddleware) GetMessage(ctx context.Context, id int) (m Message, err error) {
	d, e := mw.next.GetMessage(ctx, id)
	// TODO: request from backend when message not existed
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

func (mw backendMiddleware) PutMessage(ctx context.Context, id int, m Message) (err error) {
	e := mw.next.PutMessage(ctx, id, m)
	if e == nil {
		e = mw.backend.PostMessage(m)
	}
	return e
}

func (mw backendMiddleware) DeleteMessage(ctx context.Context, id int) (err error) {
	return mw.next.DeleteMessage(ctx, id)
}

func (mw backendMiddleware) Reboot(ctx context.Context) (err error) {
	e := mw.next.Reboot(ctx)
	if e == nil {
		e = mw.backend.Reboot()
	}
	return e
}

func (mw backendMiddleware) Unlock(ctx context.Context) (err error) {
	e := mw.next.Unlock(ctx)
	if e == nil {
		e = mw.backend.Unlock()
	}
	return e
}
