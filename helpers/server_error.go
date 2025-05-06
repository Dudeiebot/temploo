package helpers

import (
	"github.com/dudeiebot/ad-ly/config"
	"github.com/dudeiebot/ad-ly/errors"
)

var EnvProduction string = "production"

func ServerError(err error) error {
	if config.AppConfig.AppHost == EnvProduction {
		return errors.ErrSomethingWentWrong
	}
	return err
}
