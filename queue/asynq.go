package queue

import (
	"crypto/tls"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/dudeiebot/ad-ly/config"
	"github.com/dudeiebot/ad-ly/mailer"
)

var Client *asynq.Client

func Register() *asynq.ServeMux {
	redisAddr := fmt.Sprintf("%s:%s", config.DbConfig.RedisHost, config.DbConfig.RedisPort)

	redisConnection := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Username: config.DbConfig.RedisUser,
		Password: config.DbConfig.RedisPass,
	}

	if config.DbConfig.RedisScheme == "tls" {
		redisConnection.TLSConfig = &tls.Config{}
	}

	Client = asynq.NewClient(redisConnection)

	mux := asynq.NewServeMux()

	// SEND EMAIL
	mux.HandleFunc("send:email", mailer.HandleSendEmailTask)

	return mux
}
