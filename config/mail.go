package config

import (
	"errors"
	"os"
)

// we checking if it is local, prog or staging
// local and staging use mailhog and prod uses postmark
// comback

var MailConfig MailEnv

type MailEnv struct {
	MailServer string
	MailPort   string
	MailFrom   string
	MailHost   string
	MailToken  string
}

func loadMailEnv() error {
	mailServer, exists := os.LookupEnv("MAIL_SERVER")
	if !exists {
		return errors.New("MAIL_SERVER not in .env")
	}

	mailPort, exists := os.LookupEnv("MAIL_PORT")
	if !exists {
		return errors.New("MAIL_PORT not in .env")
	}

	mailFrom, exists := os.LookupEnv("MAIL_FROM")
	if !exists {
		return errors.New("MAIL_FROM not in .env")
	}

	mailHost, exists := os.LookupEnv("MAIL_HOST")
	if !exists {
		return errors.New("MAIL_HOST not in .env")
	}

	mailToken, exists := os.LookupEnv("MAIL_TOKEN")
	if !exists {
		return errors.New("MAIL_TOKEN not in .env")
	}

	MailConfig = MailEnv{
		MailServer: mailServer,
		MailPort:   mailPort,
		MailFrom:   mailFrom,
		MailHost:   mailHost,
		MailToken:  mailToken,
	}

	return nil
}
