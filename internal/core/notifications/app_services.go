package notifications

import (
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
	"context"
	"fmt"
	"time"
)

type NotificationService interface {
	Send(ctx context.Context, notif *Notification) error
	GetNotificationsByUserID(ctx context.Context, userID shared.UserID, pagination page.Pagination) (page.Page[Notification], error)
	GetNotificationBySpecification(ctx context.Context, notificationType string, pagination page.Pagination) (page.Page[Notification], error)
	GetNotificationByChannel(ctx context.Context, channel string, pagination page.Pagination) (page.Page[Notification], error)
	GenerateSummary(ctx context.Context, senderID string) ([]Notification, error)
	GetNotificationByID(ctx context.Context, id string) (Notification, error)
}

type notificationService struct {
	repository  NotificationRepository
	emailSender EmailSender
	smsSender   SMSender
}

func NewNotificationService(repository NotificationRepository, emailSender EmailSender, smsSender SMSender) NotificationService {
	return &notificationService{
		repository:  repository,
		emailSender: emailSender,
		smsSender:   smsSender,
	}
}

func (ns *notificationService) Send(ctx context.Context, notif *Notification) error {
	if err := validateForDelivery(notif); err != nil {
		return err
	}
	notif.Status = DeliveryStatusPending
	if err := ns.repository.Create(ctx, notif); err != nil {
		return err
	}
	var providerReference string
	var err error
	switch notif.Channel {
	case NotificationChannelEmail:
		if ns.emailSender == nil {
			err = fmt.Errorf("email sender is required")
		} else {
			providerReference, err = ns.emailSender.Send(ctx, notif)
		}
	case NotificationChannelSMS:
		if ns.smsSender == nil {
			err = fmt.Errorf("SMS sender is required")
		} else {
			providerReference, err = ns.smsSender.Send(ctx, notif)
		}
	default:
		err = fmt.Errorf("unsupported notification channel: %s", notif.Channel)
	}
	notif.ProviderReference = providerReference
	notif.Attempts++
	if err != nil {
		notif.Status = DeliveryStatusFailed
		notif.FailureMessage = err.Error()
		if updateErr := ns.repository.UpdateDelivery(ctx, notif); updateErr != nil {
			return fmt.Errorf("deliver notification: %w (record failure: %v)", err, updateErr)
		}
		return err
	}
	now := time.Now()
	notif.Status, notif.SentAt = DeliveryStatusSent, &now
	return ns.repository.UpdateDelivery(ctx, notif)
}

func (s *notificationService) GetNotificationsByUserID(ctx context.Context, userID shared.UserID, pagination page.Pagination) (page.Page[Notification], error) {
	notificationPage, err := s.repository.FindByUser(ctx, userID, pagination)
	if err != nil {
		return page.Page[Notification]{}, err
	}

	return notificationPage, nil
}

func (s *notificationService) GetNotificationByID(ctx context.Context, id string) (Notification, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *notificationService) GetNotificationByType(ctx context.Context, notificationType string, pagination page.Pagination) (page.Page[Notification], error) {
	return s.repository.FindByType(ctx, notificationType, pagination)
}

func (s *notificationService) GetNotificationBySpecification(ctx context.Context, notificationType string, pagination page.Pagination) (page.Page[Notification], error) {
	return s.GetNotificationByType(ctx, notificationType, pagination)
}

func (s *notificationService) GetNotificationByChannel(ctx context.Context, channel string, pagination page.Pagination) (page.Page[Notification], error) {
	notificationPage, err := s.repository.FindByChannel(ctx, channel, pagination)
	if err != nil {
		return page.Page[Notification]{}, err
	}

	return notificationPage, nil
}

func (s *notificationService) GenerateSummary(ctx context.Context, senderID string) ([]Notification, error) {
	return s.repository.FindSummary(ctx, senderID)
}

func validateForDelivery(notif *Notification) error {
	if notif == nil {
		return fmt.Errorf("notification is required")
	}
	if notif.UserID == "" {
		return fmt.Errorf("notification recipient user ID is required")
	}
	if notif.Subject == "" || notif.Message == "" {
		return fmt.Errorf("notification subject and message are required")
	}
	switch notif.Channel {
	case NotificationChannelEmail:
		if notif.UserEmail == "" {
			return fmt.Errorf("notification recipient email is required")
		}
	case NotificationChannelSMS:
		if notif.UserPhone == "" {
			return fmt.Errorf("notification recipient phone is required")
		}
	default:
		return fmt.Errorf("unsupported notification channel: %s", notif.Channel)
	}
	return nil
}
