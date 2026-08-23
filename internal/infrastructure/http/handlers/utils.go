package handlers

import (
	"context"

	"clinic-vet-api/internal/core/addresses"
	"clinic-vet-api/internal/middleware"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/errors"
	"clinic-vet-api/internal/shared/http"

	"github.com/gin-gonic/gin"
)

// EmployeeIDResolver resolves user ID to employee ID (e.g. via employee repository).
type EmployeeIDResolver func(ctx context.Context, userID uint) (uint, error)

type IdentityProvider func(c *gin.Context) (shared.UserID, error)
type EntityIDProvider func(c *gin.Context) (shared.IntegerID, error)
type OptionalEntityIDProvider func(c *gin.Context) (*shared.IntegerID, error)

type OptionalIdentityProvider func(c *gin.Context) (*shared.UserID, error)

func UserIDFromContext(c *gin.Context) (shared.UserID, error) {
	userCTX, ok := middleware.GetUserFromContext(c)
	if !ok {
		return shared.UserID{}, errors.UserNotFoundInContextError(c.Request.Context(), "Identity")
	}
	return userCTX.UserIDValueObject(), nil
}

func UserIDFromContextPtr(c *gin.Context) (*shared.UserID, error) {
	userCTX, ok := middleware.GetUserFromContext(c)
	if !ok {
		return nil, errors.UserNotFoundInContextError(c.Request.Context(), "Identity")
	}
	userID := userCTX.UserIDValueObject()
	return &userID, nil
}

func AddressIDFromParam(c *gin.Context) (shared.IntegerID, error) {
	entityID, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		return addresses.NewAddressIDEmpty(), err
	}
	return addresses.NewAddressID(entityID), nil
}

func UserIDFromParam(c *gin.Context) (shared.UserID, error) {
	userIDUInt, err := http.ParseParamToUInt(c, "user_id")
	if err != nil {
		return shared.UserID{}, err
	}
	return shared.NewUserID(userIDUInt), nil
}

func UserIDFromParamOptional(c *gin.Context) (*shared.UserID, error) {
	userIDUInt, err := http.ParseParamToUInt(c, "user_id")
	if err != nil {
		return nil, err
	}
	userID := shared.NewUserID(userIDUInt)
	return &userID, nil
}

// CustomerIDProvider returns a customer ID from the request (e.g. param or resolved from auth).
type CustomerIDProvider func(c *gin.Context) (uint, error)

// CustomerIDResolver resolves user ID to customer ID (e.g. via customer repository).
// Used to build a CustomerIDProvider for "my" endpoints.
type CustomerIDResolver func(ctx context.Context, userID uint) (uint, error)

// CustomerIDFromParam returns customer_id from the URL path param "customer_id".
func CustomerIDFromParam(c *gin.Context) (uint, error) {
	return http.ParseParamToUInt(c, "customer_id")
}

// CustomerIDFromIDParam returns customer ID from the URL path param "id" (e.g. /customers/:id/appointments).
func CustomerIDFromIDParam(c *gin.Context) (uint, error) {
	return http.ParseParamToUInt(c, "id")
}

// CustomerIDFromContext returns the customer ID for the authenticated user by calling the resolver.
func CustomerIDFromContext(c *gin.Context, resolver CustomerIDResolver) (uint, error) {
	userID, err := UserIDFromContext(c)
	if err != nil {
		return 0, err
	}
	return resolver(c.Request.Context(), userID.Value())
}

// OptionalCustomerIDProvider returns an optional customer ID from the request.
type OptionalCustomerIDProvider func(c *gin.Context) (*uint, error)

// CustomerIDFromParamOptional returns customer_id from the URL param if present; otherwise (nil, nil).
func CustomerIDFromParamOptional(c *gin.Context) (*uint, error) {
	if c.Param("customer_id") == "" {
		return nil, nil
	}
	id, err := http.ParseParamToUInt(c, "customer_id")
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// EmployeeIDProvider returns an employee ID from the request (e.g. param or resolved from auth).
type EmployeeIDProvider func(c *gin.Context) (uint, error)

// EmployeeIDFromParam returns employee_id from the URL path param "employee_id".
func EmployeeIDFromParam(c *gin.Context) (uint, error) {
	return http.ParseParamToUInt(c, "employee_id")
}

// EmployeeIDFromIDParam returns employee ID from the URL path param "id" (e.g. /employees/:id/appointments).
func EmployeeIDFromIDParam(c *gin.Context) (uint, error) {
	return http.ParseParamToUInt(c, "id")
}

// EmployeeIDFromContext returns the employee ID for the authenticated user by calling the resolver.
func EmployeeIDFromContext(c *gin.Context, resolver EmployeeIDResolver) (uint, error) {
	userID, err := UserIDFromContext(c)
	if err != nil {
		return 0, err
	}
	return resolver(c.Request.Context(), userID.Value())
}

// OptionalEmployeeIDProvider returns an optional employee ID (e.g. from param or nil).
type OptionalEmployeeIDProvider func(c *gin.Context) (*uint, error)

// EmployeeIDFromParamOptional returns employee_id from the URL param if present; otherwise (nil, nil).
func EmployeeIDFromParamOptional(c *gin.Context) (*uint, error) {
	if c.Param("employee_id") == "" {
		return nil, nil
	}
	id, err := http.ParseParamToUInt(c, "employee_id")
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// EmployeeIDFromQuery returns employee_id from the query string (e.g. ?employee_id=1).
func EmployeeIDFromQuery(c *gin.Context) (uint, error) {
	return http.ParseQueryToUInt(c, "employee_id")
}

// OptionalEmployeeIDFromQuery returns employee_id from query if present; otherwise (nil, nil).
func OptionalEmployeeIDFromQuery(c *gin.Context) (*uint, error) {
	if c.Query("employee_id") == "" {
		return nil, nil
	}
	id, err := http.ParseQueryToUInt(c, "employee_id")
	if err != nil {
		return nil, err
	}
	return &id, nil
}
