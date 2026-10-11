package handlers

import (
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/shared/http"
	"clinic-vet-api/internal/shared/page"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type EmployeeHandler struct {
	service   employees.EmployeeService
	validator *validator.Validate
	mapper    *mappers.EmployeeMapper
}

func NewEmployeeHandler(service employees.EmployeeService, validator *validator.Validate) *EmployeeHandler {
	return &EmployeeHandler{
		service:   service,
		validator: validator,
		mapper:    mappers.NewEmployeeMapper(),
	}
}

// GetEmployeeByID godoc
// @Summary      Get employee by ID
// @Description  Returns a single employee by ID. Requires admin role.
// @Tags         employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Employee ID"
// @Success      200  {object}  http.APIResponse{data=dtos.EmployeeResponse}  "Employee found"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      404  {object}  http.APIResponse  "Employee not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /employees/{id} [get]
func (s *EmployeeHandler) GetEmployeeByID(c *gin.Context) {
	employeeIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	employeeID := employees.NewEmployeeID(employeeIDUInt)
	employee, err := s.service.GetEmployeeByID(c.Request.Context(), employeeID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	employeeResponse := s.mapper.ToEmployeeResponse(employee)
	http.Found(c, employeeResponse, "Employee")
}

// GetActiveEmployees godoc
// @Summary      List active employees
// @Description  Returns active employees with pagination. Requires admin role.
// @Tags         employees
// @Produce      json
// @Security     BearerAuth
// @Param        page_number  query     int  false  "Page number"
// @Param        page_size    query     int  false  "Page size"
// @Success      200          {object}  http.APIResponse{data=[]dtos.EmployeeResponse}  "Active employees"
// @Failure      400          {object}  http.APIResponse  "Invalid pagination"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Router       /employees/active [get]
func (s *EmployeeHandler) GetActiveEmployees(c *gin.Context) {
	var requestBodyData dtos.EmployeeSearchRequest
	if err := http.ShouldBindAndValidateQuery(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	pagination := requestBodyData.PaginationRequest.ToPagination()

	employeePage, err := s.service.GetActiveEmployees(c.Request.Context(), pagination)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	employeeResponses := s.mapper.ToEmployeeResponses(employeePage)
	http.Paginated(c, &employeeResponses, "Employees")
}

// GetEmployeesBySpecialty godoc
// @Summary      List employees by specialty
// @Description  Returns employees with the requested specialty. Requires admin role.
// @Tags         employees
// @Produce      json
// @Security     BearerAuth
// @Param        specialty    path      string  true   "Veterinarian specialty"
// @Param        page_number  query     int     false  "Page number"
// @Param        page_size    query     int     false  "Page size"
// @Success      200          {object}  http.APIResponse{data=[]dtos.EmployeeResponse}  "Employees"
// @Failure      400          {object}  http.APIResponse  "Invalid specialty or pagination"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Router       /employees/specialty/{specialty} [get]
func (s *EmployeeHandler) GetEmployeesBySpecialty(c *gin.Context) {
	var pageRequest page.PaginationRequest
	if err := http.ShouldBindPageParams(&pageRequest, c, s.validator); err != nil {
		if err != nil {
			http.BadRequest(c, err)
			return
		}
	}
	specialty := c.Param("specialty")
	specialtyEnum, err := employees.ParseVetSpecialty(specialty)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	employeesPage, err := s.service.GetEmployeesBySpecialty(c.Request.Context(), specialtyEnum, pageRequest.ToPagination())
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	employeeResponses := s.mapper.ToEmployeeResponses(employeesPage)
	http.Paginated(c, &employeeResponses, "Employees")
}

// GetEmployeeStats godoc
// @Summary      Get employee statistics
// @Description  Returns aggregate employee statistics. Requires admin role.
// @Tags         employees
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  http.APIResponse{data=dtos.EmployeeStatsResponse}  "Employee statistics"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Router       /employees/stats [get]
func (s *EmployeeHandler) GetEmployeeStats(c *gin.Context) {
	stats, err := s.service.GetEmployeeStats(c.Request.Context())
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	employeeStatsResponse := s.mapper.ToEmployeeStatsResponse(stats)
	http.Found(c, employeeStatsResponse, "Employee Stats")
}

// SearchEmployees godoc
// @Summary      Search employees
// @Description  Paginated search/filter of employees. Requires admin role.
// @Tags         employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.EmployeeSearchRequest  true  "Search criteria and pagination"
// @Success      200     {object}  http.APIResponse{data=[]dtos.EmployeeResponse}  "Paginated list of employees"
// @Failure      400     {object}  http.APIResponse           "Validation error"
// @Failure      401     {object}  http.APIResponse           "Unauthorized"
// @Failure      500     {object}  http.APIResponse           "Internal server error"
// @Router       /employees [get]
func (s *EmployeeHandler) SearchEmployees(c *gin.Context) {
	var requestBodyData dtos.EmployeeSearchRequest
	if err := http.ShouldBindAndValidateQuery(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}
	spec := employees.EmployeeSpecification{Pagination: requestBodyData.PaginationRequest.WithDefaults().ToPagination(), IsActive: requestBodyData.IsActive}
	if requestBodyData.Specialty != "" {
		specialty, err := employees.ParseVetSpecialty(requestBodyData.Specialty)
		if err != nil {
			http.BadRequest(c, err)
			return
		}
		spec.Specialty = &specialty
	}
	employeePage, err := s.service.GetEmployeesBySpecification(c.Request.Context(), spec)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	responsePage := s.mapper.ToEmployeeResponses(employeePage)
	http.Paginated(c, &responsePage, "Employees")
}

// CreateEmployee godoc
// @Summary      Create employee
// @Description  Creates a new employee. Requires admin role.
// @Tags         employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.EmployeeCreateRequest  true  "Employee data"
// @Success      201    {object}  http.APIResponse             "Employee created"
// @Failure      400    {object}  http.APIResponse             "Validation error"
// @Failure      401    {object}  http.APIResponse             "Unauthorized"
// @Failure      500    {object}  http.APIResponse             "Internal server error"
// @Router       /employees [post]
func (s *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var requestBodyData dtos.EmployeeCreateRequest
	if err := http.ShouldBindAndValidateBody(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	cmd, err := s.mapper.RequestToCreateCommand(requestBodyData)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	employee, err := s.service.CreateEmployee(c.Request.Context(), cmd)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Created(c, employee.ID.Value(), "Employee")
}

// UpdateEmployee godoc
// @Summary      Update employee
// @Description  Updates an existing employee by ID. Requires admin role.
// @Tags         employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.EmployeeUpdateRequest  true  "Employee data to update"
// @Success      200     {object}  http.APIResponse            "Employee updated"
// @Failure      400     {object}  http.APIResponse            "Validation error"
// @Failure      401     {object}  http.APIResponse            "Unauthorized"
// @Failure      500     {object}  http.APIResponse            "Internal server error"
// @Router       /employees/{id} [put]
func (s *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	pathID, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, err)
		return
	}
	var requestBodyData dtos.EmployeeUpdateRequest
	if err := http.ShouldBindAndValidateBody(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}
	if requestBodyData.ID != pathID {
		http.BadRequest(c, fmt.Errorf("body id must match path id"))
		return
	}

	cmd, err := s.mapper.RequestToUpdateCommand(requestBodyData)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	if err := s.service.UpdateEmployee(c.Request.Context(), cmd); err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Updated(c, nil, "Employee")
}

// RestoreEmployee godoc
// @Summary      Restore employee
// @Description  Restores a soft-deleted employee. Requires admin role.
// @Tags         employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Employee ID"
// @Success      200  {object}  http.APIResponse  "Employee restored"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /employees/{id}/restore [post]
func (s *EmployeeHandler) RestoreEmployee(c *gin.Context) {
	employeeIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	employeeID := employees.NewEmployeeID(employeeIDUInt)
	err = s.service.RestoreEmployee(c.Request.Context(), employeeID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	http.Success(c, nil, "Employee restored")
}

// DeleteEmployee godoc
// @Summary      Delete employee
// @Description  Soft-deletes an employee by ID. Requires admin role.
// @Tags         employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Employee ID"
// @Success      200  {object}  http.APIResponse  "Employee deleted"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /employees/{id} [delete]
func (s *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	employeeIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	employeeID := employees.NewEmployeeID(employeeIDUInt)
	err = s.service.DeleteEmployee(c.Request.Context(), employeeID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	http.Success(c, nil, "Employee deleted")
}
