package example

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ralsnet/grepo"
	"github.com/ralsnet/grepo/example/usecase"
	"github.com/ralsnet/grepo/hooks"
	"github.com/ralsnet/grepo/refl"
)

func NewAPI(
	findUser grepo.Executor[usecase.FindUsersInput, usecase.FindUsersOutput],
	getUser grepo.Executor[usecase.GetUserInput, usecase.GetUserOutput],
	saveUser grepo.Executor[usecase.SaveUserInput, usecase.SaveUserOutput],
) *grepo.API {
	return grepo.NewAPIBuilder().
		WithDescription("API example").
		AddBeforeHook(hooks.HookBeforeSlog()).
		AddAfterHook(hooks.HookAfterSlog()).
		AddErrorHook(hooks.HookErrorSlog()).
		WithOptions(
			grepo.WithEnableInputValidation(),
			grepo.WithEnableOutputValidation(),
			grepo.WithCustomFieldValidators((grepo.FieldValidatorFunc(func(v reflect.Value, f *refl.Field) error {
				fmt.Println(f.Parent().Name, f.Field, f.Type.Name)
				return nil
			}))),
		).
		AddUseCase(
			grepo.NewUseCaseBuilder(findUser).
				Build(),
		).
		AddUseCase(
			grepo.NewUseCaseBuilder(getUser).
				Build(),
		).
		AddUseCase(
			grepo.NewUseCaseBuilder(saveUser).
				AddBeforeHook(func(ctx context.Context, i *usecase.SaveUserInput) (context.Context, error) {
					if i.Authority != "admin" && i.Authority != "user" {
						i.Authority = "user"
					}
					return ctx, nil
				}).
				Build(),
		).
		Build()
}
