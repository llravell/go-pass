package entity

type SyncUpdates[T any] struct {
	ToAdd    []T
	ToUpdate []T
	ToSync   []T
}

type SyncEntity interface {
	IsDeleted() bool
	GetVersion() int
}

type SyncEntityPointer[T any] interface {
	*T
	SyncEntity
}

func ChooseMostActuralEntity[T any, PT SyncEntityPointer[T]](current, incoming PT) (PT, *ConflictError[PT]) {
	if current.IsDeleted() {
		if incoming.GetVersion() > current.GetVersion() {
			return incoming, nil
		}

		return nil, NewDeletedConflictError(current, incoming)
	}

	if incoming.GetVersion() > current.GetVersion() {
		return incoming, nil
	}

	return nil, NewDiffConflictError(current, incoming)
}
