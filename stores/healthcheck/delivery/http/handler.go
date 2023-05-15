package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/x-xyz/goapi/base/ctx"
	hcdomain "github.com/x-xyz/goapi/domain/healthcheck"
)

// ResponseError represent the reseponse error struct
type ResponseError struct {
	Message string `json:"message"`
}

type healthCheckHandler struct {
	healthCheck hcdomain.HealthCheckUsecase
}

// New will initialize the healthcheck/
func New(e *echo.Echo, us hcdomain.HealthCheckUsecase) {
	handler := &healthCheckHandler{
		healthCheck: us,
	}
	g := e.Group("/health")
	g.GET("", handler.check)
}

func (h *healthCheckHandler) check(c echo.Context) error {
	context := c.Get("ctx").(ctx.Ctx)
	if err := h.healthCheck.Check(context); err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseError{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"healthy": "ok",
	})
}
