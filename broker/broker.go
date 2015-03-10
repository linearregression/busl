package broker

import (
	"github.com/heroku/busl/util"
)

type Registrar interface {
	Register(id util.UUID) error
	IsRegistered(id util.UUID) bool
}
