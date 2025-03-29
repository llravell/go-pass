package entity

import "errors"

var ErrUserConflict = errors.New("user with same login already exists")
