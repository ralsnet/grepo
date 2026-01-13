package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/ralsnet/grepo"
	"github.com/ralsnet/grepo/example/entity"
	"github.com/ralsnet/grepo/example/port"
)

const SaveUserOperation = "SaveUser"

type SaveUserInput struct {
	Name      string
	Authority string `grepo:"enum:admin,user"`
}

type SaveUserOutput struct {
	User *entity.User
}

type SaveUser struct {
	repoUser port.RepoUser
}

func NewSaveUser(repoUser port.RepoUser) grepo.Executor[SaveUserInput, SaveUserOutput] {
	return &SaveUser{
		repoUser: repoUser,
	}
}

func (uc *SaveUser) Execute(ctx context.Context, input SaveUserInput) (*SaveUserOutput, error) {
	uuidv7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	now := grepo.ExecuteTime(ctx)

	user := &entity.User{
		ID:        uuidv7.String(),
		Name:      input.Name,
		Authority: entity.Authority(input.Authority),
		Groups:    make([]*entity.Group, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.repoUser.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	return &SaveUserOutput{
		User: user,
	}, nil
}
