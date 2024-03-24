package user

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type Token struct {
	ID uuid.UUID
}

type Provider interface {
	User(ctx context.Context, token Token) (Descriptor, error)
}

type Descriptor struct {
	ID       uuid.UUID
	Username string
}
