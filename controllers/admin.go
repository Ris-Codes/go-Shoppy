package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Ris-Codes/go-Shoppy/auth"
	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type checkAdminData struct {
	Firstname   string `form:"firstname" binding:"required"`
	Lastname    string `form:"lastname" binding:"required"`
	Email       string `form:"email" binding:"required,email"`
	Password    string `form:"password" binding:"required"`
	ConfirmPassword string `form:"confirm_password" binding:"required"`
	PhoneNumber int    `form:"phone" binding:"required"`
}

func ValidateAdmin(c *gin.Context) {
	c.Get("admin")

	c.Redirect(http.StatusSeeOther, "/admin/adminpanel")
}

//----------------Admin signup-------------------

func AdminSignup(c *gin.Context) {
	var Data checkAdminData
	if err := c.ShouldBind(&Data); err != nil {
		c.JSON(400, gin.H{
			"error": "Data binding error",
		})
		return
	}

	if Data.Password != Data.ConfirmPassword {
        c.JSON(400, gin.H{
            "error": "Passwords do not match",
        })
        return
    }

	hash, err := bcrypt.GenerateFromPassword([]byte(Data.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Hashing password error",
		})
		return
	}

	db := config.DB

	var tempAdminID int
	err = db.QueryRow("SELECT id FROM admin WHERE email = $1", Data.Email).Scan(&tempAdminID)
	if err == nil {
		c.JSON(409, gin.H{
			"error": "User already exists",
		})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(500, gin.H{
			"error": "Database query error",
		})
		return
	}

	_, err = db.Exec("INSERT INTO admin (firstname, lastname, email, password, phone) VALUES ($1, $2, $3, $4, $5)",
		Data.Firstname, Data.Lastname, Data.Email, string(hash), Data.PhoneNumber,
	)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Admin data creation error",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/login")
}

//----------------Admin signout-------------------

func AdminSignout(c *gin.Context) {
	c.SetCookie("AdminAuthorization", "", -1, "", "", false, false)
	c.Redirect(http.StatusSeeOther, "/admin/login")
}

//----------------Admin Login-------------------

func AdminLogin(c *gin.Context) {
	type AdminData struct {
		Email    string `form:"email" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	var admin AdminData
	if err := c.ShouldBind(&admin); err != nil {
		c.JSON(400, gin.H{
			"error": "Login data binding error",
		})
		return
	}

	db := config.DB

	var checkAdmin models.Admin
	err := db.QueryRow("SELECT id, email, password FROM admin WHERE email = $1", admin.Email).Scan(&checkAdmin.ID, &checkAdmin.Email, &checkAdmin.Password)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{
			"error": "User not found",
		})
		return
	} else if err != nil {
		c.JSON(500, gin.H{
			"error": "Database query error",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(checkAdmin.Password), []byte(admin.Password))
	if err != nil {
		c.JSON(501, gin.H{
			"error": "Username and password invalid",
		})
		return
	}

	// Generating a JWT-token
	str := strconv.Itoa(int(checkAdmin.ID))
	tokenString := auth.TokenGeneration(str)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("AdminAuthorization", tokenString, 3600*24*30, "", "", false, true)

	c.Redirect(http.StatusFound, "/admin/adminvalidate")
}

//>>>>>>>>>>> Admin profile <<<<<<<<<<<<<<<<<<<<<<<
func AdminProfile(c *gin.Context) {
	adminidStr := c.GetString("adminid")
	adminid, err := strconv.Atoi(adminidStr)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}
	var adminData models.Admin
	db := config.DB

	err = db.QueryRow("SELECT firstname, lastname, email, phone FROM admin WHERE id = $1", adminid).Scan(&adminData.Firstname, &adminData.Lastname, &adminData.Email, &adminData.PhoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{
				"Error":   "Admin does not exist",
				"Message": "Admin does not exist",
				"Status":  "false",
			})
		} else {
			c.JSON(500, gin.H{
				"Error": err.Error(),
			})
		}
		return
	}
	
	c.HTML(http.StatusOK, "admin_profile.html", gin.H{
		"FirstName":   adminData.Firstname,
		"LastName":    adminData.Lastname,
		"Email":       adminData.Email,
		"PhoneNumber": adminData.PhoneNumber,
	})
}