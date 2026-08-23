package handlers

import (
	"errors"

	"clinic-vet-api/internal/core/medical"
	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/middleware"
	"clinic-vet-api/internal/shared/http"
	"clinic-vet-api/internal/shared/page"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// MedicalHandler exposes medical session, vaccinations, surgeries, prescriptions,
// attachments, session services, and catalogs. Customers are read-only; admins have
// unrestricted access and veterinarians are limited to assigned sessions.
type MedicalHandler struct {
	sessionSvc           medical.MedicalSessionService
	vaccinationSvc       medical.VaccinationService
	surgerySvc           medical.SurgeryService
	prescriptionSvc      medical.PrescriptionService
	attachmentSvc        medical.AttachmentService
	sessionServiceSvc    medical.SessionServiceManager
	vaccineCatalogSvc    medical.VaccineCatalogService
	medicationCatalogSvc medical.MedicationCatalogService
	serviceCatalogSvc    medical.ServiceCatalogService
	validator            *validator.Validate
	customerIDResolver   CustomerIDResolver
	employeeIDResolver   EmployeeIDResolver
	petService           pets.Service // to verify pet belongs to customer
}

func NewMedicalHandler(
	sessionSvc medical.MedicalSessionService,
	vaccinationSvc medical.VaccinationService,
	surgerySvc medical.SurgeryService,
	prescriptionSvc medical.PrescriptionService,
	attachmentSvc medical.AttachmentService,
	sessionServiceSvc medical.SessionServiceManager,
	vaccineCatalogSvc medical.VaccineCatalogService,
	medicationCatalogSvc medical.MedicationCatalogService,
	serviceCatalogSvc medical.ServiceCatalogService,
	validator *validator.Validate,
	customerIDResolver CustomerIDResolver,
	employeeIDResolver EmployeeIDResolver,
	petService pets.Service,
) *MedicalHandler {
	return &MedicalHandler{
		sessionSvc:           sessionSvc,
		vaccinationSvc:       vaccinationSvc,
		surgerySvc:           surgerySvc,
		prescriptionSvc:      prescriptionSvc,
		attachmentSvc:        attachmentSvc,
		sessionServiceSvc:    sessionServiceSvc,
		vaccineCatalogSvc:    vaccineCatalogSvc,
		medicationCatalogSvc: medicationCatalogSvc,
		serviceCatalogSvc:    serviceCatalogSvc,
		validator:            validator,
		customerIDResolver:   customerIDResolver,
		employeeIDResolver:   employeeIDResolver,
		petService:           petService,
	}
}

func parseSessionID(c *gin.Context) (medical.SessionID, error) {
	id, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		return medical.SessionID{}, err
	}
	return medical.NewSessionID(id), nil
}

func (h *MedicalHandler) getMyCustomerID(c *gin.Context) (uint, error) {
	if h.customerIDResolver == nil {
		return 0, errors.New("customer resolver not configured")
	}
	return CustomerIDFromContext(c, h.customerIDResolver)
}

func (h *MedicalHandler) isAdmin(c *gin.Context) bool {
	user, ok := middleware.GetUserFromContext(c)
	return ok && user.Role == "admin"
}

func (h *MedicalHandler) getMyEmployeeID(c *gin.Context) (uint, error) {
	if h.employeeIDResolver == nil {
		return 0, errors.New("employee resolver not configured")
	}
	return EmployeeIDFromContext(c, h.employeeIDResolver)
}

func (h *MedicalHandler) ensureSessionAccessible(c *gin.Context, sessionID medical.SessionID) (medical.MedicalSession, error) {
	session, err := h.sessionSvc.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil || h.isAdmin(c) {
		return session, err
	}
	employeeID, err := h.getMyEmployeeID(c)
	if err != nil {
		return medical.MedicalSession{}, err
	}
	if session.EmployeeID != employeeID {
		return medical.MedicalSession{}, errors.New("forbidden: session is not assigned to veterinarian")
	}
	return session, nil
}

func (h *MedicalHandler) restrictToAssignedSessions(c *gin.Context, spec *medical.MedicalSessionSpecification) error {
	if h.isAdmin(c) {
		return nil
	}
	employeeID, err := h.getMyEmployeeID(c)
	if err != nil {
		return err
	}
	spec.EmployeeIDs = []uint{employeeID}
	return nil
}

func (h *MedicalHandler) assignedEmployeeIDs(c *gin.Context) ([]uint, error) {
	if h.isAdmin(c) {
		return nil, nil
	}
	employeeID, err := h.getMyEmployeeID(c)
	if err != nil {
		return nil, err
	}
	return []uint{employeeID}, nil
}

// ensureSessionBelongsToCustomer loads the session and returns an error if it does not belong to the customer.
func (h *MedicalHandler) ensureSessionBelongsToCustomer(c *gin.Context, sessionID medical.SessionID, customerID uint) (medical.MedicalSession, error) {
	sess, err := h.sessionSvc.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		return medical.MedicalSession{}, err
	}
	if sess.CustomerID != customerID {
		return medical.MedicalSession{}, errors.New("forbidden: session does not belong to customer")
	}
	return sess, nil
}

// ensurePetBelongsToCustomer returns nil if the pet belongs to the customer.
func (h *MedicalHandler) ensurePetBelongsToCustomer(c *gin.Context, petID uint, customerID uint) error {
	_, err := h.petService.GetPetByIDAndCustomerID(c.Request.Context(), pets.NewPetID(petID), customerID)
	return err
}

// ------------------------------------------------------------
// Customer read handlers (only their data / their pets’ data)
// ------------------------------------------------------------

// @Summary List my medical sessions
// @Description Requires customer role.
// @Tags medical
// @Security BearerAuth
// @Success 200 {object} http.APIResponse
// @Router /me/medical/sessions [get]
func (h *MedicalHandler) GetMySessions(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := h.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.sessionSvc.GetSessionsByCustomer(ctx.Request.Context(), customerID, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No sessions found")
			return
		}
		p := res.(page.Page[medical.MedicalSession])
		http.Paginated(c, &p, "Sessions")
	})(c)
}

// @Summary Get my medical session
// @Description Requires customer role.
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /me/medical/sessions/{id} [get]
func (h *MedicalHandler) GetMySessionByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := h.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		return h.ensureSessionBelongsToCustomer(ctx, sessionID, customerID)
	}
	http.HandleGetRequest(h.validator, "Session", logic)(c)
}

// @Summary Get my full medical session
// @Description Requires customer role.
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /me/medical/sessions/{id}/full [get]
func (h *MedicalHandler) GetMySessionFull(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := h.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		_, err = h.ensureSessionBelongsToCustomer(ctx, sessionID, customerID)
		if err != nil {
			return nil, err
		}
		return h.sessionSvc.GetSessionFull(ctx.Request.Context(), sessionID)
	}
	http.HandleGetRequest(h.validator, "Session", logic)(c)
}

// @Summary List my pet medical sessions
// @Description Requires customer role.
// @Tags medical
// @Security BearerAuth
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} http.APIResponse
// @Router /me/medical/pets/{pet_id}/sessions [get]
func (h *MedicalHandler) GetMyPetSessions(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := h.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		petID, err := http.ParseParamToUInt(ctx, "pet_id")
		if err != nil {
			return nil, err
		}
		if err := h.ensurePetBelongsToCustomer(ctx, petID, customerID); err != nil {
			return nil, err
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.sessionSvc.GetSessionsByPet(ctx.Request.Context(), petID, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No sessions found")
			return
		}
		p := res.(page.Page[medical.MedicalSession])
		http.Paginated(c, &p, "Sessions")
	})(c)
}

// @Summary Get my pet vaccination summary
// @Description Requires customer role.
// @Tags medical
// @Security BearerAuth
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} http.APIResponse
// @Router /me/medical/pets/{pet_id}/vaccination-summary [get]
func (h *MedicalHandler) GetMyPetVaccinationSummary(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := h.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		petID, err := http.ParseParamToUInt(ctx, "pet_id")
		if err != nil {
			return nil, err
		}
		if err := h.ensurePetBelongsToCustomer(ctx, petID, customerID); err != nil {
			return nil, err
		}
		return h.vaccinationSvc.GetPetVaccinationSummary(ctx.Request.Context(), petID, nil)
	}
	http.HandleGetRequest(h.validator, "VaccinationSummary", logic)(c)
}

// @Summary List my pet vaccination history
// @Description Requires customer role.
// @Tags medical
// @Security BearerAuth
// @Param pet_id path int true "Pet ID"
// @Param vaccine_id query int false "Vaccine ID"
// @Param date_from query string false "Start date"
// @Param date_to query string false "End date"
// @Success 200 {object} http.APIResponse
// @Router /me/medical/pets/{pet_id}/vaccination-history [get]
func (h *MedicalHandler) GetMyPetVaccinationHistory(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := h.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		petID, err := http.ParseParamToUInt(ctx, "pet_id")
		if err != nil {
			return nil, err
		}
		if err := h.ensurePetBelongsToCustomer(ctx, petID, customerID); err != nil {
			return nil, err
		}
		var req dtos.VaccinationHistoryRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			return nil, err
		}
		req.PetID = petID
		spec := mappers.VaccinationHistoryRequestToSpec(req)
		return h.vaccinationSvc.GetVaccinationHistory(ctx.Request.Context(), *spec)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No history found")
			return
		}
		p := res.(page.Page[medical.SessionVaccination])
		http.Paginated(c, &p, "VaccinationHistory")
	})(c)
}

// @Summary List my pet active prescriptions
// @Description Requires customer role.
// @Tags medical
// @Security BearerAuth
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} http.APIResponse
// @Router /me/medical/pets/{pet_id}/prescriptions/active [get]
func (h *MedicalHandler) GetMyPetActivePrescriptions(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := h.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		petID, err := http.ParseParamToUInt(ctx, "pet_id")
		if err != nil {
			return nil, err
		}
		if err := h.ensurePetBelongsToCustomer(ctx, petID, customerID); err != nil {
			return nil, err
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.prescriptionSvc.GetActivePrescriptionsByPet(ctx.Request.Context(), petID, nil, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No prescriptions found")
			return
		}
		p := res.(page.Page[medical.SessionPrescription])
		http.Paginated(c, &p, "Prescriptions")
	})(c)
}

// ------------------------------------------------------------
// Staff read handlers
// ------------------------------------------------------------

// @Summary Get medical session
// @Description Requires veterinarian or admin role.
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id} [get]
func (h *MedicalHandler) GetSessionByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		return h.ensureSessionAccessible(ctx, sessionID)
	}
	http.HandleGetRequest(h.validator, "Session", logic)(c)
}

// @Summary Get full medical session
// @Description Requires veterinarian or admin role.
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id}/full [get]
func (h *MedicalHandler) GetSessionFull(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		return h.sessionSvc.GetSessionFull(ctx.Request.Context(), sessionID)
	}
	http.HandleGetRequest(h.validator, "Session", logic)(c)
}

// @Summary Search medical sessions
// @Description Requires veterinarian or admin role.
// @Tags medical
// @Security BearerAuth
// @Param pet_id query int false "Pet ID"
// @Param customer_id query int false "Customer ID"
// @Param employee_id query int false "Employee ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions [get]
func (h *MedicalHandler) GetSessionsBySpecification(c *gin.Context) {
	var req dtos.SessionSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		http.BadRequest(c, err)
		return
	}
	spec, err := mappers.SessionSearchToSpec(req)
	if err != nil {
		http.BadRequest(c, err)
		return
	}
	logic := func(ctx *gin.Context) (any, error) {
		if err := h.restrictToAssignedSessions(ctx, spec); err != nil {
			return nil, err
		}
		return h.sessionSvc.GetSessionsBySpecification(ctx.Request.Context(), *spec)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No sessions found")
			return
		}
		p := res.(page.Page[medical.MedicalSession])
		http.Paginated(c, &p, "Sessions")
	})(c)
}

// @Summary List pet medical sessions
// @Description Requires veterinarian or admin role.
// @Tags medical
// @Security BearerAuth
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/pets/{pet_id}/sessions [get]
func (h *MedicalHandler) GetSessionsByPet(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		petID, err := http.ParseParamToUInt(ctx, "pet_id")
		if err != nil {
			return nil, err
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		spec := &medical.MedicalSessionSpecification{PetIDs: []uint{petID}, Pagination: pageParams.ToPagination()}
		if err := h.restrictToAssignedSessions(ctx, spec); err != nil {
			return nil, err
		}
		return h.sessionSvc.GetSessionsBySpecification(ctx.Request.Context(), *spec)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No sessions found")
			return
		}
		p := res.(page.Page[medical.MedicalSession])
		http.Paginated(c, &p, "Sessions")
	})(c)
}

// @Summary List customer medical sessions
// @Description Requires veterinarian or admin role.
// @Tags medical
// @Security BearerAuth
// @Param customer_id path int true "Customer ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/customers/{customer_id}/sessions [get]
func (h *MedicalHandler) GetSessionsByCustomer(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := CustomerIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		spec := &medical.MedicalSessionSpecification{CustomerIDs: []uint{customerID}, Pagination: pageParams.ToPagination()}
		if err := h.restrictToAssignedSessions(ctx, spec); err != nil {
			return nil, err
		}
		return h.sessionSvc.GetSessionsBySpecification(ctx.Request.Context(), *spec)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No sessions found")
			return
		}
		p := res.(page.Page[medical.MedicalSession])
		http.Paginated(c, &p, "Sessions")
	})(c)
}

// @Summary Get medical session statistics
// @Description Requires veterinarian or admin role.
// @Tags medical
// @Security BearerAuth
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/stats [get]
func (h *MedicalHandler) GetSessionStats(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		return h.sessionSvc.GetSessionStats(ctx.Request.Context())
	}
	http.HandleGetRequest(h.validator, "SessionStats", logic)(c)
}

// @Summary List session vaccinations
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id}/vaccinations [get]
func (h *MedicalHandler) GetVaccinationsBySession(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		return h.vaccinationSvc.GetVaccinationsBySession(ctx.Request.Context(), sessionID)
	}
	http.HandleGetRequest(h.validator, "Vaccinations", logic)(c)
}

// @Summary List session surgeries
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id}/surgeries [get]
func (h *MedicalHandler) GetSurgeriesBySession(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		return h.surgerySvc.GetSurgeriesBySession(ctx.Request.Context(), sessionID)
	}
	http.HandleGetRequest(h.validator, "Surgeries", logic)(c)
}

// @Summary List session prescriptions
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id}/prescriptions [get]
func (h *MedicalHandler) GetPrescriptionsBySession(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		return h.prescriptionSvc.GetPrescriptionsBySession(ctx.Request.Context(), sessionID)
	}
	http.HandleGetRequest(h.validator, "Prescriptions", logic)(c)
}

// @Summary List session attachments
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id}/attachments [get]
func (h *MedicalHandler) GetAttachmentsBySession(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		return h.attachmentSvc.GetAttachmentsBySession(ctx.Request.Context(), sessionID)
	}
	http.HandleGetRequest(h.validator, "Attachments", logic)(c)
}

// @Summary List session services
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id}/services [get]
func (h *MedicalHandler) GetServicesBySession(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		return h.sessionServiceSvc.GetServicesBySession(ctx.Request.Context(), sessionID)
	}
	http.HandleGetRequest(h.validator, "SessionServices", logic)(c)
}

// @Summary Search vaccination history
// @Tags medical
// @Security BearerAuth
// @Param pet_id query int false "Pet ID"
// @Param vaccine_id query int false "Vaccine ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/vaccination-history [get]
func (h *MedicalHandler) GetVaccinationHistory(c *gin.Context) {
	var req dtos.VaccinationHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		http.BadRequest(c, err)
		return
	}
	spec := mappers.VaccinationHistoryRequestToSpec(req)
	logic := func(ctx *gin.Context) (any, error) {
		employeeIDs, err := h.assignedEmployeeIDs(ctx)
		if err != nil {
			return nil, err
		}
		spec.EmployeeIDs = employeeIDs
		return h.vaccinationSvc.GetVaccinationHistory(ctx.Request.Context(), *spec)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No history found")
			return
		}
		p := res.(page.Page[medical.SessionVaccination])
		http.Paginated(c, &p, "VaccinationHistory")
	})(c)
}

// @Summary Get pet vaccination summary
// @Tags medical
// @Security BearerAuth
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/pets/{pet_id}/vaccination-summary [get]
func (h *MedicalHandler) GetPetVaccinationSummary(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		petID, err := http.ParseParamToUInt(ctx, "pet_id")
		if err != nil {
			return nil, err
		}
		employeeIDs, err := h.assignedEmployeeIDs(ctx)
		if err != nil {
			return nil, err
		}
		return h.vaccinationSvc.GetPetVaccinationSummary(ctx.Request.Context(), petID, employeeIDs)
	}
	http.HandleGetRequest(h.validator, "VaccinationSummary", logic)(c)
}

// @Summary List active pet prescriptions
// @Tags medical
// @Security BearerAuth
// @Param pet_id path int true "Pet ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/pets/{pet_id}/prescriptions/active [get]
func (h *MedicalHandler) GetActivePrescriptionsByPet(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		petID, err := http.ParseParamToUInt(ctx, "pet_id")
		if err != nil {
			return nil, err
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		employeeIDs, err := h.assignedEmployeeIDs(ctx)
		if err != nil {
			return nil, err
		}
		return h.prescriptionSvc.GetActivePrescriptionsByPet(ctx.Request.Context(), petID, employeeIDs, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No prescriptions found")
			return
		}
		p := res.(page.Page[medical.SessionPrescription])
		http.Paginated(c, &p, "Prescriptions")
	})(c)
}

// Catalog read
// @Summary List vaccine catalog
// @Tags medical
// @Security BearerAuth
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/vaccines [get]
func (h *MedicalHandler) ListVaccines(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.vaccineCatalogSvc.ListVaccines(ctx.Request.Context(), pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No vaccines found")
			return
		}
		p := res.(page.Page[medical.VaccineCatalog])
		http.Paginated(c, &p, "Vaccines")
	})(c)
}

// @Summary Get vaccine catalog entry
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Vaccine ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/vaccines/{id} [get]
func (h *MedicalHandler) GetVaccineByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "id")
		if err != nil {
			return nil, err
		}
		return h.vaccineCatalogSvc.GetVaccineByID(ctx.Request.Context(), medical.NewVaccineCatalogID(id))
	}
	http.HandleGetRequest(h.validator, "Vaccine", logic)(c)
}

// @Summary List vaccines by species
// @Tags medical
// @Security BearerAuth
// @Param species path string true "Species"
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/vaccines/species/{species} [get]
func (h *MedicalHandler) ListVaccinesBySpecies(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		species := c.Param("species")
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.vaccineCatalogSvc.ListVaccinesBySpecies(ctx.Request.Context(), species, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No vaccines found")
			return
		}
		p := res.(page.Page[medical.VaccineCatalog])
		http.Paginated(c, &p, "Vaccines")
	})(c)
}

// @Summary List medication catalog
// @Tags medical
// @Security BearerAuth
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/medications [get]
func (h *MedicalHandler) ListMedications(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.medicationCatalogSvc.ListMedications(ctx.Request.Context(), pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No medications found")
			return
		}
		p := res.(page.Page[medical.Medication])
		http.Paginated(c, &p, "Medications")
	})(c)
}

// @Summary Search medication catalog
// @Tags medical
// @Security BearerAuth
// @Param term query string false "Search term"
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/medications/search [get]
func (h *MedicalHandler) SearchMedications(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		term := c.Query("term")
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.medicationCatalogSvc.SearchMedications(ctx.Request.Context(), term, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No medications found")
			return
		}
		p := res.(page.Page[medical.Medication])
		http.Paginated(c, &p, "Medications")
	})(c)
}

// @Summary Get medication catalog entry
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Medication ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/medications/{id} [get]
func (h *MedicalHandler) GetMedicationByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "id")
		if err != nil {
			return nil, err
		}
		return h.medicationCatalogSvc.GetMedicationByID(ctx.Request.Context(), medical.NewMedicationID(id))
	}
	http.HandleGetRequest(h.validator, "Medication", logic)(c)
}

// @Summary List service catalog
// @Tags medical
// @Security BearerAuth
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/services [get]
func (h *MedicalHandler) ListServices(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.serviceCatalogSvc.ListServices(ctx.Request.Context(), pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No services found")
			return
		}
		p := res.(page.Page[medical.ServiceCatalog])
		http.Paginated(c, &p, "Services")
	})(c)
}

// @Summary Get service catalog entry
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/services/{id} [get]
func (h *MedicalHandler) GetServiceByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "id")
		if err != nil {
			return nil, err
		}
		return h.serviceCatalogSvc.GetServiceByID(ctx.Request.Context(), medical.NewServiceCatalogID(id))
	}
	http.HandleGetRequest(h.validator, "Service", logic)(c)
}

// @Summary List services by category
// @Tags medical
// @Security BearerAuth
// @Param category path string true "Service category"
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/services/category/{category} [get]
func (h *MedicalHandler) ListServicesByCategory(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		catStr := c.Param("category")
		cat, err := medical.ParseServiceCategory(catStr)
		if err != nil {
			return nil, err
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.serviceCatalogSvc.ListServicesByCategory(ctx.Request.Context(), cat, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No services found")
			return
		}
		p := res.(page.Page[medical.ServiceCatalog])
		http.Paginated(c, &p, "Services")
	})(c)
}

// ------------------------------------------------------------
// Staff write handlers
// ------------------------------------------------------------

// @Summary Create medical session
// @Tags medical
// @Security BearerAuth
// @Param body body dtos.CreateSessionRequest true "Session"
// @Success 201 {object} http.APIResponse
// @Router /medical/sessions [post]
func (h *MedicalHandler) CreateSession(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.CreateSessionRequest) (any, error) {
		cmd, err := mappers.CreateSessionRequestToCommand(req)
		if err != nil {
			return nil, err
		}
		if !h.isAdmin(ctx) {
			employeeID, err := h.getMyEmployeeID(ctx)
			if err != nil {
				return nil, err
			}
			cmd.EmployeeID = employeeID
		}
		sess, err := h.sessionSvc.CreateSession(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return sess.ID.Value(), nil
	}
	http.HandleCreateRequest(h.validator, "Session", logic)(c)
}

// @Summary Update medical session
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Param body body dtos.UpdateSessionRequest true "Session"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id} [put]
func (h *MedicalHandler) UpdateSession(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.UpdateSessionRequest) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		cmd, err := mappers.UpdateSessionRequestToCommand(sessionID.Value(), req)
		if err != nil {
			return nil, err
		}
		if !h.isAdmin(ctx) {
			employeeID, err := h.getMyEmployeeID(ctx)
			if err != nil {
				return nil, err
			}
			cmd.EmployeeID = &employeeID
		}
		_, err = h.sessionSvc.UpdateSession(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	http.HandleUpdateRequest(h.validator, "Session", logic)(c)
}

// @Summary Delete medical session
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id} [delete]
func (h *MedicalHandler) SoftDeleteSession(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		return nil, h.sessionSvc.SoftDeleteSession(ctx.Request.Context(), sessionID)
	}
	http.HandleDeleteRequest(h.validator, "Session", logic)(c)
}

// @Summary Permanently delete medical session
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id}/hard [delete]
func (h *MedicalHandler) HardDeleteSession(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		return nil, h.sessionSvc.HardDeleteSession(ctx.Request.Context(), sessionID)
	}
	http.HandleDeleteRequest(h.validator, "Session", logic)(c)
}

// @Summary Restore medical session
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/sessions/{id}/restore [post]
func (h *MedicalHandler) RestoreSession(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		sessionID, err := parseSessionID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionID); err != nil {
			return nil, err
		}
		return nil, h.sessionSvc.RestoreSession(ctx.Request.Context(), sessionID)
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Updated(c, nil, "Session")
	})(c)
}

// @Summary Add vaccination
// @Tags medical
// @Security BearerAuth
// @Param body body dtos.AddVaccinationRequest true "Vaccination"
// @Success 201 {object} http.APIResponse
// @Router /medical/vaccinations [post]
func (h *MedicalHandler) AddVaccination(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AddVaccinationRequest) (any, error) {
		cmd, err := mappers.AddVaccinationRequestToCommand(req)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, cmd.SessionID); err != nil {
			return nil, err
		}
		v, err := h.vaccinationSvc.AddVaccination(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return v.ID.Value(), nil
	}
	http.HandleCreateRequest(h.validator, "Vaccination", logic)(c)
}

// @Summary Update vaccination
// @Tags medical
// @Security BearerAuth
// @Param vaccination_id path int true "Vaccination ID"
// @Param body body dtos.UpdateVaccinationRequest true "Vaccination"
// @Success 200 {object} http.APIResponse
// @Router /medical/vaccinations/{vaccination_id} [put]
func (h *MedicalHandler) UpdateVaccination(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.UpdateVaccinationRequest) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "vaccination_id")
		if err != nil {
			return nil, err
		}
		cmd, err := mappers.UpdateVaccinationRequestToCommand(id, req)
		if err != nil {
			return nil, err
		}
		vaccination, err := h.vaccinationSvc.GetVaccinationByID(ctx.Request.Context(), medical.NewVaccinationID(id))
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, vaccination.SessionID); err != nil {
			return nil, err
		}
		_, err = h.vaccinationSvc.UpdateVaccination(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	http.HandleUpdateRequest(h.validator, "Vaccination", logic)(c)
}

// @Summary Remove vaccination
// @Tags medical
// @Security BearerAuth
// @Param vaccination_id path int true "Vaccination ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/vaccinations/{vaccination_id} [delete]
func (h *MedicalHandler) RemoveVaccination(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "vaccination_id")
		if err != nil {
			return nil, err
		}
		vaccination, err := h.vaccinationSvc.GetVaccinationByID(ctx.Request.Context(), medical.NewVaccinationID(id))
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, vaccination.SessionID); err != nil {
			return nil, err
		}
		return nil, h.vaccinationSvc.RemoveVaccination(ctx.Request.Context(), medical.NewVaccinationID(id))
	}
	http.HandleDeleteRequest(h.validator, "Vaccination", logic)(c)
}

// @Summary Add surgery
// @Tags medical
// @Security BearerAuth
// @Param body body dtos.AddSurgeryRequest true "Surgery"
// @Success 201 {object} http.APIResponse
// @Router /medical/surgeries [post]
func (h *MedicalHandler) AddSurgery(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AddSurgeryRequest) (any, error) {
		cmd, err := mappers.AddSurgeryRequestToCommand(req)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, cmd.SessionID); err != nil {
			return nil, err
		}
		s, err := h.surgerySvc.AddSurgery(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return s.ID.Value(), nil
	}
	http.HandleCreateRequest(h.validator, "Surgery", logic)(c)
}

// @Summary Update surgery
// @Tags medical
// @Security BearerAuth
// @Param surgery_id path int true "Surgery ID"
// @Param body body dtos.UpdateSurgeryRequest true "Surgery"
// @Success 200 {object} http.APIResponse
// @Router /medical/surgeries/{surgery_id} [put]
func (h *MedicalHandler) UpdateSurgery(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.UpdateSurgeryRequest) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "surgery_id")
		if err != nil {
			return nil, err
		}
		cmd, err := mappers.UpdateSurgeryRequestToCommand(id, req)
		if err != nil {
			return nil, err
		}
		surgery, err := h.surgerySvc.GetSurgeryByID(ctx.Request.Context(), medical.NewSurgeryID(id))
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, surgery.SessionID); err != nil {
			return nil, err
		}
		_, err = h.surgerySvc.UpdateSurgery(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	http.HandleUpdateRequest(h.validator, "Surgery", logic)(c)
}

// @Summary Remove surgery
// @Tags medical
// @Security BearerAuth
// @Param surgery_id path int true "Surgery ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/surgeries/{surgery_id} [delete]
func (h *MedicalHandler) RemoveSurgery(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "surgery_id")
		if err != nil {
			return nil, err
		}
		surgery, err := h.surgerySvc.GetSurgeryByID(ctx.Request.Context(), medical.NewSurgeryID(id))
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, surgery.SessionID); err != nil {
			return nil, err
		}
		return nil, h.surgerySvc.RemoveSurgery(ctx.Request.Context(), medical.NewSurgeryID(id))
	}
	http.HandleDeleteRequest(h.validator, "Surgery", logic)(c)
}

// @Summary Add prescription
// @Tags medical
// @Security BearerAuth
// @Param body body dtos.AddPrescriptionRequest true "Prescription"
// @Success 201 {object} http.APIResponse
// @Router /medical/prescriptions [post]
func (h *MedicalHandler) AddPrescription(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AddPrescriptionRequest) (any, error) {
		cmd, err := mappers.AddPrescriptionRequestToCommand(req)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, cmd.SessionID); err != nil {
			return nil, err
		}
		p, err := h.prescriptionSvc.AddPrescription(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return p.ID.Value(), nil
	}
	http.HandleCreateRequest(h.validator, "Prescription", logic)(c)
}

// @Summary Update prescription
// @Tags medical
// @Security BearerAuth
// @Param prescription_id path int true "Prescription ID"
// @Param body body dtos.UpdatePrescriptionRequest true "Prescription"
// @Success 200 {object} http.APIResponse
// @Router /medical/prescriptions/{prescription_id} [put]
func (h *MedicalHandler) UpdatePrescription(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.UpdatePrescriptionRequest) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "prescription_id")
		if err != nil {
			return nil, err
		}
		cmd := mappers.UpdatePrescriptionRequestToCommand(id, req)
		prescription, err := h.prescriptionSvc.GetPrescriptionByID(ctx.Request.Context(), medical.NewPrescriptionID(id))
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, prescription.SessionID); err != nil {
			return nil, err
		}
		_, err = h.prescriptionSvc.UpdatePrescription(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	http.HandleUpdateRequest(h.validator, "Prescription", logic)(c)
}

// @Summary Remove prescription
// @Tags medical
// @Security BearerAuth
// @Param prescription_id path int true "Prescription ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/prescriptions/{prescription_id} [delete]
func (h *MedicalHandler) RemovePrescription(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "prescription_id")
		if err != nil {
			return nil, err
		}
		prescription, err := h.prescriptionSvc.GetPrescriptionByID(ctx.Request.Context(), medical.NewPrescriptionID(id))
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, prescription.SessionID); err != nil {
			return nil, err
		}
		return nil, h.prescriptionSvc.RemovePrescription(ctx.Request.Context(), medical.NewPrescriptionID(id))
	}
	http.HandleDeleteRequest(h.validator, "Prescription", logic)(c)
}

// @Summary Add attachment
// @Tags medical
// @Security BearerAuth
// @Param body body dtos.AddAttachmentRequest true "Attachment"
// @Success 201 {object} http.APIResponse
// @Router /medical/attachments [post]
func (h *MedicalHandler) AddAttachment(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AddAttachmentRequest) (any, error) {
		cmd, err := mappers.AddAttachmentRequestToCommand(req)
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, cmd.SessionID); err != nil {
			return nil, err
		}
		a, err := h.attachmentSvc.AddAttachment(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return a.ID.Value(), nil
	}
	http.HandleCreateRequest(h.validator, "Attachment", logic)(c)
}

// @Summary Remove attachment
// @Tags medical
// @Security BearerAuth
// @Param attachment_id path int true "Attachment ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/attachments/{attachment_id} [delete]
func (h *MedicalHandler) RemoveAttachment(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "attachment_id")
		if err != nil {
			return nil, err
		}
		attachment, err := h.attachmentSvc.GetAttachmentByID(ctx.Request.Context(), medical.NewAttachmentID(id))
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, attachment.SessionID); err != nil {
			return nil, err
		}
		return nil, h.attachmentSvc.RemoveAttachment(ctx.Request.Context(), medical.NewAttachmentID(id))
	}
	http.HandleDeleteRequest(h.validator, "Attachment", logic)(c)
}

// @Summary Add session service
// @Tags medical
// @Security BearerAuth
// @Param body body dtos.AddSessionServiceRequest true "Session service"
// @Success 201 {object} http.APIResponse
// @Router /medical/session-services [post]
func (h *MedicalHandler) AddSessionService(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AddSessionServiceRequest) (any, error) {
		cmd := mappers.AddSessionServiceRequestToCommand(req)
		if _, err := h.ensureSessionAccessible(ctx, cmd.SessionID); err != nil {
			return nil, err
		}
		s, err := h.sessionServiceSvc.AddService(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return s.ID.Value(), nil
	}
	http.HandleCreateRequest(h.validator, "SessionService", logic)(c)
}

// @Summary Remove session service
// @Tags medical
// @Security BearerAuth
// @Param session_service_id path int true "Session service ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/session-services/{session_service_id} [delete]
func (h *MedicalHandler) RemoveSessionService(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "session_service_id")
		if err != nil {
			return nil, err
		}
		sessionService, err := h.sessionServiceSvc.GetSessionServiceByID(ctx.Request.Context(), medical.NewSessionServiceID(id))
		if err != nil {
			return nil, err
		}
		if _, err := h.ensureSessionAccessible(ctx, sessionService.SessionID); err != nil {
			return nil, err
		}
		return nil, h.sessionServiceSvc.RemoveService(ctx.Request.Context(), medical.NewSessionServiceID(id))
	}
	http.HandleDeleteRequest(h.validator, "SessionService", logic)(c)
}

// @Summary Create vaccine catalog entry
// @Tags medical
// @Security BearerAuth
// @Param body body dtos.CreateVaccineCatalogRequest true "Vaccine"
// @Success 201 {object} http.APIResponse
// @Router /medical/catalogs/vaccines [post]
func (h *MedicalHandler) CreateVaccineCatalog(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.CreateVaccineCatalogRequest) (any, error) {
		cmd := mappers.CreateVaccineCatalogRequestToCommand(req)
		v, err := h.vaccineCatalogSvc.CreateVaccine(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return v.ID.Value(), nil
	}
	http.HandleCreateRequest(h.validator, "Vaccine", logic)(c)
}

// @Summary Deactivate vaccine catalog entry
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Vaccine ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/vaccines/{id} [delete]
func (h *MedicalHandler) DeactivateVaccine(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "id")
		if err != nil {
			return nil, err
		}
		return nil, h.vaccineCatalogSvc.DeactivateVaccine(ctx.Request.Context(), medical.NewVaccineCatalogID(id))
	}
	http.HandleDeleteRequest(h.validator, "Vaccine", logic)(c)
}

// @Summary Create medication catalog entry
// @Tags medical
// @Security BearerAuth
// @Param body body dtos.CreateMedicationCatalogRequest true "Medication"
// @Success 201 {object} http.APIResponse
// @Router /medical/catalogs/medications [post]
func (h *MedicalHandler) CreateMedicationCatalog(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.CreateMedicationCatalogRequest) (any, error) {
		cmd := mappers.CreateMedicationCatalogRequestToCommand(req)
		m, err := h.medicationCatalogSvc.CreateMedication(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return m.ID.Value(), nil
	}
	http.HandleCreateRequest(h.validator, "Medication", logic)(c)
}

// @Summary Deactivate medication catalog entry
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Medication ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/medications/{id} [delete]
func (h *MedicalHandler) DeactivateMedication(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "id")
		if err != nil {
			return nil, err
		}
		return nil, h.medicationCatalogSvc.DeactivateMedication(ctx.Request.Context(), medical.NewMedicationID(id))
	}
	http.HandleDeleteRequest(h.validator, "Medication", logic)(c)
}

// @Summary Create service catalog entry
// @Tags medical
// @Security BearerAuth
// @Param body body dtos.CreateServiceCatalogRequest true "Service"
// @Success 201 {object} http.APIResponse
// @Router /medical/catalogs/services [post]
func (h *MedicalHandler) CreateServiceCatalog(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.CreateServiceCatalogRequest) (any, error) {
		cmd, err := mappers.CreateServiceCatalogRequestToCommand(req)
		if err != nil {
			return nil, err
		}
		s, err := h.serviceCatalogSvc.CreateService(ctx.Request.Context(), cmd)
		if err != nil {
			return nil, err
		}
		return s.ID.Value(), nil
	}
	http.HandleCreateRequest(h.validator, "Service", logic)(c)
}

// @Summary Deactivate service catalog entry
// @Tags medical
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Success 200 {object} http.APIResponse
// @Router /medical/catalogs/services/{id} [delete]
func (h *MedicalHandler) DeactivateServiceCatalog(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id, err := http.ParseParamToUInt(ctx, "id")
		if err != nil {
			return nil, err
		}
		return nil, h.serviceCatalogSvc.DeactivateService(ctx.Request.Context(), medical.NewServiceCatalogID(id))
	}
	http.HandleDeleteRequest(h.validator, "Service", logic)(c)
}
