package http

import (
	"fmt"
	"net/http"

	"clinic-vet-api/internal/shared/errors"
	"clinic-vet-api/internal/shared/page"

	"github.com/gin-gonic/gin"
)

func Success(ctx *gin.Context, data any, message string) {
	response := APIResponse{}
	response.SuccessRequest(data, message)

	ctx.JSON(200, response)
}

func Paginated[T any](ctx *gin.Context, pag *page.Page[T], entity string) {
	response := APIResponse{}
	if pag == nil {
		response.SuccessRequest(nil, "No data found")
		ctx.JSON(200, response)
		return
	}

	message := fmt.Sprintf("%s retrieved successfully", entity)
	response.SuccessWithMeta(pag.Items, pag.Metadata, message)

	ctx.JSON(200, response)
}

func SuccessCreated(ctx *gin.Context, data any, message string) {
	response := APIResponse{}
	response.SuccessRequest(data, message)

	ctx.JSON(201, response)
}

func Found(ctx *gin.Context, data any, entityName string) {
	if entityName == "" {
		entityName = "Resource"
	}
	messsage := entityName + " found successfully"

	response := APIResponse{}
	response.SuccessRequest(data, messsage)

	ctx.JSON(200, response)
}

func Created(ctx *gin.Context, entityID any, entityName string) {
	messsage := entityName + " has been created successfully"
	response := APIResponse{}
	response.SuccessRequest(gin.H{"id": entityID}, messsage)

	ctx.JSON(201, response)
}

func Updated(ctx *gin.Context, data any, entityName string) {
	messsage := entityName + " has been updated successfully"
	response := APIResponse{}
	response.SuccessRequest(data, messsage)

	ctx.JSON(200, response)
}

func NotFound(ctx *gin.Context, err error) {
	response := APIResponse{}
	errorResponse := response.ErrorRequest(err)
	ctx.JSON(404, errorResponse)
}

func BadRequest(ctx *gin.Context, err error) {
	response := APIResponse{}
	errorResponse := response.ErrorRequest(err)

	ctx.JSON(400, errorResponse)
}

func Unauthorized(ctx *gin.Context, err error) {
	response := APIResponse{}
	errorResponse := response.ErrorRequest(err)

	ctx.JSON(401, errorResponse)
}

func Forbidden(ctx *gin.Context, err error) {
	response := APIResponse{}
	errorResponse := response.ErrorRequest(err)

	ctx.JSON(403, errorResponse)
}

func ApplicationError(ctx *gin.Context, err error) {
	response := APIResponse{}
	errorResponse := response.ErrorRequest(err)
	httpStatusCode := http.StatusInternalServerError

	switch e := err.(type) {
	case interface{ HTTPStatus() int }:
		ctx.JSON(e.HTTPStatus(), e)
		return
	}

	switch e := err.(type) {
	case errors.BaseApplicationError:
		ctx.JSON(e.HTTPStatus(), e)
		return
	default:
		httpStatusCode = http.StatusInternalServerError
	}

	ctx.JSON(httpStatusCode, errorResponse)
}
