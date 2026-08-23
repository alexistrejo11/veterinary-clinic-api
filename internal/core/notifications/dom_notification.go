package notifications

import (
	"time"
)

type Notification struct {
	ID                string
	UserID            string
	UserEmail         string
	UserPhone         string
	SenderID          string
	Title             string
	Subject           string
	Message           string
	Token             string
	CreatedAt         time.Time
	NType             NotificationType
	Channel           NotificationChannel
	Status            DeliveryStatus
	FailureMessage    string
	ProviderReference string
	SentAt            *time.Time
	Attempts          int
}

type DeliveryStatus string

const (
	DeliveryStatusPending DeliveryStatus = "pending"
	DeliveryStatusSent    DeliveryStatus = "sent"
	DeliveryStatusFailed  DeliveryStatus = "failed"
)

func (s DeliveryStatus) String() string { return string(s) }

func NewActivateAccountNotification(userID, userEmail string) Notification {
	return Notification{
		UserID:    userID,
		UserEmail: userEmail,
		Subject:   "Activate your account",
		Message:   "Please click the link below to activate your account.",
		NType:     NotificationTypeActivationToken,
		Channel:   NotificationChannelEmail,
	}
}
