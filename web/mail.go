package web

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendMail(email, subject, body string) (bool, error) {
	// Receiver email address.
	to := []string{
		email,
	}
	// Message.
	message := "From: no-reply@ds.10z.dev\r\n" +
		"To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n"
	// Sender data.
	from := "no-reply@ds.10z.dev"
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "outgoing.mail.go.th"
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		// regularly 465, but mail.go.th doesn't do regular, huh?
		port = "587"
	}
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	// Authentication.
	smtpAddr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, password, host)
	err := smtp.SendMail(smtpAddr, auth, from, to, []byte(message))
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	fmt.Println("Email Sent Successfully! ", email)
	fmt.Println("  > ", message)
	return true, nil
}
