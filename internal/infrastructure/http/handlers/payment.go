package handlers

import (
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/payments"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/shared/errors"
	"clinic-vet-api/internal/shared/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentHandler struct {
	queryService       payments.QueryService
	commandService     payments.CommandService
	validator          *validator.Validate
	mapper             *mappers.PaymentMapper
	customerIDResolver CustomerIDResolver
}

func NewPaymentHandler(
	queryService payments.QueryService,
	commandService payments.CommandService,
	validator *validator.Validate,
	mapper *mappers.PaymentMapper,
	customerIDResolver CustomerIDResolver,
) *PaymentHandler {
	return &PaymentHandler{
		queryService:       queryService,
		commandService:     commandService,
		validator:          validator,
		mapper:             mapper,
		customerIDResolver: customerIDResolver,
	}
}

// @Summary Get my payment
// @Description Returns a payment belonging to the authenticated customer. Requires customer role.
// @Tags payments
// @Security BearerAuth
// @Param id path int true "Payment ID"
// @Success 200 {object} http.APIResponse{data=dtos.PaymentResponse}
// @Router /me/payments/{id} [get]
func (s *PaymentHandler) GetMyPayment(c *gin.Context) {
	id, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}
	customerID, err := CustomerIDFromContext(c, s.customerIDResolver)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	payment, err := s.queryService.GetPaymentByID(c.Request.Context(), payments.NewPaymentID(id))
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	if payment.PaidFromCustomer.Int32() != int32(customerID) {
		http.NotFound(c, nil)
		return
	}
	http.Found(c, s.mapper.ToPaymentResponse(payment), "Payment")
}

// @Summary List my payments
// @Description Lists payments belonging to the authenticated customer. Requires customer role.
// @Tags payments
// @Security BearerAuth
// @Success 200 {object} http.APIResponse{data=[]dtos.PaymentResponse}
// @Router /me/payments [get]
func (s *PaymentHandler) GetMyPayments(c *gin.Context) {
	customerID, err := CustomerIDFromContext(c, s.customerIDResolver)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	paymentPage, err := s.queryService.GetPaymentsByCustomerID(c.Request.Context(), customers.NewCustomerID(customerID))
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	responsePage := s.mapper.ToPaginatedResponse(paymentPage)
	http.Paginated(c, &responsePage, "Payments")
}

// GetPaymentByID godoc
// @Summary      Get payment by ID
// @Description  Returns a single payment by ID.
// @Tags         payments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Payment ID"
// @Success      200  {object}  http.APIResponse{data=dtos.PaymentResponse}  "Payment found"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      404  {object}  http.APIResponse  "Payment not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /payments/{id} [get]
func (s *PaymentHandler) GetPaymentByID(c *gin.Context) {
	paymentIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	paymentID := payments.NewPaymentID(paymentIDUInt)
	payment, err := s.queryService.GetPaymentByID(c.Request.Context(), paymentID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	paymentResponse := s.mapper.ToPaymentResponse(payment)
	http.Found(c, paymentResponse, "Payment")
}

// @Summary Get payment by reference
// @Description Requires admin or receptionist role.
// @Tags payments
// @Security BearerAuth
// @Param reference path string true "Payment reference"
// @Success 200 {object} http.APIResponse{data=dtos.PaymentResponse}
// @Router /payments/reference/{reference} [get]
func (s *PaymentHandler) GetPaymentByReference(c *gin.Context) {
	reference := c.Param("reference")
	payment, err := s.queryService.GetPaymentByReference(c.Request.Context(), reference)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	paymentResponse := s.mapper.ToPaymentResponse(payment)
	http.Found(c, paymentResponse, "Payment")
}

// @Summary Search payments
// @Description Requires admin or receptionist role.
// @Tags payments
// @Security BearerAuth
// @Param page_number query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} http.APIResponse{data=[]dtos.PaymentResponse}
// @Router /payments [get]
func (s *PaymentHandler) GetPaymentsBySpecification(c *gin.Context) {
	var requestBodyData dtos.PaymentSearchRequest
	if err := http.ShouldBindAndValidateQuery(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	query, err := s.mapper.RequestToSpecification(requestBodyData)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	paymentPage, err := s.queryService.GetPaymentsBySpecification(c.Request.Context(), query)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	responsePage := s.mapper.ToPaginatedResponse(paymentPage)
	http.Paginated(c, &responsePage, "Payments")
}

// @Summary List customer payments
// @Description Requires admin or receptionist role.
// @Tags payments
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} http.APIResponse{data=[]dtos.PaymentResponse}
// @Router /payments/customers/{id} [get]
func (s *PaymentHandler) GetPaymentsByCustomerID(c *gin.Context) {
	customerIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	customerID := customers.NewCustomerID(customerIDUInt)
	paymentPage, err := s.queryService.GetPaymentsByCustomerID(c.Request.Context(), customerID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	responsePage := s.mapper.ToPaginatedResponse(paymentPage)
	http.Paginated(c, &responsePage, "Payments")
}

// @Summary Create payment
// @Description Requires admin or receptionist role.
// @Tags payments
// @Security BearerAuth
// @Param body body dtos.PaymentCreateRequest true "Payment"
// @Success 201 {object} http.APIResponse
// @Router /payments [post]
func (s *PaymentHandler) CreatePayment(c *gin.Context) {
	var requestBodyData dtos.PaymentCreateRequest
	if err := http.ShouldBindAndValidateBody(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command, err := s.mapper.RequestToCreateCommand(requestBodyData)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	payment, err := s.commandService.CreatePayment(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Created(c, payment.ID.Value(), "Payment")
}

// @Summary Update payment
// @Description Requires admin or receptionist role.
// @Tags payments
// @Security BearerAuth
// @Param id path int true "Payment ID"
// @Param body body dtos.PaymentUpdateRequest true "Payment"
// @Success 200 {object} http.APIResponse
// @Router /payments/{id} [put]
func (s *PaymentHandler) UpdatePayment(c *gin.Context) {
	paymentIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	var requestBodyData dtos.PaymentUpdateRequest
	if err := http.ShouldBindAndValidateBody(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	paymentID := payments.NewPaymentID(paymentIDUInt)
	current, err := s.queryService.GetPaymentByID(c.Request.Context(), paymentID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	command, err := s.mapper.RequestToUpdateCommand(paymentIDUInt, requestBodyData, current)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	if err := s.commandService.UpdatePayment(c.Request.Context(), command); err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Updated(c, nil, "Payment")
}

// @Summary Delete payment
// @Description Requires admin or receptionist role.
// @Tags payments
// @Security BearerAuth
// @Param id path int true "Payment ID"
// @Success 200 {object} http.APIResponse
// @Router /payments/{id} [delete]
func (s *PaymentHandler) DeletePayment(c *gin.Context) {
	paymentIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	paymentID := payments.NewPaymentID(paymentIDUInt)
	command := payments.DeletePaymentCommand{
		ID: paymentID,
	}

	err = s.commandService.DeletePayment(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Payment")
}

// @Summary Cancel payment
// @Description Requires admin or receptionist role.
// @Tags payments
// @Security BearerAuth
// @Param id path int true "Payment ID"
// @Param reason formData string false "Cancellation reason"
// @Success 200 {object} http.APIResponse
// @Router /payments/{id}/cancel [post]
func (s *PaymentHandler) CancelPayment(c *gin.Context) {
	paymentIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	paymentID := payments.NewPaymentID(paymentIDUInt)
	command := payments.CancelPaymentCommand{
		ID: paymentID, Reason: c.PostForm("reason"),
	}

	err = s.commandService.CancelPayment(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Payment")
}

// @Summary Process payment
// @Description Requires admin or receptionist role.
// @Tags payments
// @Security BearerAuth
// @Param id path int true "Payment ID"
// @Param reason formData string false "Processing reason"
// @Success 200 {object} http.APIResponse
// @Router /payments/{id}/process [post]
func (s *PaymentHandler) ProcessPayment(c *gin.Context) {
	paymentIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	paymentID := payments.NewPaymentID(paymentIDUInt)
	command := payments.ProcessPaymentCommand{
		ID: paymentID, Reason: c.PostForm("reason"),
	}

	err = s.commandService.ProcessPayment(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Payment")
}

// @Summary Refund payment
// @Description Requires admin or receptionist role.
// @Tags payments
// @Security BearerAuth
// @Param id path int true "Payment ID"
// @Param reason formData string false "Refund reason"
// @Success 200 {object} http.APIResponse
// @Router /payments/{id}/refund [post]
func (s *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	paymentID := payments.NewPaymentID(paymentIDUInt)
	command := payments.RefundPaymentCommand{
		ID: paymentID, Reason: c.PostForm("reason"),
	}

	err = s.commandService.RefundPayment(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Payment")
}
