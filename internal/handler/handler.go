package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imkishore16/go-cloudStorage/internal/handler/middleware"
	"github.com/imkishore16/go-cloudStorage/internal/model"
	"github.com/imkishore16/go-cloudStorage/internal/model/apperrors"
)

// Handler struct holds required services for handler to function
type Handler struct {
	UserService  model.UserService
	TokenService model.TokenService
	MaxBodyBytes int64
}

// Config will hold services that will eventually be injected into this
// handler layer on handler initialization
type Config struct {
	R               *gin.Engine
	UserService     model.UserService
	TokenService    model.TokenService
	BaseURL         string
	TimeoutDuration time.Duration
	MaxBodyBytes    int64
}

// NewHandler initializes the handler with required injected services along with http routes
// Does not return as it deals directly with a reference to the gin Engine
func NewHandler(c *Config) {
	// Create a handler (which will later have injected services)
	h := &Handler{
		UserService:  c.UserService,
		TokenService: c.TokenService,
		MaxBodyBytes: c.MaxBodyBytes,
	} // currently has no properties

	// Create an account group
	g := c.R.Group(c.BaseURL)

	if gin.Mode() != gin.TestMode {
		g.Use(middleware.Timeout(c.TimeoutDuration, apperrors.NewServiceUnavailable()))
		g.POST("/image", middleware.AuthUser(h.TokenService), h.Image)
		g.DELETE("/image", middleware.AuthUser(h.TokenService), h.DeleteImage)
	} else {
		g.POST("/image", h.Image)
		g.DELETE("/image", h.DeleteImage)
	}

	g.POST("/signup", h.Signup)
	g.POST("/signin", h.Signin)
}
