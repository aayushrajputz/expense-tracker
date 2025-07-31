package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-github/expense-tracker-backend/config"
	"github.com/your-github/expense-tracker-backend/models"
	"github.com/your-github/expense-tracker-backend/services"
	"github.com/your-github/expense-tracker-backend/utils"
)

type AuthController struct{ 
	S *services.AuthService 
	Config *config.Config
}

type registerDTO struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	Budget   float64 `json:"budget"`
}

func (c *AuthController) Register(ctx *gin.Context) {
	var in registerDTO
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u := models.User{Name: in.Name, Email: in.Email, Password: in.Password, Budget: in.Budget}
	if err := c.S.Register(&u); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "User registered successfully. Please check your email for verification code."})
}

type loginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (c *AuthController) Login(ctx *gin.Context) {
	var in loginDTO
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := c.S.Login(in.Email, in.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token, _ := utils.GenerateToken(user.ID, c.Config.JWT.Secret)
	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

type verifyOTPDTO struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

func (c *AuthController) VerifyOTP(ctx *gin.Context) {
	var in verifyOTPDTO
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.S.VerifyOTP(in.Email, in.OTP); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

type resendOTPDTO struct {
	Email string `json:"email" binding:"required,email"`
}

func (c *AuthController) ResendOTP(ctx *gin.Context) {
	var in resendOTPDTO
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.S.ResendOTP(in.Email); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OTP resent successfully"})
}
