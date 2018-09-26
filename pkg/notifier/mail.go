package notifier

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Scalify/website-content-watcher/pkg/api"
	"gopkg.in/gomail.v2"
)

var mailTemplate = `
<html>
<head>
</head>
<body style="font-family: Arial">
Hi.<br />
<br />
You are receiving this mail because you registered to get updates on job <i>{{ .JobName }}</i>.<br />
<br />

<table cellpadding="0" border="1">
	<tr>
		<th>Name/Item</th>
		<th>Old/new value</th>
	</tr>
{{ range .Diff }}
	<tr>
		<td valign="top">{{ .Item }}</td>
		<td valign="top">
			Old:  {{ .OldValue }}<br />
			New: {{ .NewValue }}
		</td>
	</tr>
{{ end }}
</table>
<br />
Have fun with that info. You're welcome.<br />
<br />
Yours, the website-content-watcher.<br />
A <i>Scalify</i> Service.
</body>
</html>
`

// MailClient is an SMTP client for sending mails
type MailClient interface {
	Send(msg *gomail.Message) error
}

// Mail is a notifier sending mails
type Mail struct {
	sender string
	client MailClient
}

// NewMail returns a new Mail instance
func NewMail(sender string, client MailClient) *Mail {
	return &Mail{
		sender: sender,
		client: client,
	}
}

// Key returns the identifier of the notifier
func (m *Mail) Key() string {
	return "mail"
}

// Notify sends an email to given target
func (m *Mail) Notify(jobName, target string, diff []api.Diff) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.sender)
	msg.SetHeader("To", target)
	msg.SetHeader("Subject", fmt.Sprintf("Update on watched job %s", jobName))

	body, err := m.renderTemplate(jobName, diff)
	if err != nil {
		return fmt.Errorf("failed to render body: %v", err)
	}

	msg.SetBody("text/html", body)

	if err := m.client.Send(msg); err != nil {
		return fmt.Errorf("failed to send mail: %v", err)
	}

	return nil
}

func (m *Mail) renderTemplate(jobName string, diff []api.Diff) (string, error) {
	t := template.New("mail")
	if _, err := t.Parse(mailTemplate); err != nil {
		return "", fmt.Errorf("failed to parse email template: %v", err)
	}

	buf := &bytes.Buffer{}
	data := struct {
		JobName string
		Diff    []api.Diff
	}{
		JobName: jobName,
		Diff:    diff,
	}

	err := t.Execute(buf, data)

	return buf.String(), err
}
