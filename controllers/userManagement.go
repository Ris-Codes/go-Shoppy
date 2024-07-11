package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
)

//<<<<<<<<<<-View all first ten users->>>>>>>>>>>>>>>>>>>>>

func ViewAllUser(c *gin.Context) {

	var userData []models.User
	db := config.DB

	rows, err := db.Query("SELECT id, firstname, lastname, email, is_blocked, phone FROM users")
	if err != nil {
		c.JSON(500, gin.H{
			"Message": "Database query error",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Isblocked, &user.PhoneNumber); err != nil {
			c.JSON(500, gin.H{
				"Message": "Error scanning user data",
			})
			return
		}
		userData = append(userData, user)
	}

	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{
			"Message": "Error iterating over user data",
		})
		return
	}

	c.HTML(http.StatusOK, "usermanagement.html", gin.H{
		"UserData": userData,
	})
}

//<<<<<<<<<<<<<<-Admin search users->>>>>>>>>>>>>>>>>>>>>>>>>

func AdminSearchUser(c *gin.Context) {

	useridStr := c.Query("userid")

	userid, err := strconv.Atoi(useridStr)
	if err != nil {
		c.JSON(404, gin.H{
			"Error": err.Error(),
		})
		return
	}

	var user models.User
	db := config.DB

	err = db.QueryRow("SELECT id, firstname, lastname, email, phone FROM users WHERE id = $1", userid).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{
				"Error": "User not found",
			})
		} else {
			c.JSON(500, gin.H{
				"Error": err.Error(),
			})
		}
		return
	}

	c.HTML(http.StatusOK, "usermanagement.html", gin.H{
		"SearchResult": user,
	})
}

//<<<<<<<<<<-Admin block user->>>>>>>>>>>>>>>>>>>>>>>>>>>>

func AdminBlockUser(c *gin.Context) {
	userid, err := strconv.Atoi(c.PostForm("userid"))
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error occure while converting string",
		})
		return
	}

	var user models.User
	db := config.DB

	// Check if the user exists
	err = db.QueryRow("SELECT is_blocked FROM users WHERE id = $1", userid).Scan(&user.Isblocked)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{
				"Message": "User not exist",
			})
		} else {
			c.JSON(500, gin.H{
				"Error": err.Error(),
			})
		}
		return
	}

	// Update the isblocked status
	newStatus := !user.Isblocked
	_, err = db.Exec("UPDATE users SET is_blocked = $1 WHERE id = $2", newStatus, userid)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": err.Error(),
		})
		return
	}

	// message := "User blocked"
	// if !newStatus {
	// 	message = "User unblocked"
	// }

	c.Redirect(http.StatusSeeOther, "/admin/user/viewuser")
}

// >>>>>>>>>> Get user profile <<<<<<<<<<<<<<<<<<<<<<<<<
func GetUserProfile(c *gin.Context) {
	id := c.Query("userId")
	var userData models.User
	db := config.DB

	err := db.QueryRow("SELECT firstname, lastname, email, phone, is_blocked FROM users WHERE id = $1", id).Scan(&userData.FirstName, &userData.LastName, &userData.Email, &userData.PhoneNumber, &userData.Isblocked)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{
				"Message": "User does not exist",
			})
		} else {
			c.JSON(500, gin.H{
				"Error": err.Error(),
			})
		}
		return
	}

	c.HTML(http.StatusOK, "userprofile.html", gin.H{
		"ID":          id,
		"FirstName":   userData.FirstName,
		"LastName":    userData.LastName,
		"Email":       userData.Email,
		"PhoneNumber": userData.PhoneNumber,
		"IsBlocked":   userData.Isblocked,
	})
}

// >>>>>>>>>>>>>>> Edit user profile <<<<<<<<<<<<<<<<<<
type ProfileData struct {
	Firstname   string `form:"firstname" binding:"required"`
	Lastname    string `form:"lastname" binding:"required"`
	PhoneNumber string `form:"phone" binding:"required"`
}

func EditUserProfileByadmin(c *gin.Context) {
	uid := c.Param("id")
	id, err := strconv.Atoi(uid)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
	}
	var userEnterdata ProfileData

	if c.Bind(&userEnterdata) != nil {
		c.JSON(400, gin.H{
			"Error": "Unable to Bind JSON data",
		})
		return
	}

	db := config.DB
	_, err = db.Exec("UPDATE users SET firstname = $1, lastname = $2, phone = $3 WHERE id = $4",
		userEnterdata.Firstname, userEnterdata.Lastname, userEnterdata.PhoneNumber, id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": err.Error(),
		})
		return
	}

	// Redirect back to the user profile page with updated details
	c.Redirect(http.StatusFound, "/admin/user/getuserprofile?userId="+uid)
}
