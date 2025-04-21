package entity_test

import (
	"testing"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/stretchr/testify/assert"
)

var pass = &entity.Password{
	Name:    "test",
	Value:   "value",
	Meta:    "meta",
	Version: 1,
}

func TestPassword_Clone(t *testing.T) {
	clone := pass.Clone()

	assert.Equal(t, pass.Name, clone.Name)
	assert.Equal(t, pass.Value, clone.Value)
	assert.Equal(t, pass.Meta, clone.Meta)
	assert.Equal(t, pass.Version, clone.Version)
	assert.Equal(t, pass.Deleted, clone.Deleted)
}

func TestPassword_Equal(t *testing.T) {
	mutations := map[string]func(*entity.Password){
		"name": func(p *entity.Password) {
			p.Name = "other name"
		},
		"value": func(p *entity.Password) {
			p.Value = "other value"
		},
		"meta": func(p *entity.Password) {
			p.Meta = "other meta"
		},
		"version": func(p *entity.Password) {
			p.Version = 10
		},
	}

	for field, mutation := range mutations {
		t.Run("false if password has different "+field, func(t *testing.T) {
			clone := pass.Clone()

			mutation(clone)

			assert.False(t, pass.Equal(clone))
		})
	}

	t.Run("true if password has same name, value, meta and version", func(t *testing.T) {
		assert.True(t, pass.Equal(pass.Clone()))
	})
}
