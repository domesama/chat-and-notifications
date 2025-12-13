package httpserverwrapper

import (
	"github.com/gin-gonic/gin"
)

// HTTPServerBuilder provides a fluent interface for configuring HTTP server
type HTTPServerBuilder struct {
	config      HTTPServerConfig
	middlewares []gin.HandlerFunc
	mode        string
}

// NewHTTPServerBuilder creates a new HTTP server builder
func NewHTTPServerBuilder(cfg HTTPServerConfig) *HTTPServerBuilder {
	return &HTTPServerBuilder{
		config: cfg,
		mode:   gin.ReleaseMode,
	}
}

// WithMiddleware adds middleware to the server
func (b *HTTPServerBuilder) WithMiddleware(middleware ...gin.HandlerFunc) *HTTPServerBuilder {
	b.middlewares = append(b.middlewares, middleware...)
	return b
}

// WithDebugMode sets the server to debug mode
func (b *HTTPServerBuilder) WithDebugMode() *HTTPServerBuilder {
	b.mode = gin.DebugMode
	return b
}

// Build creates the Gin engine with configured settings
func (b *HTTPServerBuilder) Build() *gin.Engine {
	gin.SetMode(b.mode)
	engine := gin.New()

	// Add configured middlewares
	for _, middleware := range b.middlewares {
		engine.Use(middleware)
	}

	return engine
}
