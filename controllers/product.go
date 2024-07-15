package controllers

import (
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

// >>>>>>>>>>>>>> View products <<<<<<<<<<<<<<<<<<<<<<<<<<
func ViewProducts(c *gin.Context) {
	db := config.DB
	var products []models.Product

	rows, err := db.Query("SELECT p.id, p.product_name, p.description, p.price, p.stock, b.brand_name, c.category_name, i.image FROM product p JOIN brand b ON p.brand_id = b.id JOIN category c ON p.category_id = c.id LEFT JOIN image i ON p.id = i.product_id")
	if err != nil {
		c.JSON(500, gin.H{
			"Message": "Database query error",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.ProductName, &product.Description, &product.Price, &product.Stock, &product.Brand.BrandName, &product.Category.CategoryName, &product.Image.Image); err != nil {
			c.JSON(500, gin.H{
				"Message": "Error scanning product data",
			})
			return
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{
			"Message": "Error iterating over product data",
		})
		return
	}

	// Get brands and categories for the form
	var brands []models.Brand
	var categories []models.Category

	// Query brands
	rows, err = db.Query("SELECT id, brand_name FROM brand")
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var brand models.Brand
		if err := rows.Scan(&brand.ID, &brand.BrandName); err != nil {
			c.JSON(500, gin.H{
				"Error": "Error scanning brand data",
			})
			return
		}
		brands = append(brands, brand)
	}

	// Query categories
	rows, err = db.Query("SELECT id, category_name FROM category")
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.CategoryName); err != nil {
			c.JSON(500, gin.H{
				"Error": "Error scanning category data",
			})
			return
		}
		categories = append(categories, category)
	}

	// Render the template with the products, brands, and categories
	c.HTML(http.StatusOK, "products.html", gin.H{
		"Products":   products,
		"Brands":     brands,
		"Categories": categories,
	})
}

// >>>>>>>>>>>>>> Add products <<<<<<<<<<<<<<<<<<<<<<<<<<
func AddProduct(c *gin.Context) {
	var product models.Product
	err := c.Bind(&product)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Data binding error",
		})
		return
	}

	// Bind brand and category
	brandID, err := strconv.Atoi(c.PostForm("brand_id"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid brand ID",
		})
		return
	}

	categoryID, err := strconv.Atoi(c.PostForm("category_id"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid category ID",
		})
		return
	}

	// Handle image upload
	imagePath, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Image upload error",
		})
		return
	}

	// Open the uploaded file
	file, err := imagePath.Open()
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to open image",
		})
		return
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to decode image",
		})
		return
	}

	// Resize the image to a maximum width of 800 pixels
	resizedImg := resize.Resize(800, 0, img, resize.Lanczos3)

	extension := filepath.Ext(imagePath.Filename)
	image := uuid.New().String() + extension
	savePath := "./public/images/" + image
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to save image",
		})
		return
	}
	defer out.Close()

	// Encode the resized image to the output file
	err = jpeg.Encode(out, resizedImg, nil)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to encode image",
		})
		return
	}

	db := config.DB

	// Insert the new product
	_, err = db.Exec("INSERT INTO product (product_name, description, price, stock, brand_id, category_id) VALUES ($1, $2, $3, $4, $5, $6)",
		product.ProductName, product.Description, product.Price, product.Stock, brandID, categoryID)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error adding product",
		})
		return
	}

	// Get the last inserted product ID
	var productID int
	err = db.QueryRow("SELECT lastval()").Scan(&productID)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error retrieving product ID",
		})
		return
	}

	// Insert the image data into the database
	_, err = db.Exec("INSERT INTO image (image, product_id) VALUES ($1, $2)", image, productID)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error adding image",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/viewproducts")
}

func SearchProductByAdmin(c *gin.Context) {
	productName := c.Query("name")
	db := config.DB
	var products []models.Product

	rows, err := db.Query("SELECT p.id, p.product_name, p.description, p.price, p.stock, b.brand_name, c.category_name, i.image FROM product p JOIN brand b ON p.brand_id = b.id JOIN category c ON p.category_id = c.id LEFT JOIN image i ON p.id = i.product_id WHERE product_name ILIKE $1", "%"+productName+"%")
	if err != nil {
		c.JSON(500, gin.H{
			"Message": "Database query error",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.ProductName, &product.Description, &product.Price, &product.Stock, &product.Brand.BrandName, &product.Category.CategoryName, &product.Image.Image); err != nil {
			c.JSON(500, gin.H{
				"Message": "Error scanning product data",
			})
			return
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{
			"Message": "Error iterating over product data",
		})
		return
	}

	c.HTML(http.StatusOK, "products.html", gin.H{
		"Products": products,
	})
}

func EditProductPage(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	db := config.DB

	err := db.QueryRow("SELECT id, product_name, description, price, stock, brand_id, category_id FROM product WHERE id = $1", id).Scan(&product.ID, &product.ProductName, &product.Description, &product.Price, &product.Stock, &product.BrandId, &product.CategoryId)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}

	// Get all brands
	brandRows, err := db.Query("SELECT id, brand_name FROM brand")
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}
	defer brandRows.Close()

	var brands []models.Brand
	for brandRows.Next() {
		var b models.Brand
		if err := brandRows.Scan(&b.ID, &b.BrandName); err != nil {
			c.JSON(500, gin.H{
				"Error": "Error scanning brand data",
			})
			return
		}
		brands = append(brands, b)
	}

	// Get all categories
	rows, err := db.Query("SELECT id, category_name FROM category")
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.CategoryName); err != nil {
			c.JSON(500, gin.H{
				"Error": "Error scanning category data",
			})
			return
		}
		categories = append(categories, category)
	}

	c.HTML(http.StatusOK, "editproduct.html", gin.H{
		"Product":    product,
		"Brands":     brands,
		"Categories": categories,
	})
}

func EditProduct(c *gin.Context) {
	pid := c.Param("id")
	id, err := strconv.Atoi(pid)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid product ID",
		})
		return
	}

	var product models.Product
	err = c.Bind(&product)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Data binding error",
		})
		return
	}

	db := config.DB

	// Update the product
	_, err = db.Exec("UPDATE product SET product_name = $1, description = $2, price = $3, stock = $4, brand_id = $5, category_id = $6 WHERE id = $7",
		product.ProductName, product.Description, product.Price, product.Stock, product.BrandId, product.CategoryId, id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error updating product",
		})
		return
	}

	// Check if a new image is uploaded
	imagePath, err := c.FormFile("image")
	if err == nil {
		extension := filepath.Ext(imagePath.Filename)
		image := uuid.New().String() + extension
		err = c.SaveUploadedFile(imagePath, "./public/images/"+image)
		if err != nil {
			c.JSON(500, gin.H{
				"Error": "Error saving image",
			})
			return
		}

		// Update the image in the database
		_, err = db.Exec("UPDATE image SET image = $1 WHERE product_id = $2", image, id)
		if err != nil {
			c.JSON(500, gin.H{
				"Error": "Error updating image",
			})
			return
		}
	}

	c.Redirect(http.StatusSeeOther, "/admin/viewproducts")
}