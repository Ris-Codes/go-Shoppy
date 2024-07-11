package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
)

// >>>>>>>>>>>>Veiw user profile <<<<<<<<<<<<<<<<<<<<<<<<<<
func ShowUserDetails(c *gin.Context) {
	var userData models.User
	id, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	db := config.DB

	err = db.QueryRow("SELECT id, firstname, lastname, email, phone FROM users WHERE id = $1", id).Scan(&userData.ID, &userData.FirstName, &userData.LastName, &userData.Email, &userData.PhoneNumber)
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

	c.HTML(202, "profile.html", gin.H{
		"FirstName":   userData.FirstName,
		"LastName":    userData.LastName,
		"Email":       userData.Email,
		"PhoneNumber": userData.PhoneNumber,
	})
}

// >>>>>>>>>>> Edit user profile <<<<<<<<<<<<<<<<<<<<<<<<<<<<<
func EditUserProfilePage(c *gin.Context) {
	var userData models.User
	id, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	db := config.DB

	err = db.QueryRow("SELECT id, firstname, lastname, phone FROM users WHERE id = $1", id).Scan(&userData.ID, &userData.FirstName, &userData.LastName, &userData.PhoneNumber)
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

	c.HTML(202, "edituserprofile.html", gin.H{
		"FirstName":   userData.FirstName,
		"LastName":    userData.LastName,
		"PhoneNumber": userData.PhoneNumber,
	})
}

func EditUserProfilebyUser(c *gin.Context) {
	type editData struct {
		Firstname   string `form:"firstname" binding:"required"`
		Lastname    string `form:"lastname" binding:"required"`
		PhoneNumber int    `form:"phone" binding:"required"`
	}
	var userEnterData editData
	if err := c.ShouldBind(&userEnterData); err != nil {
		c.JSON(400, gin.H{
			"error": "Data binding error",
		})
		return
	}

	id, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	db := config.DB

	var userData models.User
	err = db.QueryRow("SELECT id, firstname, lastname, phone FROM users WHERE id = $1", id).Scan(&userData.ID, &userData.FirstName, &userData.LastName, &userData.PhoneNumber)
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

	_, err = db.Exec("UPDATE users SET firstname = $1, lastname = $2, phone = $3 WHERE id = $4", userEnterData.Firstname, userEnterData.Lastname, userEnterData.PhoneNumber, id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to update user profile",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/user/viewprofile")
}
