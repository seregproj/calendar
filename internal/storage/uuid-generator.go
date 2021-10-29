package storage

import (
	"fmt"

	"github.com/gofrs/uuid"
)

type UUIDGen struct{}

func NewUUIDGen() *UUIDGen {
	return &UUIDGen{}
}

func (u *UUIDGen) Generate() (string, error) {
	euuid, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("cant generate uuid: %w", err)
	}

	return euuid.String(), nil
}
