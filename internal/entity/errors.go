package entity

import (
	"errors"
	"fmt"

	pb "github.com/llravell/go-pass/pkg/grpc"
)

var ErrUserConflict = errors.New("user with same login already exists")

var ErrPasswordDoesNotExist = errors.New("password does not exist")

var ErrPasswordAlreadyExist = errors.New("password with same name already exists")

var ErrNoSession = errors.New("user does not have active session")

var ErrUnknownConflict = errors.New("unknown conflict")

type PasswordConflictType string

const (
	PasswordDiffConflictType    PasswordConflictType = "diff"
	PasswordDeletedConflictType PasswordConflictType = "deleted"
)

type PasswordConflictError struct {
	_type    PasswordConflictType
	actual   *Password
	incoming *Password
}

func NewPasswordConflictErrorFromPB(actual *Password, conflict *pb.PasswordConflict) *PasswordConflictError {
	var conflictType PasswordConflictType

	if conflict.GetType() == pb.ConflictType_DELETED {
		conflictType = PasswordDeletedConflictType
	} else if conflict.GetType() == pb.ConflictType_DIFF {
		conflictType = PasswordDiffConflictType
	}

	return &PasswordConflictError{
		_type:    conflictType,
		actual:   actual,
		incoming: NewPasswordFromPB(conflict.GetPassword()),
	}
}

func NewPasswordDiffConflictError(actual, incoming *Password) *PasswordConflictError {
	return &PasswordConflictError{
		_type:    PasswordDiffConflictType,
		actual:   actual,
		incoming: incoming,
	}
}

func NewPasswordDeletedConflictError(actual, incoming *Password) *PasswordConflictError {
	return &PasswordConflictError{
		_type:    PasswordDeletedConflictType,
		actual:   actual,
		incoming: incoming,
	}
}

func (e *PasswordConflictError) Actual() *Password {
	return e.actual
}

func (e *PasswordConflictError) Incoming() *Password {
	return e.incoming
}

func (e *PasswordConflictError) Type() PasswordConflictType {
	return e._type
}

func (e *PasswordConflictError) TypePB() pb.ConflictType {
	if e._type == PasswordDiffConflictType {
		return pb.ConflictType_DIFF
	}

	if e._type == PasswordDeletedConflictType {
		return pb.ConflictType_DELETED
	}

	return -1
}

func (e *PasswordConflictError) Error() string {
	return fmt.Sprintf("%s conflict: actual (v%d) != incoming (v%d)", e.Type(), e.Actual().Version, e.Incoming().Version)
}
