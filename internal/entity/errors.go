package entity

import (
	"errors"
	"fmt"

	pb "github.com/llravell/go-pass/pkg/grpc"
)

var ErrUserConflict = errors.New("user with same login already exists")

var ErrPasswordDoesNotExist = errors.New("password does not exist")

var ErrCardDoesNotExist = errors.New("card does not exist")

var ErrPasswordAlreadyExist = errors.New("password with same name already exists")

var ErrNoSession = errors.New("user does not have active session")

var ErrUnknownConflict = errors.New("unknown conflict")

var ErrFileAlreadyUploading = errors.New("file has already uploading by another process")

type ConflictType string

const (
	DiffConflictType    ConflictType = "diff"
	DeletedConflictType ConflictType = "deleted"
)

type ConflictError[T SyncEntity] struct {
	_type    ConflictType
	actual   T
	incoming T
}

func NewDiffConflictError[T SyncEntity](actual, incoming T) *ConflictError[T] {
	return &ConflictError[T]{
		_type:    DiffConflictType,
		actual:   actual,
		incoming: incoming,
	}
}

func NewDeletedConflictError[T SyncEntity](actual, incoming T) *ConflictError[T] {
	return &ConflictError[T]{
		_type:    DeletedConflictType,
		actual:   actual,
		incoming: incoming,
	}
}

func (e *ConflictError[T]) Actual() T {
	return e.actual
}

func (e *ConflictError[T]) Incoming() T {
	return e.incoming
}

func (e *ConflictError[T]) Type() ConflictType {
	return e._type
}

func (e *ConflictError[T]) TypePB() pb.ConflictType {
	if e._type == DiffConflictType {
		return pb.ConflictType_DIFF
	}

	if e._type == DeletedConflictType {
		return pb.ConflictType_DELETED
	}

	return -1
}

func (e *ConflictError[T]) Error() string {
	return fmt.Sprintf(
		"%s conflict: actual (v%d) != incoming (v%d)",
		e.Type(),
		e.Actual().GetVersion(),
		e.Incoming().GetVersion(),
	)
}

func NewPasswordConflictErrorFromPB(actual *Password, conflict *pb.PasswordConflict) *ConflictError[*Password] {
	var conflictType ConflictType

	if conflict.GetType() == pb.ConflictType_DELETED {
		conflictType = DeletedConflictType
	} else if conflict.GetType() == pb.ConflictType_DIFF {
		conflictType = DiffConflictType
	}

	return &ConflictError[*Password]{
		_type:    conflictType,
		actual:   actual,
		incoming: NewPasswordFromPB(conflict.GetPassword()),
	}
}
