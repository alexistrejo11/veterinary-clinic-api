package dtos

import "time"

// SendNotificationRequest is the body for manually sending a notification (staff)
// @Description Channel (email/sms), type, optional user_id, user_email (for email), user_phone (for sms), title, subject, message, optional token (e.g. reset link).
type SendNotificationRequest struct {
	Channel   string `json:"channel" binding:"required"` // email, sms
	Type      string `json:"type" binding:"required"`    // notification type
	UserID    string `json:"user_id"`                    // optional; for in-app targeting
	UserEmail string `json:"user_email"`                 // required for email
	UserPhone string `json:"user_phone"`                 // required for sms
	Title     string `json:"title"`
	Subject   string `json:"subject" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Token     string `json:"token"` // optional; e.g. for reset links
}

// NotificationResponse represents a notification returned in HTTP responses
// @Description Notification entity: id, user_id, user_email, user_phone, title, subject, message, type, channel, created_at.
type NotificationResponse struct {
	// Unique identifier of the notification
	ID string `json:"id" example:"665f1c2ab3d4e5f6a7b8c9d0"`

	// ID of the target user (if any)
	UserID string `json:"user_id,omitempty" example:"123"`

	// Target email (for email channel)
	UserEmail string `json:"user_email,omitempty" example:"user@example.com"`

	// Target phone (for sms channel)
	UserPhone string `json:"user_phone,omitempty" example:"+1234567890"`

	// Notification title
	Title string `json:"title,omitempty" example:"Appointment reminder"`

	// Notification subject
	Subject string `json:"subject" example:"Your appointment is tomorrow"`

	// Notification message body
	Message string `json:"message" example:"Don't forget your appointment at 10:00."`

	// Notification type
	Type string `json:"type" example:"appointment_reminder"`

	// Delivery channel (email, sms)
	Channel string `json:"channel" example:"email"`

	Status            string     `json:"status" example:"sent"`
	FailureMessage    string     `json:"failure_message,omitempty"`
	ProviderReference string     `json:"provider_reference,omitempty"`
	SentAt            *time.Time `json:"sent_at,omitempty"`
	Attempts          int        `json:"attempts"`

	// Creation timestamp
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}
