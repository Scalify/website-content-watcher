package mail

import "gopkg.in/gomail.v2"

// Client handles sending of mails
type Client struct {
	dialer *gomail.Dialer
}

// New returns a new Client instance
func New(dialer *gomail.Dialer) *Client {
	return &Client{
		dialer: dialer,
	}
}

// Send sends a given mail using SMTP
func (c *Client) Send(msg *gomail.Message) error {
	return c.dialer.DialAndSend(msg)
}
