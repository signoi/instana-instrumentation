package proxy

import (
	"context"
	"errors"
	"log"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/opentracing/opentracing-go"
)

const (
	// ContextKey is a key for context value of span
	ContextKey         string = "opencensus-span"
	errCtxCancedledMsg        = "context canceled"
)

func SpanFromContext(ctx context.Context) (opentracing.Span, error) {
	span, ok := ctx.Value(ContextKey).(opentracing.Span)
	if !ok {
		return nil, errors.New("Span not present in context")
	}
	return span, nil
}

func Middleware(name string) proxy.Middleware {
	return func(next ...proxy.Proxy) proxy.Proxy {
		if len(next) > 1 {
			panic(proxy.ErrTooManyProxies)
		}

		if len(next) < 1 {
			panic(proxy.ErrNotEnoughProxies)
		}

		return func(ctx context.Context, req *proxy.Request) (*proxy.Response, error) {

			// get the span
			var span opentracing.Span
			parentSpan, err := SpanFromContext(ctx)
			if err != nil {
				log.Println(err)
				span = opentracing.StartSpan(name)

			} else {
				span = opentracing.StartSpan(name, opentracing.ChildOf(parentSpan.Context()))
			}
			ctx = opentracing.ContextWithSpan(ctx, span)
			resp, err := next[0](ctx, req)

			if err != nil {
				if err.Error() == errCtxCancedledMsg {
					span.SetTag(string(ext.Error), "context canceled")
				} else {
					span.SetTag(string(ext.Error), err.Error())
				}
			}
			span.Finish()

			return resp, err

		}
	}
}
func ProxyFactory(pf proxy.Factory) proxy.FactoryFunc {
	return func(cfg *config.EndpointConfig) (proxy.Proxy, error) {
		next, err := pf.New(cfg)
		if err != nil {
			return next, err
		}
		return Middleware("pipe-" + cfg.Endpoint)(next), nil
	}
}

func BackendFactory(bf proxy.BackendFactory) proxy.BackendFactory {
	return func(cfg *config.Backend) proxy.Proxy {
		return Middleware("backend-" + cfg.URLPattern)(bf(cfg))
	}
}
