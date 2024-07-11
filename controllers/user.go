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

type checkUserData struct {
	Firstname       string `form:"firstname" binding:"required"`
	Lastname        string `form:"lastname" binding:"required"`
	Email           string `form:"email" binding:"required"`
	Password        string `form:"password" binding:"required"`
	ConfirmPassword string `form:"confirm_password" binding:"required"`
	PhoneNumber     int    `form:"phone" binding:"required"`
	Otp             string
}

//----------User signup--------------------------------------->

func UserSignUP(c *gin.Context) {
	var Data checkUserData

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
			"Status": "False",
			"Error":  "Hashing password error",
		})
		return
	}

	db := config.DB

	var tempUserEmail string
	err = db.QueryRow("SELECT email FROM users WHERE email = $1", Data.Email).Scan(&tempUserEmail)

	if err == sql.ErrNoRows {
		otp := VerifyOTP(Data.Email)
		_, err = db.Exec("INSERT INTO users (firstname, lastname, email, password, phone, otp) VALUES ($1, $2, $3, $4, $5, $6)",
			Data.Firstname, Data.Lastname, Data.Email, string(hash), Data.PhoneNumber, otp)

		if err != nil {
			c.JSON(500, gin.H{
				"Status": "False",
				"Error":  "User data creating error",
			})
			return
		}

		c.Redirect(http.StatusSeeOther, "/user/signup/otpvalidate")
	} else if err != nil {
		c.JSON(500, gin.H{
			"Status": "False",
			"Error":  "Database query error",
		})
	} else {
		c.JSON(409, gin.H{
			"Error": "User already exists",
		})
	}
}

//------------User login------------------------>

func UserLogin(c *gin.Context) {
	type userData struct {
		Email    string `form:"email" binding:"required,email"`
		Password string `form:"password" binding:"required"`
	}

	var user userData
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(400, gin.H{
			"error": "Data binding error",
		})
		return
	}

	db := config.DB

	var checkUser models.User
	err := db.QueryRow("SELECT id, email, password, is_blocked FROM users WHERE email = $1", user.Email).Scan(&checkUser.ID, &checkUser.Email, &checkUser.Password, &checkUser.Isblocked)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{
			"Status":  "false",
			"Message": "User not found",
		})
		return
	} else if err != nil {
		c.JSON(500, gin.H{
			"Status":  "false",
			"Message": err.Error(),
		})
		return
	}

	if checkUser.Isblocked {
		c.JSON(401, gin.H{
			"Status":  "Authorization",
			"Message": "User blocked by admin",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(checkUser.Password), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status": "false",
			"error":  "Password is incorrect",
		})
		return
	}

	//>>>>>>>>>>>>>>>>> Generating a JWT-token <<<<<<<<<<<<<<<//
	str := strconv.Itoa(int(checkUser.ID))
	tokenString := auth.TokenGeneration(str)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("UserAuthorization", tokenString, 3600*24*30, "/", "", false, true)

	c.Redirect(http.StatusSeeOther, "/user/uservalidate")
}

// >>>>>>>>>>>>>>> Validate <<<<<<<<<<<<<<<<<<<<<<<<<<<<
func UserValidate(c *gin.Context) {
	c.Get("user")

	c.Redirect(http.StatusSeeOther, "/")
}

//----------------User signout-------------------

func UserSignout(c *gin.Context) {
	c.SetCookie("UserAuthorization", "", -1, "", "", false, false)
	c.Redirect(http.StatusSeeOther, "/user/login")
}

// >>>>>>>>>>>> Change password by user <<<<<<<<<<<<<<<<
type passData struct {
	OldPassword     string `form:"old_password" binding:"required"`
	Password        string `form:"password" binding:"required"`
	ConfirmPassword string `form:"confirm_password" binding:"required"`
}

func UserChangePassword(c *gin.Context) {
	var passwordData passData
	if err := c.ShouldBind(&passwordData); err != nil {
		c.JSON(400, gin.H{
			"error": "Data binding error",
		})
		return
	}

	if passwordData.Password != passwordData.ConfirmPassword {
		c.JSON(400, gin.H{
			"Message": "Passwords do not match",
		})
		return
	}

	id, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	db := config.DB

	var userData models.User
	err = db.QueryRow("SELECT id, password FROM users WHERE id = $1", id).Scan(&userData.ID, &userData.Password)
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

	err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(passwordData.OldPassword))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "false",
			"error":  "Old password is incorrect",
		})
		return
	}

	Updatepassword(c, id, passwordData.Password)
}

// >>>>>>>>>>>>> updating new user <<<<<<<<<<<<<<<<<<<<<<
func Updatepassword(c *gin.Context, id int, newPassword string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status": "False",
			"Error":  "Hashing password error",
		})
		return
	}

	db := config.DB

	_, err = db.Exec("UPDATE users SET password = $1 WHERE id = $2", hash, id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to update password",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/user/viewprofile")
}

// >>>>>>>>>>>>>>>>> User Home <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

func UserHome(c *gin.Context) {
	// Get page number from query parameter, default to 1 if not provided
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	// Get search value, category, and brand filters from query parameters
	searchValue := c.DefaultQuery("search_value", "")
	categoryIDStr := c.DefaultQuery("category", "")
	brandIDStr := c.DefaultQuery("brand", "")

	// Convert categoryID and brandID to integers
	var categoryID, brandID int
	if categoryIDStr != "" {
		categoryID, err = strconv.Atoi(categoryIDStr)
		if err != nil {
			categoryID = 0
		}
	}
	if brandIDStr != "" {
		brandID, err = strconv.Atoi(brandIDStr)
		if err != nil {
			brandID = 0
		}
	}

	// Define number of products per page
	const productsPerPage = 10
	offset := (page - 1) * productsPerPage

	db := config.DB

	// Fetch total number of products
	var totalProducts int
	countQuery := "SELECT COUNT(*) FROM product p LEFT JOIN brand b ON p.brand_id = b.id WHERE 1=1"
	if categoryID > 0 {
		countQuery += " AND p.category_id = " + strconv.Itoa(categoryID)
	}
	if brandID > 0 {
		countQuery += " AND p.brand_id = " + strconv.Itoa(brandID)
	}
	if searchValue != "" {
		countQuery += " AND (p.product_name ILIKE '%' || $1 || '%' OR b.brand_name ILIKE '%' || $1 || '%')"
	}

	if searchValue != "" {
		err = db.QueryRow(countQuery, searchValue).Scan(&totalProducts)
	} else {
		err = db.QueryRow(countQuery).Scan(&totalProducts)
	}
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}

	// Fetch the user id from the token
	userID, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(500, gin.H{"Error": "Error in string conversion"})
		return
	}

	// Fetch products for the current page
	productQuery := `
		SELECT p.id, p.product_name, p.description, p.price, p.stock, i.image, b.brand_name, c.category_name
		FROM product p 
		LEFT JOIN image i ON p.id = i.product_id
		LEFT JOIN brand b ON p.brand_id = b.id
		LEFT JOIN category c ON p.category_id = c.id
		WHERE 1=1`
	if categoryID > 0 {
		productQuery += " AND p.category_id = " + strconv.Itoa(categoryID)
	}
	if brandID > 0 {
		productQuery += " AND p.brand_id = " + strconv.Itoa(brandID)
	}
	if searchValue != "" {
		productQuery += " AND (p.product_name ILIKE '%' || $3 || '%' OR b.brand_name ILIKE '%' || $3 || '%')"
	}
	productQuery += " LIMIT $1 OFFSET $2"

	var rows *sql.Rows
	if searchValue != "" {
		rows, err = db.Query(productQuery, productsPerPage, offset, searchValue)
	} else {
		rows, err = db.Query(productQuery, productsPerPage, offset)
	}
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}
	defer rows.Close()

	var products []struct {
		ID           int
		ProductName  string
		Description  string
		Price        float64
		Stock        int
		Image        string
		BrandName    string
		CategoryName string
		InCart       bool
		InWishlist   bool
	}

	for rows.Next() {
		var product struct {
			ID           int
			ProductName  string
			Description  string
			Price        float64
			Stock        int
			Image        string
			BrandName    string
			CategoryName string
			InCart       bool
			InWishlist   bool
		}
		err := rows.Scan(&product.ID, &product.ProductName, &product.Description, &product.Price, &product.Stock, &product.Image, &product.BrandName, &product.CategoryName)
		if err != nil {
			c.JSON(500, gin.H{
				"Error": "Error scanning product data",
			})
			return
		}

		// Check if the product is in the cart
		var count1 int
		err = db.QueryRow("SELECT COUNT(*) FROM cart WHERE product_id = $1 AND user_id = $2", product.ID, userID).Scan(&count1)
		if err != nil {
			c.JSON(500, gin.H{
				"Error": "Database query error",
			})
			return
		}

		product.InCart = count1 > 0

		// Check if the product is in the wishlist
		var count2 int
		err = db.QueryRow("SELECT COUNT(*) FROM wishlist WHERE product_id = $1 AND user_id = $2", product.ID, userID).Scan(&count2)
		if err != nil {
			c.JSON(500, gin.H{
				"Error": "Database query error",
			})
			return
		}

		product.InWishlist = count2 > 0
		products = append(products, product)
	}

	totalPages := (totalProducts + productsPerPage - 1) / productsPerPage

	// Fetch categories and brands for filters
	var categories []models.Category

	rows, err = db.Query("SELECT id, category_name FROM category")
	if err != nil {
		c.JSON(500, gin.H{"Error": "Database query error"})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.CategoryName); err != nil {
			c.JSON(500, gin.H{"Error": "Error scanning category data"})
			return
		}
		categories = append(categories, category)
	}

	var brands []models.Brand
	rows, err = db.Query("SELECT id, brand_name FROM brand")
	if err != nil {
		c.JSON(500, gin.H{"Error": "Database query error"})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var brand models.Brand
		if err := rows.Scan(&brand.ID, &brand.BrandName); err != nil {
			c.JSON(500, gin.H{"Error": "Error scanning brand data"})
			return
		}
		brands = append(brands, brand)
	}

	c.HTML(http.StatusOK, "user_home.html", gin.H{
		"Products":         products,
		"CurrentPage":      page,
		"TotalPages":       totalPages,
		"Categories":       categories,
		"Brands":           brands,
		"SelectedCategory": categoryID,
		"SelectedBrand":    brandID,
		"SearchValue":      searchValue,
	})
}