package controllers

import (
	"net/http"
	"strconv"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
)

func AddAddress(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid user ID",
		})
		return
	}

	var address models.Address
	if err := c.ShouldBind(&address); err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in binding the form data",
		})
		return
	}

	db := config.DB

	_, err = db.Exec("UPDATE address SET default_add = false WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error updating addresses",
		})
		return
	}

	address.Userid = userID
	_, err = db.Exec("INSERT INTO address (user_id, name, phone, house_no, area, landmark, city, pincode, district, state, country, default_add) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)", address.Userid, address.Name, address.Phoneno, address.Houseno, address.Area, address.Landmark, address.City, address.Pincode, address.District, address.State, address.Country, true)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error adding address",
		})
		return
	}

	c.Redirect(http.StatusFound, "/user/showaddresses")
}

// >>>>>>>>>>>>> show address <<<<<<<<<<<<<<<<<<<<<
func ShowAddress(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid user ID",
		})
		return
	}

	db := config.DB
	var addresses []models.Address

	// Query to get all addresses for the user
	rows, err := db.Query("SELECT id, user_id, name, phone, house_no, area, landmark, city, pincode, district, state, country, default_add FROM address WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var address models.Address
		err := rows.Scan(&address.Addressid, &address.Userid, &address.Name, &address.Phoneno, &address.Houseno, &address.Area, &address.Landmark, &address.City, &address.Pincode, &address.District, &address.State, &address.Country, &address.Defaultadd)
		if err != nil {
			c.JSON(500, gin.H{
				"Error": "Error scanning address",
			})
			return
		}
		addresses = append(addresses, address)
	}

	// Check if there are any addresses
	hasAddresses := len(addresses) > 0

	c.HTML(http.StatusOK, "address.html", gin.H{
		"Addresses":    addresses,
		"HasAddresses": hasAddresses,
	})
}

// >>>>>>>>>>>>>> Edit Address <<<<<<<<<<<<<<<<<<<<<
func EditAddressPage(c *gin.Context) {
	id := c.Param("id")
	addressID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid address ID",
		})
		return
	}

	db := config.DB

	var address models.Address
	err = db.QueryRow("SELECT id, user_id, name, phone, house_no, area, landmark, city, pincode, district, state, country, default_add FROM address WHERE id = $1", addressID).Scan(
		&address.Addressid, &address.Userid, &address.Name, &address.Phoneno, &address.Houseno, &address.Area, &address.Landmark, &address.City, &address.Pincode, &address.District, &address.State, &address.Country, &address.Defaultadd)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error retrieving address",
		})
		return
	}

	c.HTML(http.StatusOK, "edit_address.html", gin.H{
		"Address": address,
	})
}


func EditUserAddress(c *gin.Context) {
	id := c.Param("id")

	var userAddress models.Address
	if err := c.Bind(&userAddress); err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in binding JSON data",
		})
		return
	}
	addressID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Invalid address ID",
		})
		return
	}
	userAddress.Addressid = addressID

	db := config.DB
	_, err = db.Exec("UPDATE address SET name = $1, phone = $2, house_no = $3, area = $4, landmark = $5, city = $6, pincode = $7, district = $8, state = $9, country = $10 WHERE id = $11", userAddress.Name, userAddress.Phoneno, userAddress.Houseno, userAddress.Area, userAddress.Landmark, userAddress.City, userAddress.Pincode, userAddress.District, userAddress.State, userAddress.Country, userAddress.Addressid)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error updating address",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/user/showaddresses")
}

func DeleteUserAddress(c *gin.Context) {
	id := c.Param("id")

	addressID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Invalid address ID",
		})
		return
	}


	db := config.DB
	result, err := db.Exec("DELETE FROM address WHERE id = $1", addressID)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error deleting address",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(500, gin.H{
			"Error": "No address found with the given ID",
		})
		return
	}


	c.Redirect(http.StatusSeeOther, "/user/showaddresses")
}

// >>>>>>>>>>>>>> Set Default <<<<<<<<<<<<<<<<<<<<<
func SetDefaultAddress(c *gin.Context) {
	addressID := c.Param("id")
	userID, _ := strconv.Atoi(c.GetString("userid"))

	db := config.DB

	// Set all addresses for the user to not be the default
	_, err := db.Exec("UPDATE address SET default_add = FALSE WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error updating addresses",
		})
		return
	}

	// Set the selected address as the default
	_, err = db.Exec("UPDATE address SET default_add = TRUE WHERE id = $1 AND user_id = $2", addressID, userID)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error updating address",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/user/showaddresses")
}
