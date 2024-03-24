//nolint:revive,unused
package ldap

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"guardian/internal/guardian/app/user"
)

func NewUserProvider(address string) user.Provider {
	return &userProvider{}
}

type userProvider struct {
	address string
}

func (provider *userProvider) User(ctx context.Context, token user.Token) (user.Descriptor, error) {
	return user.Descriptor{
		ID:       uuid.Must(uuid.FromString("e1790eb1-e4dd-49ea-9e55-6132a6446d55")),
		Username: "vadim.makerov",
	}, nil
}
