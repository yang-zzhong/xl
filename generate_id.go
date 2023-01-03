package xl

import (
	"github.com/google/uuid"
)

type IdGenerator interface {
	GenerateId() string
}

type generateId func() string

var (
	GenerateUUID = generateId(func() string {
		return uuid.New().String()
	})
)

func (g generateId) GenerateId() string {
	return g()
}
