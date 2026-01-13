package port

import (
	"context"

	"github.com/ralsnet/grepo/example/entity"
)

type FindUsersFilter struct {
	IDs  []string
	Name string
}

type RepoUser interface {
	FindUsers(ctx context.Context, filter FindUsersFilter) ([]*entity.User, error)
	GetUser(ctx context.Context, id string) (*entity.User, error)
	SaveUser(ctx context.Context, user *entity.User) error
}
