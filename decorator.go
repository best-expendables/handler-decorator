package handler_decorator

import (
	"context"
	"net/http"

	nrcontext "bitbucket.org/snapmartinc/newrelic-context"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

type HandlerCreatorFunc func(decorator HandlerDecorator) http.HandlerFunc

type HandlerDecorator struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func (b *HandlerDecorator) cloneWithContext(ctx context.Context) HandlerDecorator {
	clone := HandlerDecorator{}
	if b.db != nil {
		clone.db = nrcontext.SetTxnToGorm(ctx, b.db)
	}
	if b.redisClient != nil {
		clone.redisClient = nrcontext.WrapRedisClient(ctx, b.redisClient)
	}
	return clone
}

func (b *HandlerDecorator) GetDB() *gorm.DB {
	return b.db
}

func (b *HandlerDecorator) GetRedisClient() *redis.Client {
	return b.redisClient
}

type Option func(h *HandlerDecorator) *HandlerDecorator

func AddRedisToDecorator(r *redis.Client) func(h *HandlerDecorator) *HandlerDecorator {
	return func(h *HandlerDecorator) *HandlerDecorator {
		h.redisClient = r
		return h
	}
}
func AddDBToDecorator(db *gorm.DB) func(h *HandlerDecorator) *HandlerDecorator {
	return func(h *HandlerDecorator) *HandlerDecorator {
		h.db = db
		return h
	}
}

func NewHandlerDecorator(options ...Option) *HandlerDecorator {
	baseHandler := &HandlerDecorator{}
	for i := range options {
		options[i](baseHandler)
	}
	return baseHandler
}

func (b *HandlerDecorator) NewRelicDecorate(handlerCreatorFunc func(decorator HandlerDecorator) http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handlerCreatorFunc(b.cloneWithContext(request.Context())).ServeHTTP(writer, request)
	}
}
