package main

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
	password := os.Getenv("SMTP_PASSWORD")
	// Authentication.
	auth := smtp.PlainAuth("", "apikey", password, "smtp.sendgrid.net")
	err := smtp.SendMail("smtp.sendgrid.net:587", auth, from, to, []byte(message))
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	fmt.Println("Email Sent Successfully! ", email)
	fmt.Println("  > ", message)
	return true, nil
}
