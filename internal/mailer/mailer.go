package mailer

import (
	"bytes"
	"html/template"

	"github.com/Stefanuswilfrid/course-backend/internal/config"
	"gopkg.in/gomail.v2"
)

func GenerateMail(recipientEmail, subject, templateStr string, data map[string]any) (*gomail.Message, error) {
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return nil, err
	}

	var tmplOutput bytes.Buffer
	err = tmpl.Execute(&tmplOutput, data)
	if err != nil {
		return nil, err
	}

	mail := gomail.NewMessage()
	mail.SetHeader("From", "Seatudy <"+config.Env.SmtpEmail+">")
	mail.SetHeader("To", recipientEmail)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", tmplOutput.String())

	return mail, nil
}
