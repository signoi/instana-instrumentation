package ginrouter

import (
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	krakendgin "github.com/devopsfaith/krakend/router/gin"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// New returns the handlerfactory by wrapping the base handler factory
// with instrumentation around the open census
func New(hf krakendgin.HandlerFactory) krakendgin.HandlerFactory {
	return func(cfg *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		return HandlerFunc(cfg, hf(cfg, p))
	}
}

// HandlerFunc returns gin.HandlerFunc by wrapping base handler func into
// handler struct
func HandlerFunc(cfg *config.EndpointConfig, next gin.HandlerFunc) gin.HandlerFunc {

	h := &handler{
		Base: next,
		name: cfg.Endpoint,
	}

	return h.HandlerFunc
}

type handler struct {
	name string
	Base gin.HandlerFunc
}

func (h *handler) HandlerFunc(c *gin.Context) {
	// Start the span here
	var span opentracing.Span
	// lets try to create a span from parent span
	if parentSpan := opentracing.SpanFromContext(c); parentSpan != nil {
		span = opentracing.StartSpan(h.name, opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = opentracing.StartSpan(h.name)
	}
	span.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCServerEnum))
	span.SetTag(string(ext.HTTPUrl), c.Request.URL.String())
	span.SetTag(string(ext.HTTPMethod), c.Request.Method)
	span.SetTag(string(ext.HTTPStatusCode), c.Writer.Status())
	defer span.Finish()
	ctx := opentracing.ContextWithSpan(c, span)
	c.Request = c.Request.WithContext(ctx)

	// Call the handler
	h.Base(c)

}
