package notify

/*
File for sending notification emails
*/

import (
	"fmt"
	"log"

	"github.com/sfs/pkg/configs"
	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/logger"

	"gopkg.in/gomail.v2"
)

type Email struct {
	Email    string `env:"CLIENT_EMAIL"`
	Password string `env:"CLIENT_PASSWORD"`
	Host     string // smtp host
	Port     int    // smtp port
	log      *logger.Logger
}

func NewEmail() *Email {
	envCfgs := env.NewE()

	email, err := envCfgs.Get(configs.CLIENT_EMAIL)
	if err != nil {
		log.Fatal(err)
	}
	pw, err := envCfgs.Get(configs.CLIENT_PASSWORD)
	if err != nil {
		log.Fatal(err)
	}
	return &Email{
		Email:    email,
		Password: pw,
		Host:     "smtp.gmail.com",
		Port:     587,
		log:      logger.NewLogger("EMAIL_SENDER", "None"),
	}
}

func SendEmail(to string, subject string, message string) error {
	email := NewEmail()

	msg := gomail.NewMessage()
	msg.SetHeader("From", email.Email)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", message)

	dialer := gomail.NewDialer(email.Host, email.Port, email.Email, email.Password)
	if err := dialer.DialAndSend(msg); err != nil {
		email.log.Error(err.Error())
	}
	email.log.Info(fmt.Sprintf("sent email to: %s", to))

	return nil
}
