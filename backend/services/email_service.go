package services

import (
	"fmt"
	"math/rand"
	"time"

	"gopkg.in/mail.v2"
)

type EmailService struct {
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
}

func NewEmailService(smtpHost string, smtpPort int, smtpUser, smtpPass string) *EmailService {
	return &EmailService{
		SMTPHost: smtpHost,
		SMTPPort: smtpPort,
		SMTPUser: smtpUser,
		SMTPPass: smtpPass,
	}
}

// GenerateOTP generates a 6-digit OTP
func (s *EmailService) GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000 // Generates number between 100000 and 999999
	return fmt.Sprintf("%06d", otp)
}

// SendOTP sends OTP to the specified email
func (s *EmailService) SendOTP(email, otp string) error {
	// Create new message
	m := mail.NewMessage()
	m.SetHeader("From", s.SMTPUser)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "BucksInfo - Email Verification Code")

	// Email body
	body := fmt.Sprintf(`
		<html>
		<body>
			<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
				<div style="background: linear-gradient(135deg, #FFD700, #FFA500); padding: 20px; border-radius: 10px; text-align: center;">
					<h1 style="color: #333; margin: 0; font-size: 28px;">BucksInfo</h1>
					<p style="color: #333; margin: 10px 0 0 0; font-size: 16px;">Email Verification</p>
				</div>
				
				<div style="background: #f9f9f9; padding: 30px; border-radius: 10px; margin-top: 20px;">
					<h2 style="color: #333; margin: 0 0 20px 0;">Your Verification Code</h2>
					<p style="color: #666; margin: 0 0 20px 0; font-size: 16px;">
						Thank you for signing up with BucksInfo! Please use the following verification code to complete your registration:
					</p>
					
					<div style="background: #333; color: #FFD700; padding: 20px; border-radius: 10px; text-align: center; margin: 20px 0;">
						<h1 style="margin: 0; font-size: 32px; letter-spacing: 5px; font-family: 'Courier New', monospace;">%s</h1>
					</div>
					
					<p style="color: #666; margin: 20px 0 0 0; font-size: 14px;">
						This code will expire in 10 minutes. If you didn't request this code, please ignore this email.
					</p>
					
					<div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd;">
						<p style="color: #999; margin: 0; font-size: 12px;">
							This is an automated message from BucksInfo. Please do not reply to this email.
						</p>
					</div>
				</div>
			</div>
		</body>
		</html>
	`, otp)

	m.SetBody("text/html", body)

	// Create dialer
	d := mail.NewDialer(s.SMTPHost, s.SMTPPort, s.SMTPUser, s.SMTPPass)
	d.SSL = false
	d.TLSConfig = nil

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
