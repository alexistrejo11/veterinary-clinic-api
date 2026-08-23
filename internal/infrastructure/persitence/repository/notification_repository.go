package repository

import (
	"clinic-vet-api/database/sqlc"
	"clinic-vet-api/internal/core/notifications"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/mapper"
	"clinic-vet-api/internal/shared/page"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	customErr "clinic-vet-api/internal/shared/errors"
)

type SqlcNotificationRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewSqlcNotificationRepository(queries *sqlc.Queries) notifications.NotificationRepository {
	return &SqlcNotificationRepository{
		queries: queries,
		pgMap:   mapper.NewSqlcFieldMapper(),
	}
}

func (r *SqlcNotificationRepository) Create(ctx context.Context, notification *notifications.Notification) error {
	userID, err := parseUserIDFromNotification(notification.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id for notification: %w", err)
	}
	params := sqlc.CreateNotificationParams{
		UserID:           userID,
		SenderID:         optionalUserID(notification.SenderID),
		Title:            notification.Title,
		Subject:          notification.Subject,
		Message:          notification.Message,
		Token:            notification.Token,
		NotificationType: string(notification.NType),
		Channel:          string(notification.Channel),
	}
	created, err := r.queries.CreateNotification(ctx, params)
	if err != nil {
		return r.dbError(OpInsert, ErrMsgCreateNotification, err)
	}
	notification.ID = strconv.Itoa(int(created.ID))
	notification.CreatedAt = r.pgMap.PgTimestamptz.ToTime(created.CreatedAt)
	notification.Status = notifications.DeliveryStatus(created.Status)
	notification.Attempts = int(created.Attempts)
	return nil
}

func (r *SqlcNotificationRepository) UpdateDelivery(ctx context.Context, notification *notifications.Notification) error {
	id, err := strconv.ParseInt(notification.ID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid notification ID: %w", err)
	}
	params := sqlc.UpdateNotificationDeliveryParams{ID: int32(id), Status: string(notification.Status), FailureMessage: optionalText(notification.FailureMessage), ProviderReference: optionalText(notification.ProviderReference)}
	if notification.SentAt != nil {
		params.SentAt = pgtype.Timestamptz{Time: *notification.SentAt, Valid: true}
	}
	if err := r.queries.UpdateNotificationDelivery(ctx, params); err != nil {
		return r.dbError(OpUpdate, "failed to update notification delivery", err)
	}
	return nil
}

func (r *SqlcNotificationRepository) FindSummary(ctx context.Context, senderID string) ([]notifications.Notification, error) {
	sender := int32(0)
	if senderID != "" {
		value, err := strconv.ParseInt(senderID, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid sender ID: %w", err)
		}
		sender = int32(value)
	}
	rows, err := r.queries.FindNotificationSummary(ctx, sender)
	if err != nil {
		return nil, r.dbError(OpSelect, "failed to find notification summary", err)
	}
	items := make([]notifications.Notification, len(rows))
	for i, row := range rows {
		items[i] = r.toEntity(notificationRow(row.ID, row.UserID, row.SenderID, row.Title, row.Subject, row.Message, row.Token, row.NotificationType, row.Channel, row.Status, row.FailureMessage, row.ProviderReference, row.SentAt, row.Attempts, row.CreatedAt, row.UpdatedAt))
	}
	return items, nil
}

func (r *SqlcNotificationRepository) FindByID(ctx context.Context, id string) (notifications.Notification, error) {
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return notifications.Notification{}, r.notFoundError("id", id)
	}
	row, err := r.queries.GetNotificationByID(ctx, int32(idInt))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return notifications.Notification{}, r.notFoundError("id", id)
		}
		return notifications.Notification{}, r.dbError(OpSelect, ErrMsgFindNotificationByID, err)
	}
	return r.toEntity(notificationRow(row.ID, row.UserID, row.SenderID, row.Title, row.Subject, row.Message, row.Token, row.NotificationType, row.Channel, row.Status, row.FailureMessage, row.ProviderReference, row.SentAt, row.Attempts, row.CreatedAt, row.UpdatedAt)), nil
}

func (r *SqlcNotificationRepository) FindByUser(ctx context.Context, userID shared.UserID, pagination page.Pagination) (page.Page[notifications.Notification], error) {
	params := sqlc.FindNotificationsByUserIDParams{
		UserID: userID.Int32(),
		Limit:  pagination.Limit(),
		Offset: pagination.Offset(),
	}
	rows, err := r.queries.FindNotificationsByUserID(ctx, params)
	if err != nil {
		return page.Page[notifications.Notification]{}, r.dbError(OpSelect, ErrMsgFindNotificationsByUser, err)
	}
	total, err := r.queries.CountNotificationsByUserID(ctx, userID.Int32())
	if err != nil {
		return page.Page[notifications.Notification]{}, r.dbError(OpCount, ErrMsgFindNotificationsByUser, err)
	}
	items := make([]notifications.Notification, len(rows))
	for i := range rows {
		row := rows[i]
		items[i] = r.toEntity(notificationRow(row.ID, row.UserID, row.SenderID, row.Title, row.Subject, row.Message, row.Token, row.NotificationType, row.Channel, row.Status, row.FailureMessage, row.ProviderReference, row.SentAt, row.Attempts, row.CreatedAt, row.UpdatedAt))
	}
	pagReq := page.PaginationRequest{
		Page:     int32(pagination.Number),
		PageSize: pagination.Limit(),
	}
	return page.NewPage(items, total, pagReq), nil
}

func (r *SqlcNotificationRepository) FindByType(ctx context.Context, notificationType string, pagination page.Pagination) (page.Page[notifications.Notification], error) {
	params := sqlc.FindNotificationsByTypeParams{
		NotificationType: notificationType,
		Limit:            pagination.Limit(),
		Offset:           pagination.Offset(),
	}
	rows, err := r.queries.FindNotificationsByType(ctx, params)
	if err != nil {
		return page.Page[notifications.Notification]{}, r.dbError(OpSelect, "failed to find notifications by type", err)
	}
	total, err := r.queries.CountNotificationsByType(ctx, notificationType)
	if err != nil {
		return page.Page[notifications.Notification]{}, r.dbError(OpCount, "failed to count notifications by type", err)
	}
	items := make([]notifications.Notification, len(rows))
	for i := range rows {
		row := rows[i]
		items[i] = r.toEntity(notificationRow(row.ID, row.UserID, row.SenderID, row.Title, row.Subject, row.Message, row.Token, row.NotificationType, row.Channel, row.Status, row.FailureMessage, row.ProviderReference, row.SentAt, row.Attempts, row.CreatedAt, row.UpdatedAt))
	}
	pagReq := page.PaginationRequest{
		Page:     int32(pagination.Number),
		PageSize: pagination.Limit(),
	}
	return page.NewPage(items, total, pagReq), nil
}

func (r *SqlcNotificationRepository) FindByChannel(ctx context.Context, channel string, pagination page.Pagination) (page.Page[notifications.Notification], error) {
	params := sqlc.FindNotificationsByChannelParams{
		Channel: channel,
		Limit:   pagination.Limit(),
		Offset:  pagination.Offset(),
	}
	rows, err := r.queries.FindNotificationsByChannel(ctx, params)
	if err != nil {
		return page.Page[notifications.Notification]{}, r.dbError(OpSelect, "failed to find notifications by channel", err)
	}
	total, err := r.queries.CountNotificationsByChannel(ctx, channel)
	if err != nil {
		return page.Page[notifications.Notification]{}, r.dbError(OpCount, "failed to count notifications by channel", err)
	}
	items := make([]notifications.Notification, len(rows))
	for i := range rows {
		row := rows[i]
		items[i] = r.toEntity(notificationRow(row.ID, row.UserID, row.SenderID, row.Title, row.Subject, row.Message, row.Token, row.NotificationType, row.Channel, row.Status, row.FailureMessage, row.ProviderReference, row.SentAt, row.Attempts, row.CreatedAt, row.UpdatedAt))
	}
	pagReq := page.PaginationRequest{
		Page:     int32(pagination.Number),
		PageSize: pagination.Limit(),
	}
	return page.NewPage(items, total, pagReq), nil
}

func (r *SqlcNotificationRepository) toEntity(row sqlc.Notification) notifications.Notification {
	nType, _ := notifications.ParseNotificationType(row.NotificationType)
	channel, _ := notifications.ParseNotificationChannel(row.Channel)
	return notifications.Notification{
		ID:                strconv.Itoa(int(row.ID)),
		UserID:            strconv.Itoa(int(row.UserID)),
		UserEmail:         "",
		UserPhone:         "",
		SenderID:          optionalUserIDString(row.SenderID),
		Title:             row.Title,
		Subject:           row.Subject,
		Message:           row.Message,
		Token:             row.Token,
		CreatedAt:         r.pgMap.PgTimestamptz.ToTime(row.CreatedAt),
		NType:             nType,
		Channel:           channel,
		Status:            notifications.DeliveryStatus(row.Status),
		FailureMessage:    row.FailureMessage.String,
		ProviderReference: row.ProviderReference.String,
		SentAt:            optionalTime(row.SentAt),
		Attempts:          int(row.Attempts),
	}
}

func optionalUserID(value string) pgtype.Int4 {
	id, err := strconv.ParseInt(value, 10, 32)
	if value == "" || err != nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(id), Valid: true}
}
func optionalText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func optionalTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func optionalUserIDString(value pgtype.Int4) string {
	if !value.Valid {
		return ""
	}
	return strconv.Itoa(int(value.Int32))
}

func notificationRow(id, userID int32, senderID pgtype.Int4, title, subject, message, token, notificationType, channel, status string, failureMessage, providerReference pgtype.Text, sentAt pgtype.Timestamptz, attempts int32, createdAt, updatedAt pgtype.Timestamptz) sqlc.Notification {
	return sqlc.Notification{ID: id, UserID: userID, SenderID: senderID, Title: title, Subject: subject, Message: message, Token: token, NotificationType: notificationType, Channel: channel, Status: status, FailureMessage: failureMessage, ProviderReference: providerReference, SentAt: sentAt, Attempts: attempts, CreatedAt: createdAt, UpdatedAt: updatedAt}
}

func (r *SqlcNotificationRepository) dbError(operation, message string, err error) error {
	return customErr.DatabaseError(operation, TableNotifications, DriverSQL, fmt.Errorf("%s: %w", message, err))
}

func (r *SqlcNotificationRepository) notFoundError(parameterName, parameterValue string) error {
	return customErr.DBNotFoundError(parameterName, parameterValue, OpSelect, TableNotifications, DriverSQL)
}

// parseUserIDFromNotification parses domain UserID string (e.g. from Notification) to int32 for DB.
func parseUserIDFromNotification(userID string) (int32, error) {
	if userID == "" {
		return 0, fmt.Errorf("user_id is required")
	}
	id, err := strconv.ParseInt(userID, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}
