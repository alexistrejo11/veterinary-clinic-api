package notifications

import (
	"context"
	"errors"
	"testing"

	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
)

type notificationRepositoryStub struct {
	created             bool
	updated             bool
	updatedNotification Notification
}

func (r *notificationRepositoryStub) Create(_ context.Context, notification *Notification) error {
	r.created = true
	notification.ID = "1"
	return nil
}
func (r *notificationRepositoryStub) UpdateDelivery(_ context.Context, notification *Notification) error {
	r.updated = true
	r.updatedNotification = *notification
	return nil
}
func (r *notificationRepositoryStub) FindByID(context.Context, string) (Notification, error) {
	return Notification{}, nil
}
func (r *notificationRepositoryStub) FindByUser(context.Context, shared.UserID, page.Pagination) (page.Page[Notification], error) {
	return page.Page[Notification]{}, nil
}
func (r *notificationRepositoryStub) FindByType(context.Context, string, page.Pagination) (page.Page[Notification], error) {
	return page.Page[Notification]{}, nil
}
func (r *notificationRepositoryStub) FindByChannel(context.Context, string, page.Pagination) (page.Page[Notification], error) {
	return page.Page[Notification]{}, nil
}
func (r *notificationRepositoryStub) FindSummary(context.Context, string) ([]Notification, error) {
	return nil, nil
}

type emailSenderStub struct {
	reference  string
	err        error
	repository *notificationRepositoryStub
}

func (s emailSenderStub) Send(context.Context, *Notification) (string, error) {
	if !s.repository.created {
		return "", errors.New("delivery started before persistence")
	}
	return s.reference, s.err
}

type smsSenderStub struct{}

func (smsSenderStub) Send(context.Context, *Notification) (string, error) { return "", nil }

func TestSendPersistsThenMarksSent(t *testing.T) {
	repository := &notificationRepositoryStub{}
	service := NewNotificationService(repository, emailSenderStub{reference: "smtp-1", repository: repository}, smsSenderStub{})
	notification := &Notification{UserID: "7", UserEmail: "customer@example.com", Subject: "Subject", Message: "Message", Channel: NotificationChannelEmail}
	if err := service.Send(context.Background(), notification); err != nil {
		t.Fatal(err)
	}
	if !repository.updated || repository.updatedNotification.Status != DeliveryStatusSent || repository.updatedNotification.ProviderReference != "smtp-1" || repository.updatedNotification.SentAt == nil || repository.updatedNotification.Attempts != 1 {
		t.Fatalf("unexpected delivery state: %+v", repository.updatedNotification)
	}
}

func TestSendRetainsFailedDelivery(t *testing.T) {
	repository := &notificationRepositoryStub{}
	service := NewNotificationService(repository, emailSenderStub{err: errors.New("provider unavailable"), repository: repository}, smsSenderStub{})
	err := service.Send(context.Background(), &Notification{UserID: "7", UserEmail: "customer@example.com", Subject: "Subject", Message: "Message", Channel: NotificationChannelEmail})
	if err == nil || !repository.updated || repository.updatedNotification.Status != DeliveryStatusFailed || repository.updatedNotification.FailureMessage != "provider unavailable" || repository.updatedNotification.Attempts != 1 {
		t.Fatalf("unexpected failed delivery state: %+v, err=%v", repository.updatedNotification, err)
	}
}

func TestSendRejectsInvalidRecipientBeforePersistence(t *testing.T) {
	repository := &notificationRepositoryStub{}
	service := NewNotificationService(repository, emailSenderStub{repository: repository}, smsSenderStub{})
	if err := service.Send(context.Background(), &Notification{UserID: "7", Subject: "Subject", Message: "Message", Channel: NotificationChannelEmail}); err == nil || repository.created {
		t.Fatal("invalid email recipient must not create or deliver a notification")
	}
}
