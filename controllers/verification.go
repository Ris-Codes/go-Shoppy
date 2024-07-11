package controllers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	gomail "gopkg.in/gomail.v2"
)

func VerifyOTP(email string) string {
	Otp, err := getRandNum()
	if err != nil {
		panic(err)
	}

	sendMail(email, Otp)
	return Otp
}

func sendMail(email string, otp string) {

	fmt.Println("Email : ", email, " otp :", otp)
	// Sender data.
	from := os.Getenv("EMAIL")
	key := os.Getenv("PASSWORD_KEY")

	// SMTP server configuration.
	smtpHost := "smtp-relay.brevo.com"
	smtpPort := 587

	// Create a new message
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your OTP Code")
	m.SetBody("text/plain", otp+" is your OTP")

	// Dial and send the email
	d := gomail.NewDialer(smtpHost, smtpPort, from, key)

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Could not send email:", err)
		return
	}

	fmt.Println("Email sent successfully!")
}

func getRandNum() (string, error) {
	nBig, e := rand.Int(rand.Reader, big.NewInt(8999))
	if e != nil {
		return "", e
	}
	return strconv.FormatInt(nBig.Int64()+1000, 10), nil
}

// -------Otp validtioin------------->
func OtpValidation(c *gin.Context) {
	type UserOtp struct {
		Otp   string `form:"otp" binding:"required"`
		Email string `form:"email" binding:"required"`
	}
	var userOtp UserOtp
	if err := c.ShouldBind(&userOtp); err != nil {
		c.JSON(400, gin.H{
			"Error": "Could not bind the JSON Data",
		})
		return
	}

	db := config.DB

	var userData models.User
	err := db.QueryRow("SELECT id, firstname, lastname, phone, email, otp FROM users WHERE otp = $1 AND email = $2", userOtp.Otp, userOtp.Email).Scan(&userData.ID, &userData.FirstName, &userData.LastName, &userData.PhoneNumber, &userData.Email, &userData.Otp)

	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{
			"Error": "User not found or OTP incorrect",
		})
		c.Redirect(422, "/user/signup/otpvalidate")
		return
	} else if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/user/login")
}

// Reseting the password
func ChangePassword(c *gin.Context) {
	type userEnterData struct {
		Email           string `form:"email" binding:"required"`
		Otp             string `form:"otp" binding:"required"`
		Password        string `form:"password" binding:"required"`
		ConfirmPassword string `form:"confirm_password" binding:"required"`
	}
	var data userEnterData
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(400, gin.H{
			"Error": "Error when data binding",
		})
		return
	}
	if data.Password != data.ConfirmPassword {
		c.JSON(400, gin.H{
			"Error": "Passwords do not match",
		})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status": "False",
			"Error":  "Hashing password error",
		})
		return
	}

	db := config.DB

	var userData models.User
	err = db.QueryRow("SELECT email, otp FROM users WHERE email = $1", data.Email).Scan(&userData.Email, &userData.Otp)
	if err == sql.ErrNoRows {
		c.JSON(409, gin.H{
			"Error": "User does not exist",
		})
		return
	} else if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}

	if data.Otp != userData.Otp {
		c.JSON(400, gin.H{
			"Error": "Invalid OTP",
		})
		return
	}

	_, err = db.Exec("UPDATE users SET password = $1 WHERE email = $2", hash, data.Email)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to update password",
		})
		return
	}

	c.JSON(200, gin.H{
		"Message": "Password changed successfully",
	})
}
