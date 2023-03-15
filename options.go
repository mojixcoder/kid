package kid

import (
	htmlrenderer "github.com/mojixcoder/kid/html_renderer"
	"github.com/mojixcoder/kid/serializer"
)

type (
	// Option is the interface for customizing Kid.
	Option interface {
		apply(*Kid)
	}

	optionImpl func(*Kid)
)

// applyOption implements the Option interface.
func (f optionImpl) apply(k *Kid) {
	f(k)
}

// WithDebug configures Kid's debug option.
func WithDebug(debug bool) Option {
	return optionImpl(func(k *Kid) {
		k.debug = debug
	})
}

// WithHTMLRenderer configures Kid's HTML renderer.
func WithHTMLRenderer(renderer htmlrenderer.HTMLRenderer) Option {
	panicIfNil(renderer, "renderer cannot be nil")

	return optionImpl(func(k *Kid) {
		k.htmlRenderer = renderer
	})
}

// WithXMLSerializer configures Kid's XML serializer.
func WithXMLSerializer(serializer serializer.Serializer) Option {
	panicIfNil(serializer, "xml serializer cannot be nil")

	return optionImpl(func(k *Kid) {
		k.xmlSerializer = serializer
	})
}

// WithJSONSerializer configures Kid's JSON serializer.
func WithJSONSerializer(serializer serializer.Serializer) Option {
	panicIfNil(serializer, "json serializer cannot be nil")

	return optionImpl(func(k *Kid) {
		k.jsonSerializer = serializer
	})
}

// WithErrorHandler configures Kid's error handler.
func WithErrorHandler(errHandler ErrorHandler) Option {
	panicIfNil(errHandler, "error handler cannot be nil")

	return optionImpl(func(k *Kid) {
		k.errorHandler = errHandler
	})
}

// WithNotFoundHandler configures Kid's not found handler.
func WithNotFoundHandler(handler HandlerFunc) Option {
	panicIfNil(handler, "not found handler cannot be nil")

	return optionImpl(func(k *Kid) {
		k.notFoundHandler = handler
	})
}

// WithMethodNotAllowedHandler configures Kid's method not allowed handler.
func WithMethodNotAllowedHandler(handler HandlerFunc) Option {
	panicIfNil(handler, "method not allowed handler cannot be nil")

	return optionImpl(func(k *Kid) {
		k.methodNotAllowedHandler = handler
	})
}
