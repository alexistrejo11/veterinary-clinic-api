package notifications

import (
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
	"context"
)

// entity.NotificationRepository defines methods for persisting and tracking notifications
//go:generate mockgen -source=repository.go -destination=../../test/mock/notification_repository_mock.go -package=mock

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	FindByID(ctx context.Context, id string) (Notification, error)
	FindByUser(ctx context.Context, userID shared.UserID, pagination page.Pagination) (page.Page[Notification], error)
	FindByType(ctx context.Context, notificationType string, pagination page.Pagination) (page.Page[Notification], error)
	FindByChannel(ctx context.Context, channel string, pagination page.Pagination) (page.Page[Notification], error)
	UpdateDelivery(ctx context.Context, notification *Notification) error
	FindSummary(ctx context.Context, senderID string) ([]Notification, error)
}
