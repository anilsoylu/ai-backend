package email

import (
	"fmt"
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

// SendPasswordResetEmail sends a password reset email to the user
func SendPasswordResetEmail(to string, resetToken string) error {
	// Initialize Resend client
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not set")
	}
	
	log.Printf("Sending reset email to: %s", to)
	client := resend.NewClient(apiKey)

	// Create email params
	params := &resend.SendEmailRequest{
		From:    "Answer App <onboarding@resend.dev>",
		To:      []string{to},
		Subject: "Password Reset Request",
		Html: fmt.Sprintf(`
			<h1>Password Reset Request</h1>
			<p>You have requested to reset your password. Please use the following token to reset your password:</p>
			<p><strong>%s</strong></p>
			<p>This token will expire in 1 hour.</p>
			<p>If you did not request this password reset, please ignore this email.</p>
			<br>
			<p>Best regards,</p>
			<p>Answer App Team</p>
		`, resetToken),
	}

	log.Printf("Attempting to send email with params: %+v", params)

	// Send email
	resp, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully. Response ID: %s", resp.Id)
	return nil
} 