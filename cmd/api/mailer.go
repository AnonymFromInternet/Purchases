package main

import (
	"bytes"
	"embed"
	"fmt"
	mail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"time"
)

//go:embed templates
var emailTemplateFS embed.FS

func (application *application) SendEmail(from, to, subject, tmplName string, tmplData interface{}) error {
	templateToRender := fmt.Sprintf("templates/%s.gohtml", tmplName)
	tmpl, err := template.New("email-html").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		return err
	}

	var tpl bytes.Buffer

	// in ExecuteTemplate, param "name" is about section name in the template ({{define "sectionName"}})
	err = tmpl.ExecuteTemplate(&tpl, "body", tmplData)
	if err != nil {
		return err
	}

	htmlMessageAsString := tpl.String()

	templateToRender = fmt.Sprintf("templates/%s.plain.gohtml", tmplName)
	tmpl, err = template.New("email-plain").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(&tpl, "body", tmplData)
	if err != nil {
		return err
	}

	plainMessageAsString := tpl.String()

	// send email
	smtpServer := mail.NewSMTPClient()
	smtpServer.Host = application.config.smtp.host
	smtpServer.Port = application.config.smtp.port
	smtpServer.Username = application.config.smtp.username
	smtpServer.Password = application.config.smtp.password
	smtpServer.Encryption = mail.EncryptionTLS
	smtpServer.KeepAlive = false
	smtpServer.ConnectTimeout = 5 * time.Second
	smtpServer.SendTimeout = 5 * time.Second

	smtpClient, err := smtpServer.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(subject)
	email.SetBody(mail.TextHTML, htmlMessageAsString)
	email.AddAlternative(mail.TextPlain, plainMessageAsString)

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	application.infoLog.Println("send mail")

	return nil
}
