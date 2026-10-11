package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type appointmentCommandStub struct {
	appointments.CommandService
	complete  *appointments.CompleteCommand
	cancel    *appointments.CancelCommand
	notAttend *appointments.NotAttendCommand
}

func (s *appointmentCommandStub) CompleteAppointment(_ context.Context, command appointments.CompleteCommand) error {
	s.complete = &command
	return nil
}

func (s *appointmentCommandStub) CancelAppointment(_ context.Context, command appointments.CancelCommand) error {
	s.cancel = &command
	return nil
}

func (s *appointmentCommandStub) MarkAppointmentAsNotAttend(_ context.Context, command appointments.NotAttendCommand) error {
	s.notAttend = &command
	return nil
}

func TestAdministrativeAppointmentTransitionsPassOptionalEmployeeID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		handle func(*AppointmentHandler, *gin.Context)
		assert func(*testing.T, *appointmentCommandStub)
	}{
		{"complete", (*AppointmentHandler).CompleteAppointment, func(t *testing.T, stub *appointmentCommandStub) {
			if stub.complete == nil || stub.complete.EmployeeID == nil || stub.complete.EmployeeID.Value() != 7 {
				t.Fatalf("complete command employee ID = %#v, want 7", stub.complete)
			}
		}},
		{"cancel", (*AppointmentHandler).CancelAppointment, func(t *testing.T, stub *appointmentCommandStub) {
			if stub.cancel == nil || stub.cancel.EmployeeID == nil || stub.cancel.EmployeeID.Value() != 7 {
				t.Fatalf("cancel command employee ID = %#v, want 7", stub.cancel)
			}
		}},
		{"not attend", (*AppointmentHandler).NotAttendAppointment, func(t *testing.T, stub *appointmentCommandStub) {
			if stub.notAttend == nil || stub.notAttend.EmployeeID == nil || stub.notAttend.EmployeeID.Value() != 7 {
				t.Fatalf("not-attend command employee ID = %#v, want 7", stub.notAttend)
			}
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &appointmentCommandStub{}
			handler := NewAppointmentHandler(stub, nil, validator.New(), mappers.NewAppointmentResponseMapper(), nil, nil)
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			ctx.Params = gin.Params{{Key: "id", Value: "9"}}
			ctx.Request = httptest.NewRequest(http.MethodPost, "/appointments/9?employee_id=7", nil)

			test.handle(handler, ctx)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
			}
			test.assert(t, stub)
		})
	}
}

func TestAdministrativeAppointmentTransitionsRejectInvalidEmployeeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAppointmentHandler(nil, nil, validator.New(), mappers.NewAppointmentResponseMapper(), nil, nil)

	for name, handle := range map[string]func(*gin.Context){
		"complete":   handler.CompleteAppointment,
		"cancel":     handler.CancelAppointment,
		"not attend": handler.NotAttendAppointment,
	} {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			ctx.Params = gin.Params{{Key: "id", Value: "9"}}
			ctx.Request = httptest.NewRequest(http.MethodPost, "/appointments/9?employee_id=invalid", nil)

			handle(ctx)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
			}
		})
	}
}
