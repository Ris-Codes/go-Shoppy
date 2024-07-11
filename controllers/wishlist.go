package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
)

func AddToWishlist(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	type data struct {
		ProductID uint `form:"product_id" json:"product_id"`
	}
	var bindData data
	var productData models.Product

	// Binding the data from the input
	if c.ShouldBind(&bindData) != nil {
		c.JSON(400, gin.H{
			"Bad Request": "Could not bind the data",
		})
		return
	}

	db := config.DB

	// Check if the product exists

	err = db.QueryRow("SELECT id FROM product WHERE id = $1", bindData.ProductID).Scan(&productData.ID)
	if err == sql.ErrNoRows {
		c.JSON(400, gin.H{
			"Message": "Product not exist",
		})
		return
	} else if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}

	// Checking if the product is already in the cart
	err = db.QueryRow("SELECT user_id, product_id FROM wishlist WHERE product_id = $1 AND user_id = $2", bindData.ProductID, userID).Scan(&userID, &bindData.ProductID)
	if err == sql.ErrNoRows {
		// Insert product into wishlist
		_, err = db.Exec("INSERT INTO wishlist (user_id, product_id) VALUES ($1, $2)", userID, bindData.ProductID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error adding product to wishlist",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product added to wishlist successfully",
	})
}

func RemoveFromWishlist(c *gin.Context) {
	type data struct {
		ProductID uint `json:"product_id" form:"product_id"`
	}
	var bindData data

	if err := c.ShouldBind(&bindData); err != nil {
		c.JSON(400, gin.H{
			"Error": "Could not bind the JSON data",
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

	_, err = db.Exec("DELETE FROM wishlist WHERE product_id = $1 AND user_id = $2", bindData.ProductID, id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to remove from wishlist",
		})
		return
	}

	c.JSON(200, gin.H{
		"Message": "Removed from wishlist successfully",
	})
}

func Wishlist(c *gin.Context) {
	userId, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	db := config.DB

	// Execute the query
	rows, err := db.Query("SELECT p.id, p.product_name, p.description, p.stock, p.price, c.category_name, b.brand_name, i.image FROM wishlist w JOIN product p ON w.product_id = p.id JOIN category c ON p.category_id = c.id JOIN brand b ON p.brand_id = b.id LEFT JOIN image i ON p.id = i.product_id WHERE w.user_id = $1;", userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch wishlist items",
		})
		return
	}
	defer rows.Close()

	var wishlistItems []struct {
		Product      models.Product
		Category     models.Category
		Brand        models.Brand
		Image        models.Image
		InCart       bool
	}

	// Iterate through the rows and populate wishlistItems slice
	for rows.Next() {
		var item struct {
			Product      models.Product
			Category     models.Category
			Brand        models.Brand
			Image        models.Image
			InCart       bool
		}
		err := rows.Scan(&item.Product.ID, &item.Product.ProductName, &item.Product.Description, &item.Product.Stock, &item.Product.Price, &item.Category.CategoryName, &item.Brand.BrandName, &item.Image.Image)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error scanning wishlist item" + err.Error(),
			})
			return
		}

		// Check if the product is in the cart
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM cart WHERE product_id = $1 AND user_id = $2", item.Product.ID, userId).Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database query error",
			})
			return
		}
		item.InCart = count > 0

		wishlistItems = append(wishlistItems, item)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error with wishlist item rows",
		})
		return
	}

	// Return the wishlist items as JSON response
	c.HTML(http.StatusOK, "wishlist.html", gin.H{
		"Wishlist": wishlistItems,
	})
}