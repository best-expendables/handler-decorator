package handler_decorator

import (
	"net/http"

	nrcontext "bitbucket.org/snapmartinc/newrelic-context"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

type HandlerCreatorFunc func(db *gorm.DB, redis *redis.Client) http.HandlerFunc

type handlerDecorator struct {
	db          *gorm.DB
	redisClient *redis.Client
}

type Option func(h *handlerDecorator) *handlerDecorator

func AddRedisToDecorator(r *redis.Client) func(h *handlerDecorator) *handlerDecorator {
	return func(h *handlerDecorator) *handlerDecorator {
		h.redisClient = r
		return h
	}
}
func AddDBToDecorator(db *gorm.DB) func(h *handlerDecorator) *handlerDecorator {
	return func(h *handlerDecorator) *handlerDecorator {
		h.db = db
		return h
	}
}

func NewHandlerDecorator(options ...Option) *handlerDecorator {
	baseHandler := &handlerDecorator{}
	for i := range options {
		options[i](baseHandler)
	}
	return baseHandler
}

func (b *handlerDecorator) NewRelicDecorate(handlerCreatorFunc func(db *gorm.DB, redis *redis.Client) http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if b.db != nil {
			b.db = nrcontext.SetTxnToGorm(request.Context(), b.db)
		}
		if b.redisClient != nil {
			b.redisClient = nrcontext.WrapRedisClient(request.Context(), b.redisClient)
		}
		handlerCreatorFunc(b.db, b.redisClient).ServeHTTP(writer, request)
	}
}
