package usecase

import (
	"context"
	"errors"

	"github.com/ralsnet/grepo"
	"github.com/ralsnet/grepo/example/entity"
	"github.com/ralsnet/grepo/example/port"
)

const GetUserOperation = "GetUser"

type GetUserInput struct {
	ID string
}

type GetUserOutput struct {
	User *entity.User
}

type GetUser struct {
	repoUser port.RepoUser
}

func NewGetUser(repoUser port.RepoUser) grepo.Executor[GetUserInput, GetUserOutput] {
	return &GetUser{
		repoUser: repoUser,
	}
}

func (uc *GetUser) Execute(ctx context.Context, input GetUserInput) (*GetUserOutput, error) {
	user, err := uc.repoUser.GetUser(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return &GetUserOutput{
		User: user,
	}, nil
}
