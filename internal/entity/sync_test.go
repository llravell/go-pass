package entity_test

import (
	"testing"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSyncEntity struct {
	version int
	deleted bool
}

func (e *testSyncEntity) GetVersion() int {
	return e.version
}

func (e *testSyncEntity) IsDeleted() bool {
	return e.deleted
}

var _ entity.SyncEntity = (*testSyncEntity)(nil)

func TestChooseMostActualEntity(t *testing.T) {
	t.Run("incoming entity would be more actual if has greater version", func(t *testing.T) {
		current := &testSyncEntity{}
		incoming := &testSyncEntity{}

		incoming.version++

		mostActualPass, err := entity.ChooseMostActualEntity(current, incoming)
		require.Nil(t, err)

		assert.Equal(t, incoming, mostActualPass)

		current.deleted = true
		mostActualPass, err = entity.ChooseMostActualEntity(current, incoming)
		require.Nil(t, err)

		assert.Equal(t, incoming, mostActualPass)
	})

	t.Run("will be deleted conflict if current entity already deleted and has greater version", func(t *testing.T) {
		current := &testSyncEntity{}
		incoming := &testSyncEntity{}

		current.version++
		current.deleted = true

		_, err := entity.ChooseMostActualEntity(current, incoming)
		require.NotNil(t, err)

		assert.Equal(t, entity.DeletedConflictType, err.Type())
		assert.Equal(t, current, err.Actual())
		assert.Equal(t, incoming, err.Incoming())
	})

	t.Run("will be diff conflict if current entity has greater version", func(t *testing.T) {
		current := &testSyncEntity{}
		incoming := &testSyncEntity{}

		current.version++

		_, err := entity.ChooseMostActualEntity(current, incoming)
		require.NotNil(t, err)

		assert.Equal(t, entity.DiffConflictType, err.Type())
		assert.Equal(t, current, err.Actual())
		assert.Equal(t, incoming, err.Incoming())
	})
}
