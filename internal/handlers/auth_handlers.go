package handlers

import (
	"log"
	"my-project/internal/auth"
	"my-project/internal/database"
	"my-project/internal/models"
	"my-project/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// Login godoc
// @Summary      Logs in a user
// @Description  Logs in a user with phone number and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        login  body      LoginRequest  true  "Login credentials"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("phone_number = ?", req.PhoneNumber).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateJWT(user.PhoneNumber, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

type RequestCodeRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// RequestSMSCode godoc
// @Summary      Requests an SMS verification code
// @Description  Sends a verification code to the user's phone number
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        phone  body      RequestCodeRequest  true  "Phone number"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /login/sms/request [post]
func RequestSMSCode(c *gin.Context) {
	var req RequestCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("phone_number = ?", req.PhoneNumber).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	code := utils.GenerateRandomCode(6)
	user.VerificationCode = code
	user.VerificationCodeExpiresAt = time.Now().Add(5 * time.Minute)

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save verification code"})
		return
	}

	// In a real application, you would send the code via SMS here.
	// For this example, we'll just log it.
	log.Printf("Verification code for %s: %s\n", user.PhoneNumber, code)

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent"})
}

type VerifyCodeRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

// VerifySMSCode godoc
// @Summary      Verifies an SMS code
// @Description  Verifies the SMS code and returns a JWT
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        verification  body      VerifyCodeRequest  true  "Phone number and code"
// @Success      200           {object}  map[string]string
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      404           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Router       /login/sms/verify [post]
func VerifySMSCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("phone_number = ?", req.PhoneNumber).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.VerificationCode == "" || user.VerificationCode != req.Code || time.Now().After(user.VerificationCodeExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired verification code"})
		return
	}

	// Invalidate the code
	user.VerificationCode = ""
	database.DB.Save(&user)

	token, err := auth.GenerateJWT(user.PhoneNumber, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

type RequestEmailCodeRequest struct {
	Email string `json:"email" binding:"required"`
}

// RequestEmailCode godoc
// @Summary      Requests an email verification code
// @Description  Sends a verification code to the user's email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        email  body      RequestEmailCodeRequest  true  "Email"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /login/email/request [post]
func RequestEmailCode(c *gin.Context) {
	var req RequestEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	code := utils.GenerateRandomCode(6)
	user.EmailVerificationCode = code
	user.EmailVerificationCodeExpiresAt = time.Now().Add(5 * time.Minute)

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save verification code"})
		return
	}

	log.Printf("Verification code for %s: %s\n", user.Email, code)

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent to email"})
}

type VerifyEmailCodeRequest struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// VerifyEmailCode godoc
// @Summary      Verifies an email code
// @Description  Verifies the email code and returns a JWT
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        verification  body      VerifyEmailCodeRequest  true  "Email and code"
// @Success      200           {object}  map[string]string
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      404           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Router       /login/email/verify [post]
func VerifyEmailCode(c *gin.Context) {
	var req VerifyEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.EmailVerificationCode == "" || user.EmailVerificationCode != req.Code || time.Now().After(user.EmailVerificationCodeExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired verification code"})
		return
	}

	user.EmailVerificationCode = ""
	database.DB.Save(&user)

	token, err := auth.GenerateJWT(user.PhoneNumber, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
