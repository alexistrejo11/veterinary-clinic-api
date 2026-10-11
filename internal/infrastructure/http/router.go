package http

import (
	"errors"

	"clinic-vet-api/internal/infrastructure/http/handlers"
)

type APIRouter struct {
	appHandlers *AppHandlers
	config      *APIConfig
}

type AppHandlers struct {
	address      *handlers.AddressHandler
	auth         *handlers.AuthHandler
	user         *handlers.UserHandler
	pets         *handlers.PetHandler
	profile      *handlers.ProfileHandler
	appointment  *handlers.AppointmentHandler
	customer     *handlers.CustomerHandler
	employee     *handlers.EmployeeHandler
	payment      *handlers.PaymentHandler
	medical      *handlers.MedicalHandler
	notification *handlers.NotificationHandler
	home         *handlers.HomeHandler
}

func (r *APIRouter) Validate() error {
	if r.appHandlers == nil {
		return errors.New("handlers are required")
	}
	if err := r.appHandlers.Validate(); err != nil {
		return err
	}
	if r.appHandlers.auth == nil {
		return errors.New("auth handler is required")
	}
	if r.config.AuthMiddleware == nil {
		return errors.New("auth middleware is required")
	}
	if r.config.Router == nil {
		return errors.New("router group is required")
	}
	return nil
}

func (a *AppHandlers) Validate() error {
	if a.address == nil {
		return errors.New("address handler is required")
	}
	if a.auth == nil {
		return errors.New("auth handler is required")
	}
	if a.user == nil {
		return errors.New("user handler is required")
	}
	if a.pets == nil {
		return errors.New("pets handler is required")
	}
	if a.profile == nil {
		return errors.New("profile handler is required")
	}
	if a.appointment == nil {
		return errors.New("appointment handler is required")
	}
	if a.customer == nil {
		return errors.New("customer handler is required")
	}
	if a.employee == nil {
		return errors.New("employee handler is required")
	}
	if a.payment == nil {
		return errors.New("payment handler is required")
	}
	if a.medical == nil {
		return errors.New("medical handler is required")
	}
	if a.notification == nil {
		return errors.New("notification handler is required")
	}
	if a.home == nil {
		return errors.New("home handler is required")
	}
	return nil
}

func NewAPIRouter(
	appHandlers *AppHandlers,
	config *APIConfig,
) (*APIRouter, error) {
	router := &APIRouter{
		appHandlers: appHandlers,
		config:      config,
	}

	if err := router.Validate(); err != nil {
		return nil, err
	}
	return router, nil
}

// RegisterRoutes mounts all route groups (auth, profile, users, appointments, etc.).
// Call this after NewAPIRouter so that routes are registered.
func (r *APIRouter) RegisterRoutes() {
	r.authRoutes()
	r.profileRoutes()
	r.userRoutes()
	r.customerRoutes()
	r.employeeRoutes()
	r.appointmentRoutes()
	r.addressRoutes()
	r.petRoutes()
	r.paymentRoutes()
	r.medicalRoutes()
	r.notificationRoutes()
	r.homeRoutes()
}

// ------------------------------------------------------------
// Auth Routes
// ------------------------------------------------------------

func (r *APIRouter) authRoutes() {
	publicAuthRoutes := r.config.Router.Group("/api/v2/auth")
	{
		publicAuthRoutes.POST("/register", r.appHandlers.auth.Register)
		publicAuthRoutes.POST("/login", r.appHandlers.auth.Login)
		publicAuthRoutes.POST("/activate", r.appHandlers.auth.ActivateAccount)
		publicAuthRoutes.POST("/verify-email", r.appHandlers.auth.VerifyEmail)
		publicAuthRoutes.POST("/refresh", r.appHandlers.auth.RefreshToken)
		publicAuthRoutes.POST("/2fa/verify", r.appHandlers.auth.VerifyTwoFactor)
		publicAuthRoutes.POST("/reset-password/request", r.appHandlers.auth.RequestResetPassword)
		publicAuthRoutes.POST("/reset-password/reset", r.appHandlers.auth.ResetPassword)
	}

	authenticatedAuthRoutes := r.config.Router.Group("/api/v2/auth")
	authenticatedAuthRoutes.Use(r.config.AuthMiddleware.Authenticate())
	{
		authenticatedAuthRoutes.POST("/logout", r.appHandlers.auth.Logout)
		authenticatedAuthRoutes.POST("/logout-all", r.appHandlers.auth.LogoutAll)
		authenticatedAuthRoutes.PUT("/password", r.appHandlers.auth.UpdatePassword)
		authenticatedAuthRoutes.DELETE("/me", r.appHandlers.auth.DeleteMyAccount)
	}

	twoFactorAuthRoutes := r.config.Router.Group("/api/v2/auth/2fa")
	twoFactorAuthRoutes.Use(r.config.AuthMiddleware.Authenticate())
	{
		twoFactorAuthRoutes.POST("/enable", r.appHandlers.auth.EnableTwoFactor)
		twoFactorAuthRoutes.POST("/disable", r.appHandlers.auth.DisableTwoFactor)
	}
}

// ------------------------------------------------------------
// Profile Routes
// ------------------------------------------------------------

func (r *APIRouter) profileRoutes() {
	profileRoutes := r.config.Router.Group("/api/v2/profile")
	profileRoutes.Use(r.config.AuthMiddleware.Authenticate())
	{
		profileRoutes.GET("/", r.appHandlers.profile.GetProfile)
		profileRoutes.PUT("/", r.appHandlers.profile.UpdateProfile)
	}
}

// ------------------------------------------------------------
// User Routes
// ------------------------------------------------------------

func (r *APIRouter) userRoutes() {
	userRoutes := r.config.Router.Group("/api/v2/users")
	userRoutes.Use(r.config.AuthMiddleware.Authenticate())
	userRoutes.Use(r.config.AuthMiddleware.RequireAnyRole("admin"))
	{
		userRoutes.GET("/", r.appHandlers.user.SearchUsers)
		userRoutes.GET("/email/:email", r.appHandlers.user.GetUserByEmail)
		userRoutes.GET("/phone", r.appHandlers.user.GetUserByPhone)
		userRoutes.GET("/:id", r.appHandlers.user.GetUserByID)
		userRoutes.POST("/", r.appHandlers.user.CreateUser)
		userRoutes.POST("/:id/restore", r.appHandlers.user.RestoreUser)
		userRoutes.POST("/:id/status", r.appHandlers.user.UpdateUserStatus)
		userRoutes.DELETE("/:id", r.appHandlers.user.DeleteUser)
	}
}

// ------------------------------------------------------------
// Appointment Routes
// ------------------------------------------------------------

func (r *APIRouter) appointmentRoutes() {
	// ----- Customer: "my" appointments (customer sees only their own) -----
	meAppointments := r.config.Router.Group("/api/v2/me/appointments")
	meAppointments.Use(r.config.AuthMiddleware.Authenticate())
	meAppointments.Use(r.config.AuthMiddleware.RequireAnyRole("customer"))
	{
		meAppointments.GET("/", r.appHandlers.appointment.GetMyAppointments)
		meAppointments.GET("/:id", r.appHandlers.appointment.GetMyAppointment)
		meAppointments.POST("/", r.appHandlers.appointment.RequestAppointment)
		meAppointments.POST("/:id/cancel", r.appHandlers.appointment.CancelMyAppointment)
	}

	// ----- Veterinarian: "my" assigned appointments -----
	employeeAppointments := r.config.Router.Group("/api/v2/employees/appointments")
	employeeAppointments.Use(r.config.AuthMiddleware.Authenticate())
	employeeAppointments.Use(r.config.AuthMiddleware.RequireAnyRole("veterinarian"))
	{
		employeeAppointments.GET("/", r.appHandlers.appointment.GetMyAppointmentsAsEmployee)
		employeeAppointments.GET("/:id", r.appHandlers.appointment.GetMyAppointmentAsEmployee)
		employeeAppointments.PUT("/:id", r.appHandlers.appointment.UpdateMyAppointment)
		employeeAppointments.POST("/:id/reschedule", r.appHandlers.appointment.RescheduleMyAppointment)
		employeeAppointments.POST("/:id/complete", r.appHandlers.appointment.CompleteMyAppointment)
		employeeAppointments.POST("/:id/cancel", r.appHandlers.appointment.CancelAssignedAppointment)
		employeeAppointments.POST("/:id/not-attend", r.appHandlers.appointment.NotAttendMyAppointment)
	}

	// ----- Receptionist/Admin: all appointments (search, CRUD, by customer/employee/pet) -----
	administrativeAppointments := r.config.Router.Group("/api/v2/appointments")
	administrativeAppointments.Use(r.config.AuthMiddleware.Authenticate())
	administrativeAppointments.Use(r.config.AuthMiddleware.RequireAnyRole("admin", "receptionist"))
	{
		administrativeAppointments.GET("/", r.appHandlers.appointment.SearchAppointments)
		administrativeAppointments.GET("/:id", r.appHandlers.appointment.GetAppointmentByID)
		administrativeAppointments.POST("/", r.appHandlers.appointment.CreateAppointment)
		administrativeAppointments.PUT("/:id", r.appHandlers.appointment.UpdateAppointment)
		administrativeAppointments.DELETE("/:id", r.appHandlers.appointment.DeleteAppointment)
		administrativeAppointments.POST("/:id/reschedule", r.appHandlers.appointment.RescheduleAppointment)
		administrativeAppointments.POST("/:id/confirm", r.appHandlers.appointment.ConfirmAppointment)      // ?employee_id=1
		administrativeAppointments.POST("/:id/complete", r.appHandlers.appointment.CompleteAppointment)    // ?employee_id=1&notes=
		administrativeAppointments.POST("/:id/cancel", r.appHandlers.appointment.CancelAppointment)        // ?employee_id=1&reason=
		administrativeAppointments.POST("/:id/not-attend", r.appHandlers.appointment.NotAttendAppointment) // ?employee_id=1
	}

	// Receptionist/Admin: list appointments by customer (use :id to match /customers/:id)
	customersAppointments := r.config.Router.Group("/api/v2/customers/:id/appointments")
	customersAppointments.Use(r.config.AuthMiddleware.Authenticate())
	customersAppointments.Use(r.config.AuthMiddleware.RequireAnyRole("admin", "receptionist"))
	{
		customersAppointments.GET("/", r.appHandlers.appointment.GetAppointmentsByCustomerID)
	}

	// Receptionist/Admin: list appointments by employee (use :id to match /employees/:id)
	employeesAppointments := r.config.Router.Group("/api/v2/employees/:id/appointments")
	employeesAppointments.Use(r.config.AuthMiddleware.Authenticate())
	employeesAppointments.Use(r.config.AuthMiddleware.RequireAnyRole("admin", "receptionist"))
	{
		employeesAppointments.GET("/", r.appHandlers.appointment.GetAppointmentsByEmployeeID)
	}

	// Receptionist/Admin: list appointments by pet
	petsAppointments := r.config.Router.Group("/api/v2/pets/:id/appointments")
	petsAppointments.Use(r.config.AuthMiddleware.Authenticate())
	petsAppointments.Use(r.config.AuthMiddleware.RequireAnyRole("admin", "receptionist"))
	{
		petsAppointments.GET("/", r.appHandlers.appointment.GetAppointmentsByPetID)
	}
}

// ------------------------------------------------------------
// Customer Routes
// ------------------------------------------------------------

func (r *APIRouter) customerRoutes() {
	customerRoutes := r.config.Router.Group("/api/v2/customers")
	customerRoutes.Use(r.config.AuthMiddleware.Authenticate())
	customerRoutes.Use(r.config.AuthMiddleware.RequireAnyRole("admin", "receptionist"))
	{
		customerRoutes.GET("/", r.appHandlers.customer.SearchCustomers)
		customerRoutes.GET("/:id", r.appHandlers.customer.GetCustomerByID)
		customerRoutes.POST("/", r.appHandlers.customer.CreateCustomer)
		customerRoutes.PUT("/:id", r.appHandlers.customer.UpdateCustomer)
		customerRoutes.DELETE("/:id", r.appHandlers.customer.DeleteCustomer)
		customerRoutes.POST("/:id/restore", r.appHandlers.customer.RestoreCustomer)
	}
}

// ------------------------------------------------------------
// Employee Routes
// ------------------------------------------------------------

func (r *APIRouter) employeeRoutes() {
	employeeRoutes := r.config.Router.Group("/api/v2/employees")
	employeeRoutes.Use(r.config.AuthMiddleware.Authenticate())
	employeeRoutes.Use(r.config.AuthMiddleware.RequireAnyRole("admin"))
	{
		employeeRoutes.GET("/", r.appHandlers.employee.SearchEmployees)
		employeeRoutes.GET("/active", r.appHandlers.employee.GetActiveEmployees)
		employeeRoutes.GET("/specialty/:specialty", r.appHandlers.employee.GetEmployeesBySpecialty)
		employeeRoutes.GET("/stats", r.appHandlers.employee.GetEmployeeStats)
		employeeRoutes.GET("/:id", r.appHandlers.employee.GetEmployeeByID)
		employeeRoutes.POST("/", r.appHandlers.employee.CreateEmployee)
		employeeRoutes.PUT("/:id", r.appHandlers.employee.UpdateEmployee)
		employeeRoutes.DELETE("/:id", r.appHandlers.employee.DeleteEmployee)
		employeeRoutes.POST("/:id/restore", r.appHandlers.employee.RestoreEmployee)
	}
}

// ------------------------------------------------------------
// Address Routes
// ------------------------------------------------------------

func (r *APIRouter) addressRoutes() {
	addresses := r.config.Router.Group("/api/v2/addresses")
	addresses.Use(r.config.AuthMiddleware.Authenticate())
	addresses.Use(r.config.AuthMiddleware.RequireAnyRole("admin", "receptionist"))
	{
		addresses.GET("/", r.appHandlers.address.SearchAddresses)
		addresses.GET("/users/:user_id", r.appHandlers.address.GetAddressesByUserID)
		addresses.GET("/users/:user_id/:id", r.appHandlers.address.GetAddressByIDForUser)
		addresses.POST("/users/:user_id", r.appHandlers.address.CreateAddress)
		addresses.PUT("/users/:user_id/:id", r.appHandlers.address.UpdateAddress)
		addresses.DELETE("/users/:user_id/:id", r.appHandlers.address.DeleteAddress)
	}

	myAddresses := r.config.Router.Group("/api/v2/me/addresses")
	myAddresses.Use(r.config.AuthMiddleware.Authenticate())
	{
		myAddresses.GET("/", r.appHandlers.address.GetMyAddresses)
		myAddresses.GET("/:id", r.appHandlers.address.GetAddressByID)
		myAddresses.POST("/", r.appHandlers.address.CreateAddressForMe)
		myAddresses.PUT("/:id", r.appHandlers.address.UpdateMyAddress)
		myAddresses.DELETE("/:id", r.appHandlers.address.DeleteMyAddress)
	}
}

// ------------------------------------------------------------
// Pet Routes
// ------------------------------------------------------------

func (r *APIRouter) petRoutes() {
	pets := r.config.Router.Group("/api/v2/pets")
	pets.Use(r.config.AuthMiddleware.Authenticate())
	pets.Use(r.config.AuthMiddleware.RequireAnyRole("admin", "receptionist"))
	{
		pets.GET("/", r.appHandlers.pets.SearchPets)
		pets.GET("/:id", r.appHandlers.pets.GetPetByID)
		pets.POST("/customers/:customer_id", r.appHandlers.pets.CreatePet)
		pets.GET("/customers/:customer_id", r.appHandlers.pets.GetPetsByCustomerID)
		pets.PUT("/:id", r.appHandlers.pets.UpdatePet)
		pets.DELETE("/:id", r.appHandlers.pets.DeletePet)
		pets.POST("/:id/restore", r.appHandlers.pets.RestorePet)
	}

	myPets := r.config.Router.Group("/api/v2/me/pets")
	myPets.Use(r.config.AuthMiddleware.Authenticate())
	myPets.Use(r.config.AuthMiddleware.RequireAnyRole("customer"))
	{
		myPets.GET("/", r.appHandlers.pets.GetMyPets)
		myPets.GET("/:id", r.appHandlers.pets.GetMyPet)
		myPets.POST("/", r.appHandlers.pets.CreatePetForMe)
		myPets.PUT("/:id", r.appHandlers.pets.UpdateMyPet)
		myPets.DELETE("/:id", r.appHandlers.pets.DeleteMyPet)
	}
}

// ------------------------------------------------------------
// Payment Routes
// ------------------------------------------------------------

func (r *APIRouter) paymentRoutes() {
	myPayments := r.config.Router.Group("/api/v2/me/payments")
	myPayments.Use(r.config.AuthMiddleware.Authenticate())
	myPayments.Use(r.config.AuthMiddleware.RequireAnyRole("customer"))
	{
		myPayments.GET("/", r.appHandlers.payment.GetMyPayments)
		myPayments.GET("/:id", r.appHandlers.payment.GetMyPayment)
	}

	payments := r.config.Router.Group("/api/v2/payments")
	payments.Use(r.config.AuthMiddleware.Authenticate())
	payments.Use(r.config.AuthMiddleware.RequireAnyRole("admin", "receptionist"))
	{
		payments.GET("/", r.appHandlers.payment.GetPaymentsBySpecification)
		payments.GET("/:id", r.appHandlers.payment.GetPaymentByID)
		payments.GET("/reference/:reference", r.appHandlers.payment.GetPaymentByReference)
		payments.GET("/customers/:id", r.appHandlers.payment.GetPaymentsByCustomerID)
		payments.POST("/", r.appHandlers.payment.CreatePayment)
		payments.PUT("/:id", r.appHandlers.payment.UpdatePayment)
		payments.DELETE("/:id", r.appHandlers.payment.DeletePayment)
		payments.POST("/:id/process", r.appHandlers.payment.ProcessPayment)
		payments.POST("/:id/cancel", r.appHandlers.payment.CancelPayment)
		payments.POST("/:id/refund", r.appHandlers.payment.RefundPayment)
	}
}

// ------------------------------------------------------------
// Medical Routes
// ------------------------------------------------------------

func (r *APIRouter) medicalRoutes() {
	m := r.appHandlers.medical

	// Customer: read-only (my sessions, my pets’ data)
	meMedical := r.config.Router.Group("/api/v2/me/medical")
	meMedical.Use(r.config.AuthMiddleware.Authenticate())
	meMedical.Use(r.config.AuthMiddleware.RequireAnyRole("customer"))
	{
		meMedical.GET("/sessions", m.GetMySessions)
		meMedical.GET("/sessions/:id/full", m.GetMySessionFull)
		meMedical.GET("/sessions/:id", m.GetMySessionByID)
		meMedical.GET("/pets/:pet_id/sessions", m.GetMyPetSessions)
		meMedical.GET("/pets/:pet_id/vaccination-summary", m.GetMyPetVaccinationSummary)
		meMedical.GET("/pets/:pet_id/vaccination-history", m.GetMyPetVaccinationHistory)
		meMedical.GET("/pets/:pet_id/prescriptions/active", m.GetMyPetActivePrescriptions)
	}

	// Medical staff (veterinarian + admin): read and write
	medical := r.config.Router.Group("/api/v2/medical")
	medical.Use(r.config.AuthMiddleware.Authenticate())
	medical.Use(r.config.AuthMiddleware.RequireAnyRole("veterinarian", "admin"))
	{
		// Sessions (more specific routes first)
		medical.GET("/sessions", m.GetSessionsBySpecification)
		medical.GET("/sessions/stats", r.config.AuthMiddleware.RequireAnyRole("admin"), m.GetSessionStats)
		medical.GET("/sessions/:id/full", m.GetSessionFull)
		medical.GET("/sessions/:id", m.GetSessionByID)
		medical.POST("/sessions", m.CreateSession)
		medical.PUT("/sessions/:id", m.UpdateSession)
		medical.DELETE("/sessions/:id", m.SoftDeleteSession)
		medical.DELETE("/sessions/:id/hard", m.HardDeleteSession)
		medical.POST("/sessions/:id/restore", m.RestoreSession)
		medical.GET("/customers/:customer_id/sessions", m.GetSessionsByCustomer)
		medical.GET("/pets/:pet_id/sessions", m.GetSessionsByPet)

		// Session extensions
		medical.GET("/sessions/:id/vaccinations", m.GetVaccinationsBySession)
		medical.POST("/vaccinations", m.AddVaccination)
		medical.PUT("/vaccinations/:vaccination_id", m.UpdateVaccination)
		medical.DELETE("/vaccinations/:vaccination_id", m.RemoveVaccination)

		medical.GET("/sessions/:id/surgeries", m.GetSurgeriesBySession)
		medical.POST("/surgeries", m.AddSurgery)
		medical.PUT("/surgeries/:surgery_id", m.UpdateSurgery)
		medical.DELETE("/surgeries/:surgery_id", m.RemoveSurgery)

		medical.GET("/sessions/:id/prescriptions", m.GetPrescriptionsBySession)
		medical.POST("/prescriptions", m.AddPrescription)
		medical.PUT("/prescriptions/:prescription_id", m.UpdatePrescription)
		medical.DELETE("/prescriptions/:prescription_id", m.RemovePrescription)
		medical.GET("/pets/:pet_id/prescriptions/active", m.GetActivePrescriptionsByPet)

		medical.GET("/sessions/:id/attachments", m.GetAttachmentsBySession)
		medical.POST("/attachments", m.AddAttachment)
		medical.DELETE("/attachments/:attachment_id", m.RemoveAttachment)

		medical.GET("/sessions/:id/services", m.GetServicesBySession)
		medical.POST("/session-services", m.AddSessionService)
		medical.DELETE("/session-services/:session_service_id", m.RemoveSessionService)

		medical.GET("/vaccination-history", m.GetVaccinationHistory)
		medical.GET("/pets/:pet_id/vaccination-summary", m.GetPetVaccinationSummary)

		// Catalogs (read + write)
		medical.GET("/catalogs/vaccines", m.ListVaccines)
		medical.GET("/catalogs/vaccines/species/:species", m.ListVaccinesBySpecies)
		medical.GET("/catalogs/vaccines/:id", m.GetVaccineByID)
		medical.POST("/catalogs/vaccines", m.CreateVaccineCatalog)
		medical.DELETE("/catalogs/vaccines/:id", m.DeactivateVaccine)

		medical.GET("/catalogs/medications", m.ListMedications)
		medical.GET("/catalogs/medications/search", m.SearchMedications)
		medical.GET("/catalogs/medications/:id", m.GetMedicationByID)
		medical.POST("/catalogs/medications", m.CreateMedicationCatalog)
		medical.DELETE("/catalogs/medications/:id", m.DeactivateMedication)

		medical.GET("/catalogs/services", m.ListServices)
		medical.GET("/catalogs/services/category/:category", m.ListServicesByCategory)
		medical.GET("/catalogs/services/:id", m.GetServiceByID)
		medical.POST("/catalogs/services", m.CreateServiceCatalog)
		medical.DELETE("/catalogs/services/:id", m.DeactivateServiceCatalog)
	}
}

// ------------------------------------------------------------
// Notification Routes
// ------------------------------------------------------------

func (r *APIRouter) notificationRoutes() {
	n := r.appHandlers.notification

	// Customer: read-only (my notifications)
	meNotif := r.config.Router.Group("/api/v2/me/notifications")
	meNotif.Use(r.config.AuthMiddleware.Authenticate())
	meNotif.Use(r.config.AuthMiddleware.RequireAnyRole("customer", "veterinarian", "receptionist", "admin"))
	{
		meNotif.GET("", n.GetMyNotifications)
		meNotif.GET("/:id", n.GetMyNotificationByID)
	}

	// Staff: monitoring (by type, channel, id, summary) + manual send
	notif := r.config.Router.Group("/api/v2/notifications")
	notif.Use(r.config.AuthMiddleware.Authenticate())
	notif.Use(r.config.AuthMiddleware.RequireAnyRole("veterinarian", "receptionist", "admin"))
	{
		notif.GET("/type/:type", n.GetNotificationsByType)
		notif.GET("/channel/:channel", n.GetNotificationsByChannel)
		notif.GET("/summary", n.GetNotificationSummary)
		notif.GET("/:id", n.GetNotificationByID)
		notif.POST("", n.SendNotification)
	}
}

// ------------------------------------------------------------
// Home Routes
// ------------------------------------------------------------

func (r *APIRouter) homeRoutes() {
	homeRoutes := r.config.Router.Group("")
	homeRoutes.GET("/", r.appHandlers.home.Home)
	homeRoutes.GET("/home", r.appHandlers.home.Home)
	homeRoutes.GET("/health", r.appHandlers.home.HealthCheck)
}
