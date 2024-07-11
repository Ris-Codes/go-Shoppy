package controllers

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// view brand
func ViewBrand(c *gin.Context) {
	var brandData []models.Brand
	db := config.DB

	rows, err := db.Query("SELECT id, brand_name FROM brand")
	if err != nil {
		c.JSON(500, gin.H{
			"Message": "Database query error",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var brand models.Brand
		if err := rows.Scan(&brand.ID, &brand.BrandName); err != nil {
			c.JSON(500, gin.H{
				"Message": "Error scanning brand data",
			})
			return
		}
		brandData = append(brandData, brand)
	}

	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{
			"Message": "Error iterating over brand data",
		})
		return
	}

	// Render the Brands page
	c.HTML(http.StatusOK, "brands.html", gin.H{
		"BrandData": brandData,
	})
}

// Admin add the brand
func AddBrands(c *gin.Context) {
	var addbrand models.Brand
	if c.Bind(&addbrand) != nil {
		c.JSON(400, gin.H{
			"Error": "Could not bind JSON data",
		})
		return
	}

	db := config.DB

	// Insert the new brand
	_, err := db.Exec("INSERT INTO brand (brand_name) VALUES ($1)", addbrand.BrandName)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error adding brand",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/brand")
}

// Edit brand by admin
func EditBrand(c *gin.Context) {
	bid := c.Param("id")
	id, err := strconv.Atoi(bid)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	var editbrands models.Brand
	if err := c.Bind(&editbrands); err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in binding the form data",
		})
		return
	}

	db := config.DB

	// Update the brand name
	_, err = db.Exec("UPDATE brand SET brand_name = $1 WHERE id = $2", editbrands.BrandName, id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error updating the brand",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/brand")
}

// Adding the catogery by admin
func AddCategories(c *gin.Context) {
	type Data struct {
		CategoryName string `form:"category_name" binding:"required"`
	}

	var category Data
	if err := c.ShouldBind(&category); err != nil {
		c.JSON(400, gin.H{
			"Error": "could not bind the form data",
		})
		return
	}

	db := config.DB

	// Check if the category already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM category WHERE category_name = $1", category.CategoryName).Scan(&count)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}

	if count == 0 {
		// Insert the new category
		_, err := db.Exec("INSERT INTO category (category_name) VALUES ($1)", category.CategoryName)
		if err != nil {
			c.JSON(500, gin.H{
				"Error": "Error adding category",
			})
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/viewproducts")
	} else {
		c.JSON(400, gin.H{
			"message": "Category already exists",
		})
	}
}

// Add image of the product by admin
func AddImages(c *gin.Context) {
	imagepath, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "No image file provided",
		})
		return
	}

	pid, err := strconv.Atoi(c.PostForm("product_id"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid product ID",
		})
		return
	}

	db := config.DB

	// Check if the product exists
	var productID int
	err = db.QueryRow("SELECT id FROM product WHERE id = $1", pid).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{
				"Error": "Product not found",
			})
		} else {
			c.JSON(500, gin.H{
				"Error": "Database query error",
			})
		}
		return
	}

	// Save the image file
	extension := filepath.Ext(imagepath.Filename)
	image := uuid.New().String() + extension
	savePath := "./public/images/" + image
	err = c.SaveUploadedFile(imagepath, savePath)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to save image",
		})
		return
	}

	// Insert the image data into the database
	_, err = db.Exec("INSERT INTO image (image, product_id) VALUES ($1, $2)", image, pid)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error adding image",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/viewproducts")
}
