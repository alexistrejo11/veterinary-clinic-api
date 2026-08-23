package mappers

import (
	"clinic-vet-api/internal/core/notifications"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/shared/page"
)

// SendNotificationRequestToNotification builds a Notification from the request for manual send.
func SendNotificationRequestToNotification(req dtos.SendNotificationRequest) (*notifications.Notification, error) {
	channel, err := notifications.ParseNotificationChannel(req.Channel)
	if err != nil {
		return nil, err
	}
	nType, err := notifications.ParseNotificationType(req.Type)
	if err != nil {
		return nil, err
	}
	return &notifications.Notification{
		UserID:    req.UserID,
		UserEmail: req.UserEmail,
		UserPhone: req.UserPhone,
		Title:     req.Title,
		Subject:   req.Subject,
		Message:   req.Message,
		Token:     req.Token,
		NType:     nType,
		Channel:   channel,
	}, nil
}

// NotificationToResponse maps a domain notification to its response DTO.
func NotificationToResponse(n notifications.Notification) dtos.NotificationResponse {
	return dtos.NotificationResponse{
		ID:                n.ID,
		UserID:            n.UserID,
		UserEmail:         n.UserEmail,
		UserPhone:         n.UserPhone,
		Title:             n.Title,
		Subject:           n.Subject,
		Message:           n.Message,
		Type:              n.NType.String(),
		Channel:           n.Channel.String(),
		Status:            n.Status.String(),
		FailureMessage:    n.FailureMessage,
		ProviderReference: n.ProviderReference,
		SentAt:            n.SentAt,
		Attempts:          n.Attempts,
		CreatedAt:         n.CreatedAt,
	}
}

// NotificationsToResponsePage maps a page of domain notifications to a page of response DTOs.
func NotificationsToResponsePage(p page.Page[notifications.Notification]) page.Page[dtos.NotificationResponse] {
	return page.MapItems(p, NotificationToResponse)
}

// NotificationsToResponses maps a slice of domain notifications to response DTOs.
func NotificationsToResponses(list []notifications.Notification) []dtos.NotificationResponse {
	responses := make([]dtos.NotificationResponse, len(list))
	for i, n := range list {
		responses[i] = NotificationToResponse(n)
	}
	return responses
}
