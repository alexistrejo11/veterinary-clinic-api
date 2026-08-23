package dtos

import (
	"time"

	"clinic-vet-api/internal/shared/page"
)

// CustomerResponse represents the customer data returned in HTTP responses
// @Description Customer entity: id, name, gender, date of birth, photo, user_id, is_active, timestamps.
type CustomerResponse struct {
	ID          uint      `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Gender      string    `json:"gender"`
	DateOfBirth time.Time `json:"date_of_birth"`
	PhotoURL    string    `json:"photo_url,omitempty"`
	UserID      uint      `json:"user_id,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CustomerSearchRequest represents filters for listing/searching customers
// @Description Pagination plus optional search text and is_active filter.
type CustomerSearchRequest struct {
	page.PaginationRequest

	// Optional text search over first and last name.
	Search string `json:"search,omitempty" form:"search"`

	// Optional active filter.
	IsActive *bool `json:"is_active,omitempty" form:"is_active"`
}

// CustomerCreateRequest represents the body for creating a new customer
// @Description First name, last name, email, gender, date of birth, optional photo URL, is_active, and optional user_id to link account.
type CustomerCreateRequest struct {
	FirstName   string  `json:"first_name" binding:"required"`
	LastName    string  `json:"last_name" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	Gender      string  `json:"gender" binding:"required,oneof=male female not_specified other"`
	DateOfBirth string  `json:"date_of_birth" binding:"required,datetime=2006-01-02"`
	PhotoURL    string  `json:"photo_url,omitempty" binding:"omitempty,url"`
	IsActive    *bool   `json:"is_active,omitempty"`
	UserID      *uint   `json:"user_id,omitempty"`
}

// CustomerUpdateRequest represents the body for updating an existing customer
// @Description ID (required) plus optional fields to update. ID can also be taken from URL path.
type CustomerUpdateRequest struct {
	ID          uint     `json:"id" binding:"required"`
	FirstName   *string  `json:"first_name,omitempty" binding:"omitempty"`
	LastName    *string  `json:"last_name,omitempty" binding:"omitempty"`
	Gender      *string  `json:"gender,omitempty" binding:"omitempty,oneof=male female not_specified other"`
	DateOfBirth *string  `json:"date_of_birth,omitempty" binding:"omitempty,datetime=2006-01-02"`
	PhotoURL    *string  `json:"photo_url,omitempty" binding:"omitempty,url"`
	UserID      *uint    `json:"user_id,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}
