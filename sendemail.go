package sendemail

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jordan-wright/email"
)

func New(c Config) *App {
	return &App{
		config:     c,
		domainList: domainList(c),
	}
}

type App struct {
	config     Config
	domainList map[string]string
}

type Message struct {
	ToEmail         string `form:"toEmail"`
	FromEmailPrefix string `form:"fromEmailPrefix"`
	FromDomainId    string `form:"fromDomainId"`
	Subject         string `form:"subject"`
	Body            string `form:"body"`
}

var emailRegex = regexp.MustCompile(`[^\s\,\;]+@[^\s\,\;]+`)
var fromRegex = regexp.MustCompile(`[^\s\,\;^@]+`)

func splitEmails(input string) []string {
	return emailRegex.FindAllString(input, -1)
}

func domainList(c Config) map[string]string {
	m := make(map[string]string)
	for k, v := range c.Sender {
		m[k] = v.FromDomain
	}
	return m
}

func (a *App) render(c *gin.Context, m Message) {
	c.HTML(http.StatusOK, "main.tmpl", gin.H{
		"title":   a.config.Server.Title,
		"message": m,
		"domains": a.domainList,
	})
}

func (a *App) renderNotification(c *gin.Context, m Message, n string, success bool) {
	status := "danger"
	if success {
		status = "success"
	}
	log.Printf("Message: %+v", m)
	c.HTML(http.StatusOK, "main.tmpl", gin.H{
		"title":              a.config.Server.Title,
		"message":            m,
		"notification":       n,
		"notificationStatus": status,
		"domains":            a.domainList,
	})
}

func (a *App) sendEmail(ctx context.Context, msg Message) error {
	domainConfig, ok := a.config.Sender[msg.FromDomainId]
	if !ok {
		return fmt.Errorf("Could not find domain ID '%s' in config with %d domains", msg.FromDomainId, len(a.config.Sender))
	}

	e := email.NewEmail()
	e.From = fmt.Sprintf("%s@%s", msg.FromEmailPrefix, domainConfig.FromDomain)
	e.To = splitEmails(msg.ToEmail)
	e.Subject = msg.Subject
	e.Bcc = domainConfig.Bcc
	e.Text = []byte(msg.Body)

	smtpAddr := fmt.Sprintf("%s:%d", domainConfig.Host, domainConfig.Port)
	smtpAuth := smtp.PlainAuth("", domainConfig.Username, domainConfig.Password, domainConfig.Host)
	log.Printf("Sending email: %+v\n %+v\n %+v\n", e, smtpAddr, smtpAuth)
	if a.config.Server.DryRun {
		log.Println("Not actually sending an email because DryRun is enabled.")
		return nil
	}
	return e.Send(smtpAddr, smtpAuth)
}

func (a *App) Main(c *gin.Context) {
	if c.Request.Method != "POST" {
		var message Message
		a.render(c, message)
		return
	}
	var message Message
	err := c.ShouldBind(&message)
	if err != nil {
		log.Println(err)
		return
	}
	var errors []string
	if message.ToEmail == "" {
		errors = append(errors, "To Email required.")
	}
	if message.FromEmailPrefix == "" {
		errors = append(errors, "From Email required.")
	}
	if message.Subject == "" {
		errors = append(errors, "Subject required.")
	}
	if message.Body == "" {
		errors = append(errors, "Email message required.")
	}
	if len(errors) > 0 {
		errorCat := strings.Join(errors, " ")
		a.renderNotification(c, message, errorCat, false)
		return
	}
	err = a.sendEmail(c, message)
	if err != nil {
		a.renderNotification(c, message, fmt.Sprintf("%s", err), false)
		return
	}
	a.renderNotification(c, message, "Email sent!", true)
}
