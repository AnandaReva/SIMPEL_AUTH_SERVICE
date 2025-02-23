package mail

import (
	"auth_service/logger"
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail mengirimkan email dengan SMTP
func SendEmail(emailDestination, subject, message string) error {
	// Ambil credential dari environment variable

	SMTPSERVER := os.Getenv("SMTPSERVER")
	SMTPPORT := os.Getenv("SMTPPORT")
	SMTPUSER := os.Getenv("SMTPUSER")
	SMTPPASS := os.Getenv("SMTPPASS")
	SMTPFROM := os.Getenv("SMTPFROM")

	if SMTPUSER == "" || SMTPPASS == "" || SMTPSERVER == "" || SMTPPORT == "" || SMTPFROM == "" {
		logger.Error("SendMail", "SMTP credentials are missing")
		return fmt.Errorf("SMTP credentials are missing")
	}

	// logger.Info("SendMail", "-----------SMTP CONF : ")
	// logger.Info("SendMail", "SMTPUSER: ", SMTPUSER)
	// logger.Info("SendMail", "SMTPPASS: ", SMTPPASS)
	// logger.Info("SendMail", "SMTPSERVER: ", SMTPSERVER)
	// logger.Info("SendMail", "SMTPPORT: ", SMTPPORT)

	logger.Info("SendMail", "EMAIL TO: ", emailDestination)
	logger.Info("SendMail", "SUBJECT: ", subject)
	logger.Info("SendMail", "MESSAGE: ", message)

	// Konfigurasi autentikasi SMTP
	auth := smtp.PlainAuth("", SMTPUSER, SMTPPASS, SMTPSERVER)

	// Format pesan email
	msg := []byte("To: " + emailDestination + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		message + "\r\n")

	// Kirim email
	err := smtp.SendMail(SMTPSERVER+":"+SMTPPORT, auth, SMTPFROM, []string{emailDestination}, msg)
	if err != nil {
		logger.Error("SendMail", "Failed to send email:", err)
		return err
	}

	logger.Info("SendMail", "Email successfully sent to:", emailDestination)
	return nil
}
