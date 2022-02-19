package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func SendEmail(username string, emailAddr string, code string) {
	// The sender's data.
	from := "jakobstanleywarth@gmail.com"
	password := os.Getenv("MAILPSW")

	// Receiver email address.
	to := []string{emailAddr}

	// Smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Email Headers
	hfrom := "From: SmokeStop <smokestopper@jakobthesheep.com>\n"
	hto := fmt.Sprintf("To: %s <%s>\n", username, emailAddr)
	subject := "Subject: Confirm your email address\n\n"
	body := fmt.Sprintf("Dear %s,\n\nuse this code to confirm your email address: %s\n\n", username, code)

	// Message.
	message := []byte(fmt.Sprintf("%s%s%s%s", hfrom, hto, subject, body))

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		log.Fatalf("Unable to send email: %v", err)
		return
	}
	log.Printf("Confirmation email sent to %s!\n", emailAddr)
}
