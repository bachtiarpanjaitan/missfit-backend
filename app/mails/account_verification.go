package mails

import (
	"lumos/app/facades"
	"lumos/app/models"

	"github.com/goravel/framework/contracts/mail"
)

type AccountVerification struct {
	VerificationLink string
	User             models.User
}

// Headers add custom headers to the mail.
func (receiver *AccountVerification) Headers() map[string]string {
	return map[string]string{
		"X-Mailer": "Ihand Lumos",
	}
}

// Attachments attach files to the mail
func (receiver *AccountVerification) Attachments() []string {
	return []string{}
}

// Content set the content of the mail
func (receiver *AccountVerification) Content() *mail.Content {
	return &mail.Content{
		View: "account_verification.tmpl",
		With: map[string]interface{}{
			"verificationLink": receiver.VerificationLink,
			"name":             receiver.User.Name,
		},
	}
}

// Envelope set the envelope of the mail
func (receiver *AccountVerification) Envelope() *mail.Envelope {
	fromAddress := facades.Config().GetString("mail.from.address")
	fromName := facades.Config().GetString("mail.from.name")
	return &mail.Envelope{
		Subject: "Verifikasi Akun Ihand Lumos",
		From: mail.Address{
			Address: fromAddress,
			Name:    fromName,
		},
		To: []string{receiver.User.Email},
	}
}

// Queue set the queue of the mail
func (receiver *AccountVerification) Queue() *mail.Queue {
	return &mail.Queue{
		Connection: "database",
		Queue:      "default",
	}
}
