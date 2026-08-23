package shared

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrEntityNotFound = errors.New("entity not supported")

type IntegerID interface {
	Value() uint
	Equals(number uint) bool
	String() string
	IsZero() bool
	Int32() int32
}

type BaseID struct {
	value uint
}

func NewBaseIDEmpty() IntegerID {
	return NewBaseID(0)
}

func NewBaseID(value uint) IntegerID {
	return BaseID{value: value}
}

func (id BaseID) Value() uint {
	return id.value
}

func (id BaseID) Equals(number uint) bool {
	return id.Value() == number
}

func (id BaseID) String() string {
	return fmt.Sprintf("%d", id.Value())
}

func (id BaseID) IsZero() bool {
	return id.Value() == 0
}

func (id BaseID) Int32() int32 {
	return int32(id.Value())
}

// USER ID Moved here to avoid import cycles conflicts
type UserID struct{ BaseID }

func NewUserID(value uint) UserID {
	return UserID{BaseID{value: value}}
}

func ParseUserIDFromString(idStr string) (UserID, error) {
	idStr = strings.TrimSpace(idStr)
	if idStr == "" {
		return UserID{}, fmt.Errorf("user ID cannot be empty")
	}

	intValue, err := strconv.Atoi(idStr)
	if err != nil {
		return UserID{}, fmt.Errorf("user ID must be a valid number")
	}

	if intValue < 0 {
		return UserID{}, fmt.Errorf("user ID cannot be negative")
	}

	return NewUserID(uint(intValue)), nil
}
