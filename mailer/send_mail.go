package mailer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"net/textproto"
	"time"

	"github.com/Dudeiebot/dlog"
	"github.com/hibiken/asynq"

	"github.com/dudeiebot/ad-ly/config"
)

var logger = dlog.NewLog(dlog.LevelTrace)

const EnvProduction = "production"

const PostmartUri = "https://api.postmarkapp.com/email"

// Send dispatches an email to the specified recipient with the given subject, body, and attachments.
type EmailSender interface {
	Send(to, subject string, body []byte, attachments []*Attachment) error
}

type EmailPayload struct {
	TemplateName string
	To           string
	Subject      string
	Data         map[string]interface{}
	Attachments  []*Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Content     []byte
}

// local env
type mailhogSender struct{}

// pass ApiToken from production
type postmarkSender struct {
	ApiToken string
}

// mailhogSender
func (s *mailhogSender) Send(to, subject string, body []byte, attachments []*Attachment) error {
	addr := fmt.Sprintf("%s:%s", config.MailConfig.MailHost, config.MailConfig.MailPort)
	// mailhog doesnot require authentication
	return smtp.SendMail(addr, nil, config.MailConfig.MailFrom, []string{to}, body)
}

func (s *postmarkSender) Send(to, subject string, body []byte, attachments []*Attachment) error {
	reqBody := map[string]interface{}{
		"From":        config.MailConfig.MailFrom,
		"To":          to,
		"Subject":     subject,
		"HtmlBody":    string(body),
		"Attachments": buildPostmarkAttachments(attachments),
	}

	// query api link
	b, _ := json.Marshal(reqBody)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		PostmartUri,
		bytes.NewBuffer(b),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Postmark-Server-Token", s.ApiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("postmark failed with status: %w", resp.StatusCode)
	}

	return nil
}

func renderEmailTemplate(name string, data map[string]interface{}) (string, error) {
	tmpl, err := template.ParseFiles(fmt.Sprintf("./templates/email/%s.html", name))
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func buildPostmarkAttachments(attachments []*Attachment) []map[string]string {
	var out []map[string]string
	for _, a := range attachments {
		if a == nil {
			continue
		}
		out = append(out, map[string]string{
			"Name":        a.Filename,
			"Content":     base64.StdEncoding.EncodeToString(a.Content),
			"ContentType": a.ContentType,
		})
	}
	return out
}

func buildMimeMessage(p *EmailPayload, html string) ([]byte, error) {
	var msgBuffer bytes.Buffer
	writer := multipart.NewWriter(&msgBuffer)

	// Headers
	headers := map[string][]string{
		"From": {
			fmt.Sprintf("%s <%s>", config.AppConfig.AppName, config.MailConfig.MailFrom),
		},
		"To":           {p.To},
		"Subject":      {p.Subject},
		"Content-Type": {"multipart/mixed; boundary=" + writer.Boundary()},
	}

	for key, values := range headers {
		for _, value := range values {
			msgBuffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	msgBuffer.WriteString("\r\n")

	// HTML part
	htmlPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"text/html; charset=UTF-8"},
	})
	if err != nil {
		return nil, err
	}
	if _, err := htmlPart.Write([]byte(html)); err != nil {
		return nil, err
	}

	// Attachments
	for _, a := range p.Attachments {
		if a == nil {
			continue
		}
		encoded := base64.StdEncoding.EncodeToString(a.Content)
		header := textproto.MIMEHeader{
			"Content-Type":              {a.ContentType},
			"Content-Transfer-Encoding": {"base64"},
			"Content-Disposition":       {fmt.Sprintf("attachment; filename=\"%s\"", a.Filename)},
		}
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write([]byte(encoded)); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return msgBuffer.Bytes(), nil
}

func HandleSendEmailTask(ctx context.Context, t *asynq.Task) error {
	var p EmailPayload

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to Unmarshal payload: %w", err)
	}

	htmlBody, err := renderEmailTemplate(p.TemplateName, p.Data)
	if err != nil {
		return err
	}

	var sender EmailSender
	if config.AppConfig.AppHost == EnvProduction {
		// attachments is being built with the Send func here so we dont need to build mime message for this
		sender = &postmarkSender{ApiToken: config.MailConfig.MailToken}
		if err := sender.Send(p.To, p.Subject, []byte(htmlBody), p.Attachments); err != nil {
			logger.Error("Failed To send email", err)
			return err
		}
		logger.Info("Email sent successfully to: ", p.To)
	} else {
		// build mime message with attachments for postmark
		msgBody, err := buildMimeMessage(&p, htmlBody)
		if err != nil {
			return err
		}
		sender = &mailhogSender{}
		if err := sender.Send(p.To, p.Subject, msgBody, p.Attachments); err != nil {
			logger.Error("Failed To send email", err)
			return err
		}
		logger.Info("Email sent successfully to: ", p.To)
	}
	return nil
}

func EnqueueEmailTask(client *asynq.Client, payload EmailPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask("send:email", data)

	_, err = client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}
