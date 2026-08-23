package handlers

import (
	"clinic-vet-api/internal/core/notifications"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/errors"
	"clinic-vet-api/internal/shared/http"
	"clinic-vet-api/internal/shared/page"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// NotificationHandler exposes notification read (monitoring) and manual send (write).
// Customers read only their notifications; staff can read by type/channel/id/summary and send.
type NotificationHandler struct {
	service   notifications.NotificationService
	validator *validator.Validate
}

func NewNotificationHandler(service notifications.NotificationService, validator *validator.Validate) *NotificationHandler {
	return &NotificationHandler{service: service, validator: validator}
}

// ------------------------------------------------------------
// Customer read (my notifications)
// ------------------------------------------------------------

// GetMyNotifications godoc
// @Summary      Get my notifications
// @Description  Returns paginated notifications for the authenticated user. Requires any authenticated role.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page_number  query  int  false  "Page number"  default(1)
// @Param        page_size    query  int  false  "Page size"    default(10)
// @Success      200          {object}  http.APIResponse{data=[]dtos.NotificationResponse}  "Paginated notifications"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /me/notifications [get]
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		userID, err := UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.service.GetNotificationsByUserID(ctx.Request.Context(), userID, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No notifications found")
			return
		}
		p := res.(page.Page[notifications.Notification])
		responsePage := mappers.NotificationsToResponsePage(p)
		http.Paginated(c, &responsePage, "Notifications")
	})(c)
}

// GetMyNotificationByID godoc
// @Summary      Get my notification by ID
// @Description  Returns a single notification by ID for the authenticated user. Requires any authenticated role.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Notification ID"
// @Success      200  {object}  http.APIResponse{data=dtos.NotificationResponse}  "Notification"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      403  {object}  http.APIResponse  "Forbidden"
// @Failure      404  {object}  http.APIResponse  "Not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /me/notifications/{id} [get]
func (h *NotificationHandler) GetMyNotificationByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		userID, err := UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		id := ctx.Param("id")
		if id == "" {
			return nil, shared.ErrEntityNotFound
		}
		notif, err := h.service.GetNotificationByID(ctx.Request.Context(), id)
		if err != nil {
			return nil, err
		}
		if notif.UserID != userID.String() {
			return nil, errors.ForbiddenError(ctx.Request.Context(), "get_notification", "notification", "notification does not belong to user")
		}
		return mappers.NotificationToResponse(notif), nil
	}
	http.HandleGetRequest(h.validator, "Notification", logic)(c)
}

// ------------------------------------------------------------
// Staff read (monitoring: by type, channel, id, summary)
// ------------------------------------------------------------

// GetNotificationsByType godoc
// @Summary      Get notifications by type (staff)
// @Description  Returns paginated notifications filtered by type. Requires veterinarian, receptionist, or admin role.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        type         path   string  true   "Notification type"
// @Param        page_number  query  int     false  "Page number"  default(1)
// @Param        page_size    query  int     false  "Page size"    default(10)
// @Success      200          {object}  http.APIResponse{data=[]dtos.NotificationResponse}  "Paginated notifications"
// @Failure      400          {object}  http.APIResponse  "Missing or invalid type"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /notifications/type/{type} [get]
func (h *NotificationHandler) GetNotificationsByType(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		notificationType := ctx.Param("type")
		if notificationType == "" {
			notificationType = ctx.Query("type")
		}
		if notificationType == "" {
			return nil, errors.MissingFieldError(ctx.Request.Context(), "type", "notification type is required", "get_notifications_by_type")
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.service.GetNotificationBySpecification(ctx.Request.Context(), notificationType, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No notifications found")
			return
		}
		p := res.(page.Page[notifications.Notification])
		responsePage := mappers.NotificationsToResponsePage(p)
		http.Paginated(c, &responsePage, "Notifications")
	})(c)
}

// GetNotificationsByChannel godoc
// @Summary      Get notifications by channel (staff)
// @Description  Returns paginated notifications filtered by delivery channel (email, sms). Requires veterinarian, receptionist, or admin role.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        channel      path   string  true   "Delivery channel (email, sms)"
// @Param        page_number  query  int     false  "Page number"  default(1)
// @Param        page_size    query  int     false  "Page size"    default(10)
// @Success      200          {object}  http.APIResponse{data=[]dtos.NotificationResponse}  "Paginated notifications"
// @Failure      400          {object}  http.APIResponse  "Missing or invalid channel"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /notifications/channel/{channel} [get]
func (h *NotificationHandler) GetNotificationsByChannel(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		channel := ctx.Param("channel")
		if channel == "" {
			channel = ctx.Query("channel")
		}
		if channel == "" {
			return nil, errors.MissingFieldError(ctx.Request.Context(), "channel", "channel is required", "get_notifications_by_channel")
		}
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, h.validator); err != nil {
			return nil, err
		}
		return h.service.GetNotificationByChannel(ctx.Request.Context(), channel, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(h.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No notifications found")
			return
		}
		p := res.(page.Page[notifications.Notification])
		responsePage := mappers.NotificationsToResponsePage(p)
		http.Paginated(c, &responsePage, "Notifications")
	})(c)
}

// GetNotificationByID godoc
// @Summary      Get notification by ID (staff)
// @Description  Returns a single notification by ID. Requires veterinarian, receptionist, or admin role.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Notification ID"
// @Success      200  {object}  http.APIResponse{data=dtos.NotificationResponse}  "Notification"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      404  {object}  http.APIResponse  "Not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /notifications/{id} [get]
func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		id := ctx.Param("id")
		if id == "" {
			return nil, shared.ErrEntityNotFound
		}
		notif, err := h.service.GetNotificationByID(ctx.Request.Context(), id)
		if err != nil {
			return nil, err
		}
		return mappers.NotificationToResponse(notif), nil
	}
	http.HandleGetRequest(h.validator, "Notification", logic)(c)
}

// GetNotificationSummary godoc
// @Summary      Get notification summary (staff)
// @Description  Returns a summary of sent notifications, optionally filtered by sender. Requires veterinarian, receptionist, or admin role.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        sender_id  query     string  false  "Sender ID to filter by"
// @Success      200        {object}  http.APIResponse{data=[]dtos.NotificationResponse}  "Notification summary"
// @Failure      401        {object}  http.APIResponse  "Unauthorized"
// @Failure      500        {object}  http.APIResponse  "Internal server error"
// @Router       /notifications/summary [get]
func (h *NotificationHandler) GetNotificationSummary(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		senderID := ctx.Query("sender_id")
		if senderID == "" {
			senderID = ctx.Param("sender_id")
		}
		summary, err := h.service.GenerateSummary(ctx.Request.Context(), senderID)
		if err != nil {
			return nil, err
		}
		return mappers.NotificationsToResponses(summary), nil
	}
	http.HandleGetRequest(h.validator, "NotificationSummary", logic)(c)
}

// ------------------------------------------------------------
// Staff write (manual send only)
// ------------------------------------------------------------

// SendNotification godoc
// @Summary      Send notification (staff)
// @Description  Manually sends a notification through the given channel (email, sms). Requires veterinarian, receptionist, or admin role.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.SendNotificationRequest  true  "Notification to send"
// @Success      200    {object}  http.APIResponse  "Notification sent"
// @Failure      400    {object}  http.APIResponse  "Validation error"
// @Failure      401    {object}  http.APIResponse  "Unauthorized"
// @Failure      500    {object}  http.APIResponse  "Internal server error"
// @Router       /notifications [post]
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.SendNotificationRequest) (any, error) {
		notif, err := mappers.SendNotificationRequestToNotification(req)
		if err != nil {
			return nil, err
		}
		senderID, err := UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		notif.SenderID = senderID.String()
		err = h.service.Send(ctx.Request.Context(), notif)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	http.HandleRequestWithResponder(h.validator, logic, func(c *gin.Context, _ any) {
		http.Success(c, nil, "Notification sent successfully")
	})(c)
}
