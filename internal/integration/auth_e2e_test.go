//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"clinic-vet-api/internal/testkit"
)

func TestCustomerRegistrationActivationAndLogin(t *testing.T) {
	env := testkit.NewEnvironment(t)

	response := request(t, env, http.MethodPost, "/api/v2/auth/register", map[string]any{
		"role":          "customer",
		"email":         "alice@example.com",
		"password":      "SecurePass123!",
		"first_name":    "Alice",
		"last_name":     "Example",
		"date_of_birth": "1990-01-02",
		"gender":        "female",
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}

	var userID uint
	var customerID uint
	var customerActive bool
	if err := env.DB.QueryRow(t.Context(), "SELECT u.id, c.id, c.is_active FROM users u JOIN customers c ON c.user_id = u.id WHERE u.email = $1", "alice@example.com").Scan(&userID, &customerID, &customerActive); err != nil {
		t.Fatalf("read persisted customer: %v", err)
	}
	if customerActive {
		t.Fatal("customer profile is active before account activation")
	}
	if len(env.Email.Sent) != 1 {
		t.Fatalf("activation emails = %d, want 1", len(env.Email.Sent))
	}
	code := strings.TrimPrefix(env.Email.Sent[0].Message, "Use this verification code to activate your account: ")

	response = request(t, env, http.MethodPost, "/api/v2/auth/activate", map[string]any{"user_id": strconv.FormatUint(uint64(userID), 10), "code": code})
	if response.Code != http.StatusOK {
		t.Fatalf("activate status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if err := env.DB.QueryRow(t.Context(), "SELECT is_active FROM customers WHERE user_id = $1", userID).Scan(&customerActive); err != nil {
		t.Fatalf("read activated customer: %v", err)
	}
	if !customerActive {
		t.Fatal("customer profile remains inactive after activation")
	}

	response = request(t, env, http.MethodPost, "/api/v2/auth/login", map[string]any{"email": "alice@example.com", "password": "SecurePass123!"})
	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var login struct {
		Data struct {
			AccessToken struct {
				Token string `json:"token"`
			} `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &login); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if login.Data.AccessToken.Token == "" {
		t.Fatalf("login did not return an access token: %s", response.Body.String())
	}

	profileResponse := authenticatedRequest(t, env, http.MethodGet, "/api/v2/profile/", nil, login.Data.AccessToken.Token)
	if profileResponse.Code != http.StatusOK {
		t.Fatalf("profile status = %d, want %d: %s", profileResponse.Code, http.StatusOK, profileResponse.Body.String())
	}
	notificationResponse := authenticatedRequest(t, env, http.MethodGet, "/api/v2/me/notifications", nil, login.Data.AccessToken.Token)
	if notificationResponse.Code != http.StatusOK {
		t.Fatalf("list notifications status = %d, want %d: %s", notificationResponse.Code, http.StatusOK, notificationResponse.Body.String())
	}

	petResponse := authenticatedRequest(t, env, http.MethodPost, "/api/v2/me/pets/", map[string]any{
		"name": "Milo", "species": "cat", "gender": "male", "customer_id": customerID,
	}, login.Data.AccessToken.Token)
	if petResponse.Code != http.StatusCreated {
		t.Fatalf("create pet status = %d, want %d: %s", petResponse.Code, http.StatusCreated, petResponse.Body.String())
	}
	var petID uint
	if err := env.DB.QueryRow(t.Context(), "SELECT id FROM pets WHERE customer_id = $1 AND name = $2", customerID, "Milo").Scan(&petID); err != nil {
		t.Fatalf("read persisted pet: %v", err)
	}
	petResponse = authenticatedRequest(t, env, http.MethodGet, "/api/v2/me/pets/"+strconv.FormatUint(uint64(petID), 10), nil, login.Data.AccessToken.Token)
	if petResponse.Code != http.StatusOK {
		t.Fatalf("get pet status = %d, want %d: %s", petResponse.Code, http.StatusOK, petResponse.Body.String())
	}

	scheduledDate := time.Now().AddDate(0, 0, 4)
	scheduledDate = time.Date(scheduledDate.Year(), scheduledDate.Month(), scheduledDate.Day(), 10, 0, 0, 0, time.Local)
	appointmentResponse := authenticatedRequest(t, env, http.MethodPost, "/api/v2/me/appointments/", map[string]any{
		"pet_id":         petID,
		"scheduled_date": scheduledDate.Format(time.RFC3339),
		"reason":         "routine_checkup",
		"service":        "general_consultation",
	}, login.Data.AccessToken.Token)
	if appointmentResponse.Code != http.StatusOK {
		t.Fatalf("request appointment status = %d, want %d: %s", appointmentResponse.Code, http.StatusOK, appointmentResponse.Body.String())
	}
	var appointmentID uint
	var appointmentStatus string
	if err := env.DB.QueryRow(t.Context(), "SELECT id, status FROM appointments WHERE customer_id = $1 AND pet_id = $2", customerID, petID).Scan(&appointmentID, &appointmentStatus); err != nil {
		t.Fatalf("read persisted appointment: %v", err)
	}
	if appointmentStatus != "pending" {
		t.Fatalf("appointment status = %q, want pending", appointmentStatus)
	}
	appointmentResponse = authenticatedRequest(t, env, http.MethodGet, "/api/v2/me/appointments/"+strconv.FormatUint(uint64(appointmentID), 10), nil, login.Data.AccessToken.Token)
	if appointmentResponse.Code != http.StatusOK {
		t.Fatalf("get appointment status = %d, want %d: %s", appointmentResponse.Code, http.StatusOK, appointmentResponse.Body.String())
	}
	appointmentResponse = authenticatedRequest(t, env, http.MethodPost, "/api/v2/me/appointments/"+strconv.FormatUint(uint64(appointmentID), 10)+"/cancel?reason=changed+plans", nil, login.Data.AccessToken.Token)
	if appointmentResponse.Code != http.StatusOK {
		t.Fatalf("cancel appointment status = %d, want %d: %s", appointmentResponse.Code, http.StatusOK, appointmentResponse.Body.String())
	}
	if err := env.DB.QueryRow(t.Context(), "SELECT status FROM appointments WHERE id = $1", appointmentID).Scan(&appointmentStatus); err != nil {
		t.Fatalf("read cancelled appointment: %v", err)
	}
	if appointmentStatus != "cancelled" {
		t.Fatalf("appointment status = %q, want cancelled", appointmentStatus)
	}

	const staffUserID = 999
	if _, err := env.DB.Exec(t.Context(), "INSERT INTO users (id, email, role, status, name, oauth_provider, two_fa_method) VALUES ($1, $2, 'admin', 'active', 'E2E Admin', 'local', 'none')", staffUserID, "staff@example.com"); err != nil {
		t.Fatalf("create staff fixture: %v", err)
	}
	staffToken := testkit.AccessToken(t, staffUserID, "admin")
	var employeeID uint
	if err := env.DB.QueryRow(t.Context(), "INSERT INTO employees (first_name, last_name, gender, date_of_birth, photo, license_number, speciality, years_of_experience, user_id) VALUES ('E2E', 'Veterinarian', 'female', '1980-01-01', 'https://example.com/vet.png', 'E2E-999', 'general_practice', 10, $1) RETURNING id", staffUserID).Scan(&employeeID); err != nil {
		t.Fatalf("create veterinarian fixture: %v", err)
	}
	medicalResponse := authenticatedRequest(t, env, http.MethodPost, "/api/v2/medical/sessions", map[string]any{
		"pet_id":         petID,
		"customer_id":    customerID,
		"employee_id":    employeeID,
		"appointment_id": appointmentID,
		"clinic_service": "general_consultation",
		"visit_type":     "consultation",
		"visit_date":     time.Now().UTC().Format(time.RFC3339),
	}, staffToken)
	if medicalResponse.Code != http.StatusCreated {
		t.Fatalf("create medical session status = %d, want %d: %s", medicalResponse.Code, http.StatusCreated, medicalResponse.Body.String())
	}
	var sessionID uint
	if err := env.DB.QueryRow(t.Context(), "SELECT id FROM medical_sessions WHERE pet_id = $1 AND employee_id = $2", petID, employeeID).Scan(&sessionID); err != nil {
		t.Fatalf("read persisted medical session: %v", err)
	}
	veterinarianToken := testkit.AccessToken(t, staffUserID, "veterinarian")
	medicalResponse = authenticatedRequest(t, env, http.MethodGet, "/api/v2/medical/sessions/"+strconv.FormatUint(uint64(sessionID), 10), nil, veterinarianToken)
	if medicalResponse.Code != http.StatusOK {
		t.Fatalf("get assigned veterinarian medical session status = %d, want %d: %s", medicalResponse.Code, http.StatusOK, medicalResponse.Body.String())
	}
	medicalResponse = authenticatedRequest(t, env, http.MethodGet, "/api/v2/me/medical/sessions/"+strconv.FormatUint(uint64(sessionID), 10), nil, login.Data.AccessToken.Token)
	if medicalResponse.Code != http.StatusOK {
		t.Fatalf("get customer medical session status = %d, want %d: %s", medicalResponse.Code, http.StatusOK, medicalResponse.Body.String())
	}
	if response := authenticatedRequest(t, env, http.MethodGet, "/api/v2/medical/sessions", nil, login.Data.AccessToken.Token); response.Code != http.StatusForbidden {
		t.Fatalf("customer medical staff access status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
	}
	if response := authenticatedRequest(t, env, http.MethodGet, "/api/v2/me/medical/sessions", nil, staffToken); response.Code != http.StatusForbidden {
		t.Fatalf("admin customer medical access status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
	}
	paymentResponse := authenticatedRequest(t, env, http.MethodPost, "/api/v2/payments/", map[string]any{
		"amount":         125.50,
		"currency":       "MXN",
		"method":         "cash",
		"reference":      "e2e-payment-001",
		"customer_id":    customerID,
		"appointment_id": appointmentID,
	}, staffToken)
	if paymentResponse.Code != http.StatusCreated {
		t.Fatalf("create payment status = %d, want %d: %s", paymentResponse.Code, http.StatusCreated, paymentResponse.Body.String())
	}
	var paymentID uint
	var paymentStatus string
	if err := env.DB.QueryRow(t.Context(), "SELECT id, status FROM payments WHERE reference = $1", "e2e-payment-001").Scan(&paymentID, &paymentStatus); err != nil {
		t.Fatalf("read persisted payment: %v", err)
	}
	if paymentStatus != "authorized" {
		t.Fatalf("payment status = %q, want authorized", paymentStatus)
	}
	paymentResponse = authenticatedRequest(t, env, http.MethodGet, "/api/v2/me/payments/"+strconv.FormatUint(uint64(paymentID), 10), nil, login.Data.AccessToken.Token)
	if paymentResponse.Code != http.StatusOK {
		t.Fatalf("get customer payment status = %d, want %d: %s", paymentResponse.Code, http.StatusOK, paymentResponse.Body.String())
	}
	if response := authenticatedRequest(t, env, http.MethodGet, "/api/v2/payments/", nil, login.Data.AccessToken.Token); response.Code != http.StatusForbidden {
		t.Fatalf("customer payment staff access status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
	}
	paymentResponse = authenticatedFormRequest(t, env, http.MethodPost, "/api/v2/payments/"+strconv.FormatUint(uint64(paymentID), 10)+"/cancel", "reason=customer+request", staffToken)
	if paymentResponse.Code != http.StatusOK {
		t.Fatalf("cancel payment status = %d, want %d: %s", paymentResponse.Code, http.StatusOK, paymentResponse.Body.String())
	}
	if err := env.DB.QueryRow(t.Context(), "SELECT status FROM payments WHERE id = $1", paymentID).Scan(&paymentStatus); err != nil {
		t.Fatalf("read cancelled payment: %v", err)
	}
	if paymentStatus != "cancelled" {
		t.Fatalf("payment status = %q, want cancelled", paymentStatus)
	}

	emailsBeforeManualSend := len(env.Email.Sent)
	notificationResponse = authenticatedRequest(t, env, http.MethodPost, "/api/v2/notifications", map[string]any{
		"channel":    "email",
		"type":       "info",
		"user_id":    strconv.FormatUint(uint64(userID), 10),
		"user_email": "alice@example.com",
		"subject":    "Clinic update",
		"message":    "Your appointment was updated.",
	}, staffToken)
	if notificationResponse.Code != http.StatusOK {
		t.Fatalf("send notification status = %d, want %d: %s", notificationResponse.Code, http.StatusOK, notificationResponse.Body.String())
	}
	if len(env.Email.Sent) != emailsBeforeManualSend+1 {
		t.Fatalf("emails sent = %d, want %d", len(env.Email.Sent), emailsBeforeManualSend+1)
	}
}

func TestRegistrationRejectsIncompletePublicProfile(t *testing.T) {
	env := testkit.NewEnvironment(t)
	response := request(t, env, http.MethodPost, "/api/v2/auth/register", map[string]any{
		"role":     "customer",
		"email":    "missing@example.com",
		"password": "SecurePass123!",
	})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("register status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestHealthAndProtectedProfileContract(t *testing.T) {
	env := testkit.NewEnvironment(t)
	response := request(t, env, http.MethodGet, "/health", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	response = request(t, env, http.MethodGet, "/api/v2/profile/", nil)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated profile status = %d, want %d: %s", response.Code, http.StatusUnauthorized, response.Body.String())
	}
}

func TestAdminUserAndEmployeeLookupEndpoints(t *testing.T) {
	env := testkit.NewEnvironment(t)
	const adminID = 700
	const employeeUserID = 701
	if _, err := env.DB.Exec(t.Context(), `
		INSERT INTO users (id, email, phone_number, role, status, name, oauth_provider, two_fa_method)
		VALUES
			($1, 'admin-lookup@example.com', '+525500000000', 'admin', 'active', 'Lookup Admin', 'local', 'none'),
			($2, 'vet-lookup@example.com', '+525511111111', 'veterinarian', 'active', 'Lookup Vet', 'local', 'none')`, adminID, employeeUserID); err != nil {
		t.Fatalf("create lookup users: %v", err)
	}
	if _, err := env.DB.Exec(t.Context(), `
		INSERT INTO employees (first_name, last_name, gender, date_of_birth, photo, license_number, speciality, years_of_experience, is_active, user_id)
		VALUES ('Lookup', 'Vet', 'female', '1980-01-01', 'https://example.com/vet.png', 'LOOKUP-701', 'general_practice', 10, true, $1)`, employeeUserID); err != nil {
		t.Fatalf("create lookup employee: %v", err)
	}
	adminToken := testkit.AccessToken(t, adminID, "admin")

	for _, endpoint := range []string{
		"/api/v2/users/email/vet-lookup@example.com",
		"/api/v2/users/phone?phone=%2B525511111111",
		"/api/v2/employees/active",
		"/api/v2/employees/specialty/general_practice",
		"/api/v2/employees/stats",
	} {
		response := authenticatedRequest(t, env, http.MethodGet, endpoint, nil, adminToken)
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d: %s", endpoint, response.Code, http.StatusOK, response.Body.String())
		}
	}

	for _, expectedError := range []struct {
		endpoint string
		status   int
	}{
		{"/api/v2/users/email/not-an-email", http.StatusUnprocessableEntity},
		{"/api/v2/users/email/missing@example.com", http.StatusNotFound},
		{"/api/v2/users/phone", http.StatusBadRequest},
		{"/api/v2/employees/specialty/not-a-specialty", http.StatusBadRequest},
	} {
		response := authenticatedRequest(t, env, http.MethodGet, expectedError.endpoint, nil, adminToken)
		if response.Code != expectedError.status {
			t.Fatalf("%s status = %d, want %d: %s", expectedError.endpoint, response.Code, expectedError.status, response.Body.String())
		}
	}
}

func TestAuthenticatedPasswordAndAccountEndpoints(t *testing.T) {
	env := testkit.NewEnvironment(t)
	response := request(t, env, http.MethodPost, "/api/v2/auth/register", map[string]any{
		"role": "customer", "email": "account@example.com", "password": "SecurePass123!",
		"first_name": "Account", "last_name": "Owner", "date_of_birth": "1990-01-02", "gender": "female",
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}
	var userID uint
	if err := env.DB.QueryRow(t.Context(), "SELECT id FROM users WHERE email = $1", "account@example.com").Scan(&userID); err != nil {
		t.Fatalf("read account user: %v", err)
	}
	code := strings.TrimPrefix(env.Email.Sent[0].Message, "Use this verification code to activate your account: ")
	response = request(t, env, http.MethodPost, "/api/v2/auth/activate", map[string]any{"user_id": strconv.FormatUint(uint64(userID), 10), "code": code})
	if response.Code != http.StatusOK {
		t.Fatalf("activate status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	response = request(t, env, http.MethodPost, "/api/v2/auth/login", map[string]any{"email": "account@example.com", "password": "SecurePass123!"})
	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var login struct {
		Data struct {
			AccessToken struct {
				Token string `json:"token"`
			} `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &login); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	response = authenticatedRequest(t, env, http.MethodPut, "/api/v2/auth/password", map[string]any{
		"current_password": "wrong-password", "new_password": "NewSecurePass123!",
	}, login.Data.AccessToken.Token)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("wrong current password status = %d, want %d: %s", response.Code, http.StatusUnprocessableEntity, response.Body.String())
	}
	response = authenticatedRequest(t, env, http.MethodPut, "/api/v2/auth/password", map[string]any{
		"current_password": "SecurePass123!", "new_password": "NewSecurePass123!",
	}, login.Data.AccessToken.Token)
	if response.Code != http.StatusOK {
		t.Fatalf("update password status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	response = authenticatedRequest(t, env, http.MethodDelete, "/api/v2/auth/me", nil, login.Data.AccessToken.Token)
	if response.Code != http.StatusOK {
		t.Fatalf("delete account status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var deletedAt *time.Time
	if err := env.DB.QueryRow(t.Context(), "SELECT deleted_at FROM users WHERE id = $1", userID).Scan(&deletedAt); err != nil {
		t.Fatalf("read deleted account: %v", err)
	}
	if deletedAt == nil {
		t.Fatal("deleted account has no deleted_at timestamp")
	}
}

func request(t *testing.T, env *testkit.Environment, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request: %v", err)
		}
		payload = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, payload)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	env.Router.ServeHTTP(response, req)
	return response
}

func authenticatedRequest(t *testing.T, env *testkit.Environment, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request: %v", err)
		}
		payload = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, payload)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	env.Router.ServeHTTP(response, req)
	return response
}

func authenticatedFormRequest(t *testing.T, env *testkit.Environment, method, path, form, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(form))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	env.Router.ServeHTTP(response, req)
	return response
}
