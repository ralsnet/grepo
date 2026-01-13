package internal

import (
	"github.com/ralsnet/grepo"
	"github.com/ralsnet/grepo/example"
	"github.com/ralsnet/grepo/example/internal/local"
	"github.com/ralsnet/grepo/example/usecase"
)

func InitializeAPI() *grepo.API {
	repoUser := local.NewRepoUser(".")

	ucFindUser := usecase.NewFindUsers(repoUser)
	ucGetUser := usecase.NewGetUser(repoUser)
	ucSaveUser := usecase.NewSaveUser(repoUser)

	return example.NewAPI(ucFindUser, ucGetUser, ucSaveUser)
}
