package broker

import (
	"testing"

	"github.com/heroku/busl/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/heroku/busl/util"
)

func newRegUUID() (*RedisRegistrar, string) {
	reg := NewRedisRegistrar()
	uuid, _ := util.NewUUID()

	return reg, uuid
}

func TestRegisteredIsRegistered(t *testing.T) {
	reg, uuid := newRegUUID()
	reg.Register(uuid)
	assert.True(t, reg.IsRegistered(uuid))
}

func TestUnregisteredIsNotRegistered(t *testing.T) {
	reg, uuid := newRegUUID()
	assert.False(t, reg.IsRegistered(uuid))
}

func TestUnregisteredErrNotRegistered(t *testing.T) {
	_, uuid := newRegUUID()

	_, err := NewReader(uuid)
	assert.Equal(t, err, ErrNotRegistered)

	_, err = NewWriter(uuid)
	assert.Equal(t, err, ErrNotRegistered)
}

func TestRegisteredNoError(t *testing.T) {
	reg, uuid := newRegUUID()
	reg.Register(uuid)
	_, err := NewReader(uuid)
	assert.Nil(t, err)

	_, err = NewWriter(uuid)
	assert.Nil(t, err)
}
