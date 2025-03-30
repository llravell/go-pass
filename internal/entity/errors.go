package entity

import "errors"

var ErrUserConflict = errors.New("user with same login already exists")

var ErrPasswordDoesNotExist = errors.New("password does not exist")

var ErrPasswordAlreadyExist = errors.New("password with same name already exists")

var ErrNoSession = errors.New("user does not have active session")
