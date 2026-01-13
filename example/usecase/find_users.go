package usecase

import (
	"context"

	"github.com/ralsnet/grepo"
	"github.com/ralsnet/grepo/example/entity"
	"github.com/ralsnet/grepo/example/port"
)

const FindUsersOperation = "FindUsers"

type FindUsersInput struct {
	IDs  []string `grepo:"optional:true"`
	Name string   `grepo:"optional:true"`
}

type FindUsersOutput struct {
	Users []*entity.User `grepo:"optional:true"`
}

type FindUsers struct {
	repoUser port.RepoUser
}

func NewFindUsers(repoUser port.RepoUser) grepo.Executor[FindUsersInput, FindUsersOutput] {
	return &FindUsers{
		repoUser: repoUser,
	}
}

func (uc *FindUsers) Execute(ctx context.Context, input FindUsersInput) (*FindUsersOutput, error) {
	users, err := uc.repoUser.FindUsers(ctx, port.FindUsersFilter{
		IDs:  input.IDs,
		Name: input.Name,
	})
	if err != nil {
		return nil, err
	}

	return &FindUsersOutput{
		Users: users,
	}, nil
}
