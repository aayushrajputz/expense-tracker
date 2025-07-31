package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/your-github/expense-tracker-backend/config"
	"github.com/your-github/expense-tracker-backend/models"
	"github.com/your-github/expense-tracker-backend/utils"
)

type AuthService struct {
	DB       *gorm.DB
	EmailSvc *EmailService
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		DB:       db,
		EmailSvc: NewEmailService(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Pass),
	}
}

func (s *AuthService) Register(user *models.User) error {
	// Check if user already exists
	var existingUser models.User
	if err := s.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return errors.New("user with this email already exists")
	}

	hash, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hash
	user.Budget = GetDefaultBudget(user.Budget)

	// Create user
	if err := s.DB.Create(user).Error; err != nil {
		return err
	}

	// Generate and send OTP
	otp := s.EmailSvc.GenerateOTP()
	otpModel := models.OTP{
		Email:     user.Email,
		Code:      otp,
		ExpiresAt: time.Now().Add(10 * time.Minute), // OTP expires in 10 minutes
		Used:      false,
	}

	// Save OTP to database
	if err := s.DB.Create(&otpModel).Error; err != nil {
		return err
	}

	// Send OTP email
	if err := s.EmailSvc.SendOTP(user.Email, otp); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Login(email, pw string) (models.User, error) {
	var u models.User
	if err := s.DB.Where("email = ?", email).First(&u).Error; err != nil {
		return u, errors.New("invalid credentials")
	}
	if !utils.CheckPassword(u.Password, pw) {
		return u, errors.New("invalid credentials")
	}
	return u, nil
}

func (s *AuthService) VerifyOTP(email, otpCode string) error {
	// Find the most recent OTP for this email
	var otp models.OTP
	if err := s.DB.Where("email = ? AND used = ?", email, false).
		Order("created_at DESC").
		First(&otp).Error; err != nil {
		return errors.New("invalid or expired OTP")
	}

	// Check if OTP is expired
	if time.Now().After(otp.ExpiresAt) {
		return errors.New("OTP has expired")
	}

	// Check if OTP code matches
	if otp.Code != otpCode {
		return errors.New("invalid OTP code")
	}

	// Mark OTP as used
	otp.Used = true
	if err := s.DB.Save(&otp).Error; err != nil {
		return err
	}

	// Mark user as verified (you might want to add a verified field to User model)
	var user models.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ResendOTP(email string) error {
	// Check if user exists
	var user models.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Generate new OTP
	otp := s.EmailSvc.GenerateOTP()
	otpModel := models.OTP{
		Email:     email,
		Code:      otp,
		ExpiresAt: time.Now().Add(10 * time.Minute), // OTP expires in 10 minutes
		Used:      false,
	}

	// Save new OTP to database
	if err := s.DB.Create(&otpModel).Error; err != nil {
		return err
	}

	// Send OTP email
	if err := s.EmailSvc.SendOTP(email, otp); err != nil {
		return err
	}

	return nil
}

func GetDefaultBudget(b float64) float64 {
	if b <= 0 {
		return 1000.0 // Default budget of $1000
	}
	return b
}
