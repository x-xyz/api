package delivery

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/service/query"
)

type JsonResponseStatus string

const (
	JsonResponseStatusSuccess JsonResponseStatus = "success"
	JsonResponseStatusFail    JsonResponseStatus = "fail"
)

type JsonResponse struct {
	Data   interface{}        `json:"data"`
	Status JsonResponseStatus `json:"status"`
}

func MakeJsonResp(c echo.Context, status int, data interface{}) error {
	if err, ok := data.(error); ok {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, query.ErrNotFound) {
			status = http.StatusNotFound
		}
		data = err.Error()
	}

	if status >= 400 {
		return c.JSON(status, JsonResponse{data, JsonResponseStatusFail})
	}

	if status >= 200 && status < 300 {
		return c.JSON(status, JsonResponse{data, JsonResponseStatusSuccess})
	}

	return c.JSON(status, data)
}
