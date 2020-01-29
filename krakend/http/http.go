package http

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"

	transport "github.com/devopsfaith/krakend/transport/http/client"
)

// HTTPRequestExecutor ...
func HTTPRequestExecutor(clientFactory transport.HTTPClientFactory) transport.HTTPRequestExecutor {
	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		client := clientFactory(ctx)
		span := opentracing.SpanFromContext(ctx)

		headersCarrier := opentracing.HTTPHeadersCarrier(req.Header)
		if err := opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			headersCarrier,
		); err != nil {
			return nil, err
		}

		return client.Do(req.WithContext(opentracing.ContextWithSpan(ctx, span)))

	}
}
