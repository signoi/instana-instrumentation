package adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	// ContextKey is a key for context value of span
	ContextKey string = "opencensus-span"
)

// Trace defines the middleware for gin that wraps around the handler
// and creates span
func Trace() gin.HandlerFunc {

	return func(c *gin.Context) {

		tracer := opentracing.GlobalTracer()
		wireContext, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		var span opentracing.Span

		if err != nil {
			span = tracer.StartSpan(
				c.HandlerName(),
			)
		} else {
			span = tracer.StartSpan(
				c.HandlerName(),
				ext.RPCServerOption(wireContext),
			)
		}

		span.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
		span.SetTag(string(ext.PeerHostname), c.Request.Host)
		span.SetTag(string(ext.HTTPUrl), c.Request.URL.Path)
		span.SetTag(string(ext.HTTPMethod), c.Request.Method)

		// Wrap the request with span context
		c.Request = wrapRequestWithSpanContext(c.Request, span)
		c.Set(ContextKey, span)
		// call the next
		// set the response headers
		tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Writer.Header()))
		c.Next()
		span.SetTag(string(ext.HTTPStatusCode), c.Writer.Status())

		span.Finish()

	}
}

func wrapRequestWithSpanContext(req *http.Request, span opentracing.Span) *http.Request {
	ctx := opentracing.ContextWithSpan(req.Context(), span)
	req = req.WithContext(ctx)
	return req
}
