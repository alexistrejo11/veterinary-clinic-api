package handlers

import (
	"errors"

	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/shared/http"
	"clinic-vet-api/internal/shared/page"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AppointmentHandler handles appointment HTTP with customer-, veterinarian-, and receptionist-scoped endpoints.
type AppointmentHandler struct {
	commandService     appointments.CommandService
	queryService       appointments.QueryService
	validator          *validator.Validate
	mapper             *mappers.AppointmentResponseMapper
	customerIDResolver CustomerIDResolver
	employeeIDResolver EmployeeIDResolver
}

func NewAppointmentHandler(
	commandService appointments.CommandService,
	queryService appointments.QueryService,
	validator *validator.Validate,
	mapper *mappers.AppointmentResponseMapper,
	customerIDResolver CustomerIDResolver,
	employeeIDResolver EmployeeIDResolver,
) *AppointmentHandler {
	if mapper == nil {
		mapper = mappers.NewAppointmentResponseMapper()
	}
	return &AppointmentHandler{
		commandService:     commandService,
		queryService:       queryService,
		validator:          validator,
		mapper:             mapper,
		customerIDResolver: customerIDResolver,
		employeeIDResolver: employeeIDResolver,
	}
}

func parseAppointmentIDFromParam(c *gin.Context) (appointments.AppointmentID, error) {
	id, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		return appointments.AppointmentID{}, err
	}
	return appointments.NewAppointmentID(id), nil
}

// ------------------------------------------------------------
// Internal handlers
// ------------------------------------------------------------

func (h *AppointmentHandler) getAppointmentByIDInternal(
	ctx *gin.Context,
	apptID appointments.AppointmentID,
	optCustomerID *uint,
	optEmployeeID *uint,
) (any, error) {
	if optCustomerID != nil {
		cid := customers.NewCustomerID(*optCustomerID)
		return h.queryService.GetByIDAndCustomerID(ctx.Request.Context(), apptID, cid)
	}
	if optEmployeeID != nil {
		eid := employees.NewEmployeeID(*optEmployeeID)
		return h.queryService.GetByIDAndEmployeeID(ctx.Request.Context(), apptID, eid)
	}
	return h.queryService.GetByID(ctx.Request.Context(), apptID)
}

func (h *AppointmentHandler) getAppointmentsByCustomerIDInternal(
	ctx *gin.Context,
	getCustomerID CustomerIDProvider,
	pagination page.Pagination,
) (any, error) {
	customerID, err := getCustomerID(ctx)
	if err != nil {
		return nil, err
	}
	cid := customers.NewCustomerID(customerID)
	return h.queryService.GetByCustomerID(ctx.Request.Context(), cid, pagination)
}

func (h *AppointmentHandler) getAppointmentsByEmployeeIDInternal(
	ctx *gin.Context,
	getEmployeeID EmployeeIDProvider,
	pagination page.Pagination,
) (any, error) {
	employeeID, err := getEmployeeID(ctx)
	if err != nil {
		return nil, err
	}
	eid := employees.NewEmployeeID(employeeID)
	return h.queryService.GetByEmployeeID(ctx.Request.Context(), eid, pagination)
}

func (h *AppointmentHandler) getAppointmentsBySpecificationInternal(
	ctx *gin.Context,
	searchReq dtos.AppointmentSearchRequest,
) (any, error) {
	query, err := h.mapper.RequestToSearchQuery(searchReq)
	if err != nil {
		return nil, err
	}
	return h.queryService.GetBySpecfication(ctx.Request.Context(), query)
}

func (h *AppointmentHandler) requestAppointmentInternal(
	ctx *gin.Context,
	req dtos.AppointmentRequestByCustomerRequest,
	getCustomerID CustomerIDProvider,
) (any, error) {
	customerID, err := getCustomerID(ctx)
	if err != nil {
		return nil, err
	}
	cid := customers.NewCustomerID(customerID)
	command, err := h.mapper.RequestToRequestByCustomerCommand(req, cid)
	if err != nil {
		return nil, err
	}
	err = h.commandService.RequestAppointment(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *AppointmentHandler) createAppointmentInternal(
	ctx *gin.Context,
	req dtos.AppointmentCreateRequest,
	optEmployeeID *uint,
) (any, error) {
	command, err := h.mapper.RequestToCreateCommand(req, optEmployeeID)
	if err != nil {
		return nil, err
	}
	created, err := h.commandService.CreateAppointment(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return created.ID.Value(), nil
}

func (h *AppointmentHandler) updateAppointmentInternal(
	ctx *gin.Context,
	req dtos.AppointmentUpdateGeneralInfoRequest,
	apptID uint,
	optEmployeeID *uint,
) (any, error) {
	cmd, err := h.mapper.RequestToUpdateCommand(req, apptID, optEmployeeID)
	if err != nil {
		return nil, err
	}
	err = h.commandService.UpdateAppointmentGeneralInfo(ctx.Request.Context(), cmd)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *AppointmentHandler) deleteAppointmentInternal(ctx *gin.Context, apptID uint, isHard bool) (any, error) {
	command := h.mapper.ToDeleteCommand(apptID, isHard)
	err := h.commandService.DeleteAppointment(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *AppointmentHandler) rescheduleInternal(
	ctx *gin.Context,
	req dtos.RescheduleAppointmentRequest,
	apptID uint,
	optEmployeeID *uint,
) (any, error) {
	command := h.mapper.RequestToRescheduleCommand(req, apptID, optEmployeeID)
	err := h.commandService.RescheduleAppointment(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *AppointmentHandler) confirmInternal(ctx *gin.Context, apptID uint, employeeID uint) (any, error) {
	command := h.mapper.ToConfirmCommand(apptID, employeeID)
	err := h.commandService.ConfirmAppointment(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *AppointmentHandler) completeInternal(ctx *gin.Context, apptID uint, optEmployeeID *uint, notes string) (any, error) {
	command := h.mapper.ToCompleteCommand(apptID, optEmployeeID, notes)
	err := h.commandService.CompleteAppointment(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *AppointmentHandler) cancelInternal(ctx *gin.Context, apptID uint, optEmployeeID, optCustomerID *uint, reason string) (any, error) {
	command := h.mapper.ToCancelCommand(apptID, optEmployeeID, optCustomerID, reason)
	err := h.commandService.CancelAppointment(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *AppointmentHandler) notAttendInternal(ctx *gin.Context, apptID uint, optEmployeeID *uint) (any, error) {
	command := h.mapper.ToNotAttendCommand(apptID, optEmployeeID)
	err := h.commandService.MarkAppointmentAsNotAttend(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *AppointmentHandler) getMyCustomerID(c *gin.Context) (uint, error) {
	if h.customerIDResolver == nil {
		return 0, errors.New("customer resolver not configured")
	}
	return CustomerIDFromContext(c, h.customerIDResolver)
}

func (h *AppointmentHandler) getMyEmployeeID(c *gin.Context) (uint, error) {
	if h.employeeIDResolver == nil {
		return 0, errors.New("employee resolver not configured")
	}
	return EmployeeIDFromContext(c, h.employeeIDResolver)
}

// ------------------------------------------------------------
// Customer handlers (only their appointments)
// ------------------------------------------------------------

// GetMyAppointment godoc
// @Summary      Get my appointment by ID (customer)
// @Description  Returns a single appointment belonging to the authenticated customer. Requires customer role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Appointment ID"
// @Success      200  {object}  http.APIResponse{data=dtos.AppointmentResponse}  "Appointment found"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      404  {object}  http.APIResponse  "Appointment not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /me/appointments/{id} [get]
func (h *AppointmentHandler) GetMyAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := h.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.getAppointmentByIDInternal(ctx, apptID, &customerID, nil)
	}
	http.HandleGetRequest(h.validator, "Appointment", logic)(c)
}

// GetMyAppointments godoc
// @Summary      Get my appointments (customer)
// @Description  Returns paginated appointments for the authenticated customer. Requires customer role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page_number  query  int  false  "Page number"  default(1)
// @Param        page_size    query  int  false  "Page size"    default(10)
// @Success      200          {object}  http.APIResponse{data=[]dtos.AppointmentResponse}  "Paginated appointments"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /me/appointments [get]
func (h *AppointmentHandler) GetMyAppointments(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		getCustomerID := func(c *gin.Context) (uint, error) { return h.getMyCustomerID(c) }
		return h.getAppointmentsByCustomerIDInternal(ctx, getCustomerID, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No appointments found")
			return
		}
		p := res.(page.Page[appointments.Appointment])
		responsePage := h.mapper.ToResponsePage(p)
		http.Paginated(c, responsePage, "Appointments")
	})(c)
}

// RequestAppointment godoc
// @Summary      Request appointment (customer)
// @Description  Submits an appointment request as a customer. Requires customer role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.AppointmentRequestByCustomerRequest  true  "Appointment request"
// @Success      200    {object}  http.APIResponse                          "Request submitted"
// @Failure      400    {object}  http.APIResponse                          "Validation error"
// @Failure      401    {object}  http.APIResponse                          "Unauthorized"
// @Failure      500    {object}  http.APIResponse                          "Internal server error"
// @Router       /me/appointments [post]
func (h *AppointmentHandler) RequestAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AppointmentRequestByCustomerRequest) (any, error) {
		getCustomerID := func(c *gin.Context) (uint, error) { return h.getMyCustomerID(c) }
		return h.requestAppointmentInternal(ctx, req, getCustomerID)
	}
	http.HandleRequestWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment request submitted successfully")
	})(c)
}

// CancelMyAppointment godoc
// @Summary Cancel my appointment (customer)
// @Description Cancels a cancellable appointment belonging to the authenticated customer. Requires customer role.
// @Tags appointments
// @Security BearerAuth
// @Param id path int true "Appointment ID"
// @Param reason query string false "Cancellation reason"
// @Success 200 {object} http.APIResponse "Appointment cancelled"
// @Failure 400 {object} http.APIResponse "Invalid ID or appointment cannot be cancelled"
// @Failure 401 {object} http.APIResponse "Unauthorized"
// @Failure 404 {object} http.APIResponse "Appointment not found"
// @Router /me/appointments/{id}/cancel [post]
func (h *AppointmentHandler) CancelMyAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := h.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.cancelInternal(ctx, apptID.Value(), nil, &customerID, ctx.Query("reason"))
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment cancelled successfully")
	})(c)
}

// ------------------------------------------------------------
// Employee handlers (only their assigned appointments)
// ------------------------------------------------------------

// GetMyAppointmentAsEmployee godoc
// @Summary      Get my assigned appointment by ID (veterinarian)
// @Description  Returns a single appointment assigned to the authenticated veterinarian. Requires veterinarian role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Appointment ID"
// @Success      200  {object}  http.APIResponse{data=dtos.AppointmentResponse}  "Appointment found"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      404  {object}  http.APIResponse  "Appointment not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /employees/appointments/{id} [get]
func (h *AppointmentHandler) GetMyAppointmentAsEmployee(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		employeeID, err := h.getMyEmployeeID(ctx)
		if err != nil {
			return nil, err
		}
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.getAppointmentByIDInternal(ctx, apptID, nil, &employeeID)
	}
	http.HandleGetRequest(h.validator, "Appointment", logic)(c)
}

// GetMyAppointmentsAsEmployee godoc
// @Summary      Get my assigned appointments (veterinarian)
// @Description  Returns paginated appointments assigned to the authenticated veterinarian. Requires veterinarian role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page_number  query  int  false  "Page number"  default(1)
// @Param        page_size    query  int  false  "Page size"    default(10)
// @Success      200          {object}  http.APIResponse{data=[]dtos.AppointmentResponse}  "Paginated appointments"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /employees/appointments [get]
func (h *AppointmentHandler) GetMyAppointmentsAsEmployee(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		getEmployeeID := func(c *gin.Context) (uint, error) { return h.getMyEmployeeID(c) }
		return h.getAppointmentsByEmployeeIDInternal(ctx, getEmployeeID, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No appointments found")
			return
		}
		p := res.(page.Page[appointments.Appointment])
		responsePage := h.mapper.ToResponsePage(p)
		http.Paginated(c, responsePage, "Appointments")
	})(c)
}

// UpdateMyAppointment godoc
// @Summary      Update my appointment (veterinarian)
// @Description  Updates general info (reason, notes, service) of an appointment assigned to the authenticated veterinarian. Requires veterinarian role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int                                        true  "Appointment ID"
// @Param        body   body      dtos.AppointmentUpdateGeneralInfoRequest   true  "Fields to update"
// @Success      200    {object}  http.APIResponse  "Appointment updated"
// @Failure      400    {object}  http.APIResponse  "Validation error"
// @Failure      401    {object}  http.APIResponse  "Unauthorized"
// @Failure      404    {object}  http.APIResponse  "Appointment not found"
// @Failure      500    {object}  http.APIResponse  "Internal server error"
// @Router       /employees/appointments/{id} [put]
func (h *AppointmentHandler) UpdateMyAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AppointmentUpdateGeneralInfoRequest) (any, error) {
		employeeID, err := h.getMyEmployeeID(ctx)
		if err != nil {
			return nil, err
		}
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.updateAppointmentInternal(ctx, req, apptID.Value(), &employeeID)
	}
	http.HandleUpdateRequest(h.validator, "Appointment", logic)(c)
}

// RescheduleMyAppointment godoc
// @Summary      Reschedule my appointment (veterinarian)
// @Description  Reschedules an appointment assigned to the authenticated veterinarian to a new date/time. Requires veterinarian role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int                                  true  "Appointment ID"
// @Param        body   body      dtos.RescheduleAppointmentRequest    true  "New date/time and optional reason"
// @Success      200    {object}  http.APIResponse  "Appointment rescheduled"
// @Failure      400    {object}  http.APIResponse  "Validation error"
// @Failure      401    {object}  http.APIResponse  "Unauthorized"
// @Failure      404    {object}  http.APIResponse  "Appointment not found"
// @Failure      500    {object}  http.APIResponse  "Internal server error"
// @Router       /employees/appointments/{id}/reschedule [post]
func (h *AppointmentHandler) RescheduleMyAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.RescheduleAppointmentRequest) (any, error) {
		employeeID, err := h.getMyEmployeeID(ctx)
		if err != nil {
			return nil, err
		}
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.rescheduleInternal(ctx, req, apptID.Value(), &employeeID)
	}
	http.HandleRequestWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment rescheduled successfully")
	})(c)
}

// CompleteMyAppointment godoc
// @Summary      Complete my appointment (veterinarian)
// @Description  Marks an appointment assigned to the authenticated veterinarian as completed. Requires veterinarian role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path   int     true   "Appointment ID"
// @Param        notes  query  string  false  "Completion notes"
// @Success      200    {object}  http.APIResponse  "Appointment completed"
// @Failure      400    {object}  http.APIResponse  "Invalid ID"
// @Failure      401    {object}  http.APIResponse  "Unauthorized"
// @Failure      404    {object}  http.APIResponse  "Appointment not found"
// @Failure      500    {object}  http.APIResponse  "Internal server error"
// @Router       /employees/appointments/{id}/complete [post]
func (h *AppointmentHandler) CompleteMyAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		employeeID, err := h.getMyEmployeeID(ctx)
		if err != nil {
			return nil, err
		}
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		notes := ctx.Query("notes")
		return h.completeInternal(ctx, apptID.Value(), &employeeID, notes)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment completed successfully")
	})(c)
}

// CancelAssignedAppointment godoc
// @Summary      Cancel my appointment (veterinarian)
// @Description  Cancels an appointment assigned to the authenticated veterinarian. Requires veterinarian role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int     true   "Appointment ID"
// @Param        reason  query  string  false  "Cancellation reason"
// @Success      200     {object}  http.APIResponse  "Appointment cancelled"
// @Failure      400     {object}  http.APIResponse  "Invalid ID"
// @Failure      401     {object}  http.APIResponse  "Unauthorized"
// @Failure      404     {object}  http.APIResponse  "Appointment not found"
// @Failure      500     {object}  http.APIResponse  "Internal server error"
// @Router       /employees/appointments/{id}/cancel [post]
func (h *AppointmentHandler) CancelAssignedAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		employeeID, err := h.getMyEmployeeID(ctx)
		if err != nil {
			return nil, err
		}
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		reason := ctx.Query("reason")
		return h.cancelInternal(ctx, apptID.Value(), &employeeID, nil, reason)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment cancelled successfully")
	})(c)
}

// NotAttendMyAppointment godoc
// @Summary      Mark my appointment as not attended (veterinarian)
// @Description  Marks an appointment assigned to the authenticated veterinarian as not attended. Requires veterinarian role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Appointment ID"
// @Success      200  {object}  http.APIResponse  "Appointment marked as not attended"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      404  {object}  http.APIResponse  "Appointment not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /employees/appointments/{id}/not-attend [post]
func (h *AppointmentHandler) NotAttendMyAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		employeeID, err := h.getMyEmployeeID(ctx)
		if err != nil {
			return nil, err
		}
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.notAttendInternal(ctx, apptID.Value(), &employeeID)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment marked as not attended successfully")
	})(c)
}

// ------------------------------------------------------------
// Manager/Admin handlers (all appointments)
// ------------------------------------------------------------

// GetAppointmentByID godoc
// @Summary      Get appointment by ID (receptionist/admin)
// @Description  Returns a single appointment by ID. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Appointment ID"
// @Success      200  {object}  http.APIResponse{data=dtos.AppointmentResponse}  "Appointment found"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      404  {object}  http.APIResponse  "Appointment not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /appointments/{id} [get]
func (h *AppointmentHandler) GetAppointmentByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.getAppointmentByIDInternal(ctx, apptID, nil, nil)
	}
	http.HandleGetRequest(h.validator, "Appointment", logic)(c)
}

// SearchAppointments godoc
// @Summary      Search appointments (receptionist/admin)
// @Description  Paginated search of all appointments with filters. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.AppointmentSearchRequest  true  "Search criteria and pagination"
// @Success      200    {object}  http.APIResponse{data=[]dtos.AppointmentResponse}  "Paginated appointments"
// @Failure      400    {object}  http.APIResponse               "Validation error"
// @Failure      401    {object}  http.APIResponse               "Unauthorized"
// @Failure      500    {object}  http.APIResponse               "Internal server error"
// @Router       /appointments [get]
func (h *AppointmentHandler) SearchAppointments(c *gin.Context) {
	searchReq, err := dtos.NewApptSearchRequestFromContext(c)
	if err != nil {
		http.BadRequest(c, err)
		return
	}
	if err := h.validator.Struct(searchReq); err != nil {
		http.BadRequest(c, err)
		return
	}
	logic := func(ctx *gin.Context) (any, error) {
		return h.getAppointmentsBySpecificationInternal(ctx, searchReq)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No appointments found")
			return
		}
		p := res.(page.Page[appointments.Appointment])
		responsePage := h.mapper.ToResponsePage(p)
		http.Paginated(c, responsePage, "Appointments")
	})(c)
}

// GetAppointmentsByCustomerID godoc
// @Summary      Get appointments by customer (receptionist/admin)
// @Description  Returns paginated appointments for a given customer. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path   int  true   "Customer ID"
// @Param        page_number  query  int  false  "Page number"  default(1)
// @Param        page_size    query  int  false  "Page size"    default(10)
// @Success      200          {object}  http.APIResponse{data=[]dtos.AppointmentResponse}  "Paginated appointments"
// @Failure      400          {object}  http.APIResponse  "Invalid ID"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /customers/{id}/appointments [get]
func (h *AppointmentHandler) GetAppointmentsByCustomerID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.getAppointmentsByCustomerIDInternal(ctx, CustomerIDFromIDParam, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No appointments found")
			return
		}
		p := res.(page.Page[appointments.Appointment])
		responsePage := h.mapper.ToResponsePage(p)
		http.Paginated(c, responsePage, "Appointments")
	})(c)
}

// GetAppointmentsByEmployeeID godoc
// @Summary      Get appointments by employee (receptionist/admin)
// @Description  Returns paginated appointments for a given employee. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path   int  true   "Employee ID"
// @Param        page_number  query  int  false  "Page number"  default(1)
// @Param        page_size    query  int  false  "Page size"    default(10)
// @Success      200          {object}  http.APIResponse{data=[]dtos.AppointmentResponse}  "Paginated appointments"
// @Failure      400          {object}  http.APIResponse  "Invalid ID"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /employees/{id}/appointments [get]
func (h *AppointmentHandler) GetAppointmentsByEmployeeID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.getAppointmentsByEmployeeIDInternal(ctx, EmployeeIDFromIDParam, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No appointments found")
			return
		}
		p := res.(page.Page[appointments.Appointment])
		responsePage := h.mapper.ToResponsePage(p)
		http.Paginated(c, responsePage, "Appointments")
	})(c)
}

// GetAppointmentsByPetID godoc
// @Summary      Get appointments by pet (receptionist/admin)
// @Description  Returns paginated appointments for a given pet. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path   int  true   "Pet ID"
// @Param        page_number  query  int  false  "Page number"  default(1)
// @Param        page_size    query  int  false  "Page size"    default(10)
// @Success      200          {object}  http.APIResponse{data=[]dtos.AppointmentResponse}  "Paginated appointments"
// @Failure      400          {object}  http.APIResponse  "Invalid ID"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /pets/{id}/appointments [get]
func (h *AppointmentHandler) GetAppointmentsByPetID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		petIDRaw, err := http.ParseParamToUInt(ctx, "id")
		if err != nil {
			return nil, err
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, c, h.validator); err != nil {
			return nil, err
		}
		petID := pets.NewPetID(petIDRaw)
		page, err := h.queryService.GetByPetID(c.Request.Context(), petID, pageParams.ToPagination())
		if err != nil {
			return nil, err
		}
		return page, nil
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No appointments found")
			return
		}
		p := res.(page.Page[appointments.Appointment])
		responsePage := h.mapper.ToResponsePage(p)
		http.Paginated(c, responsePage, "Appointments")
	})(c)
}

// CreateAppointment godoc
// @Summary      Create appointment (receptionist/admin)
// @Description  Creates a new appointment, optionally assigned to a veterinarian via query param. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        employee_id  query  int                            false  "Employee ID to assign"
// @Param        body         body   dtos.AppointmentCreateRequest  true   "Appointment data"
// @Success      201          {object}  http.APIResponse  "Appointment created (data contains new appointment ID)"
// @Failure      400          {object}  http.APIResponse  "Validation error"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /appointments [post]
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AppointmentCreateRequest) (any, error) {
		optEmployeeID, _ := OptionalEmployeeIDFromQuery(ctx)
		return h.createAppointmentInternal(ctx, req, optEmployeeID)
	}
	http.HandleCreateRequest(h.validator, "Appointment", logic)(c)
}

// UpdateAppointment godoc
// @Summary      Update appointment (receptionist/admin)
// @Description  Updates general info (reason, notes, service) of an appointment. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int                                       true  "Appointment ID"
// @Param        body   body      dtos.AppointmentUpdateGeneralInfoRequest  true  "Fields to update"
// @Success      200    {object}  http.APIResponse  "Appointment updated"
// @Failure      400    {object}  http.APIResponse  "Validation error"
// @Failure      401    {object}  http.APIResponse  "Unauthorized"
// @Failure      404    {object}  http.APIResponse  "Appointment not found"
// @Failure      500    {object}  http.APIResponse  "Internal server error"
// @Router       /appointments/{id} [put]
func (h *AppointmentHandler) UpdateAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AppointmentUpdateGeneralInfoRequest) (any, error) {
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.updateAppointmentInternal(ctx, req, apptID.Value(), nil)
	}
	http.HandleUpdateRequest(h.validator, "Appointment", logic)(c)
}

// DeleteAppointment godoc
// @Summary      Delete appointment (receptionist/admin)
// @Description  Deletes an appointment. Soft-delete by default; pass hard=true for permanent deletion. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path   int     true   "Appointment ID"
// @Param        hard  query  bool    false  "Hard delete (permanent)"  default(false)
// @Success      200   {object}  http.APIResponse  "Appointment deleted"
// @Failure      400   {object}  http.APIResponse  "Invalid ID"
// @Failure      401   {object}  http.APIResponse  "Unauthorized"
// @Failure      404   {object}  http.APIResponse  "Appointment not found"
// @Failure      500   {object}  http.APIResponse  "Internal server error"
// @Router       /appointments/{id} [delete]
func (h *AppointmentHandler) DeleteAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		isHard := ctx.Query("hard") == "true"
		return h.deleteAppointmentInternal(ctx, apptID.Value(), isHard)
	}
	http.HandleDeleteRequest(h.validator, "Appointment", logic)(c)
}

// RescheduleAppointment godoc
// @Summary      Reschedule appointment (receptionist/admin)
// @Description  Reschedules an appointment to a new date/time. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int                                true  "Appointment ID"
// @Param        body   body      dtos.RescheduleAppointmentRequest  true  "New date/time and optional reason"
// @Success      200    {object}  http.APIResponse  "Appointment rescheduled"
// @Failure      400    {object}  http.APIResponse  "Validation error"
// @Failure      401    {object}  http.APIResponse  "Unauthorized"
// @Failure      404    {object}  http.APIResponse  "Appointment not found"
// @Failure      500    {object}  http.APIResponse  "Internal server error"
// @Router       /appointments/{id}/reschedule [post]
func (h *AppointmentHandler) RescheduleAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.RescheduleAppointmentRequest) (any, error) {
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.rescheduleInternal(ctx, req, apptID.Value(), nil)
	}
	http.HandleRequestWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment rescheduled successfully")
	})(c)
}

// ConfirmAppointment godoc
// @Summary      Confirm appointment (receptionist/admin)
// @Description  Confirms an appointment and assigns the veterinarian given in the query param. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path   int  true  "Appointment ID"
// @Param        employee_id  query  int  true  "Employee ID to assign"
// @Success      200          {object}  http.APIResponse  "Appointment confirmed"
// @Failure      400          {object}  http.APIResponse  "Invalid ID or missing employee_id"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      404          {object}  http.APIResponse  "Appointment not found"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /appointments/{id}/confirm [post]
func (h *AppointmentHandler) ConfirmAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		// Manager passes employee_id as query: ?employee_id=1
		employeeID, err := EmployeeIDFromQuery(ctx)
		if err != nil {
			return nil, err
		}
		return h.confirmInternal(ctx, apptID.Value(), employeeID)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment confirmed successfully")
	})(c)
}

// CompleteAppointment godoc
// @Summary      Complete appointment (receptionist/admin)
// @Description  Marks an appointment as completed, with optional veterinarian and notes. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path   int     true   "Appointment ID"
// @Param        employee_id  query  int     false  "Employee ID"
// @Param        notes        query  string  false  "Completion notes"
// @Success      200          {object}  http.APIResponse  "Appointment completed"
// @Failure      400          {object}  http.APIResponse  "Invalid ID"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      404          {object}  http.APIResponse  "Appointment not found"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /appointments/{id}/complete [post]
func (h *AppointmentHandler) CompleteAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		notes := ctx.Query("notes")
		return h.completeInternal(ctx, apptID.Value(), nil, notes)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment completed successfully")
	})(c)
}

// CancelAppointment godoc
// @Summary      Cancel appointment (receptionist/admin)
// @Description  Cancels an appointment, with optional veterinarian and reason. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path   int     true   "Appointment ID"
// @Param        employee_id  query  int     false  "Employee ID"
// @Param        reason       query  string  false  "Cancellation reason"
// @Success      200          {object}  http.APIResponse  "Appointment cancelled"
// @Failure      400          {object}  http.APIResponse  "Invalid ID"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      404          {object}  http.APIResponse  "Appointment not found"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /appointments/{id}/cancel [post]
func (h *AppointmentHandler) CancelAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		reason := ctx.Query("reason")
		return h.cancelInternal(ctx, apptID.Value(), nil, nil, reason)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment cancelled successfully")
	})(c)
}

// NotAttendAppointment godoc
// @Summary      Mark appointment as not attended (receptionist/admin)
// @Description  Marks an appointment as not attended, with optional veterinarian. Requires admin or receptionist role.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path   int  true   "Appointment ID"
// @Param        employee_id  query  int  false  "Employee ID"
// @Success      200          {object}  http.APIResponse  "Appointment marked as not attended"
// @Failure      400          {object}  http.APIResponse  "Invalid ID"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      404          {object}  http.APIResponse  "Appointment not found"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /appointments/{id}/not-attend [post]
func (h *AppointmentHandler) NotAttendAppointment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		apptID, err := parseAppointmentIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return h.notAttendInternal(ctx, apptID.Value(), nil)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Appointment marked as not attended successfully")
	})(c)
}
