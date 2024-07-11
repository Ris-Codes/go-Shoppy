package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
)

// Adding the product to the cart
func AddToCart(c *gin.Context) {
	type data struct {
		ProductID uint `form:"product_id" json:"product_id"`
		Quantity  uint `form:"quantity" json:"quantity"`
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

	// Fetching the user id from the token
	id, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	db := config.DB

	// Checking if the product exists
	err = db.QueryRow("SELECT id, price, stock FROM product WHERE id = $1", bindData.ProductID).Scan(&productData.ID, &productData.Price, &productData.Stock)
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

	// Checking stock quantity
	if bindData.Quantity > productData.Stock {
		c.JSON(404, gin.H{
			"Message": "Out of Stock",
		})
		return
	}

	var sum uint
	var totalPrice float64

	// Checking if the product is already in the cart
	err = db.QueryRow("SELECT quantity, total_price FROM cart WHERE product_id = $1 AND user_id = $2", bindData.ProductID, id).Scan(&sum, &totalPrice)
	if err == sql.ErrNoRows {
		totalPrice = float64(bindData.Quantity) * float64(productData.Price)
		_, err = db.Exec("INSERT INTO cart (product_id, quantity, price, total_price, user_id) VALUES ($1, $2, $3, $4, $5)",
			bindData.ProductID, bindData.Quantity, productData.Price, totalPrice, id)
		if err != nil {
			c.JSON(500, gin.H{
				"Error": "Error adding product to cart",
			})
			return
		}

		c.JSON(200, gin.H{
			"Message": "Added to the Cart Successfully",
		})
		return
	} else if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}

	// Calculating the total quantity and total price
	totalQuantity := sum + bindData.Quantity
	totalPrice = float64(totalQuantity) * float64(productData.Price)

	// Updating the quantity and the total price in the carts
	_, err = db.Exec("UPDATE cart SET quantity = $1, total_price = $2 WHERE product_id = $3 AND userid = $4",
		totalQuantity, totalPrice, bindData.ProductID, id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error updating cart",
		})
		return
	}

	c.JSON(200, gin.H{
		"Message": "Quantity added Successfully",
	})
}

// View cart items using user id
func ViewCart(c *gin.Context) {
	id, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	type Cartdata struct {
		ProductID   uint   
		ProductName string 
		Quantity    uint   
		TotalPrice  float64   
		Image       string 
		Price       float64 
		Stock		uint  
	}

	var datas []Cartdata
	var totalQuantity uint
	var totalPrice float64

	db := config.DB

	rows, err := db.Query("SELECT p.id, p.product_name, p.stock, i.image, c.quantity, c.price, c.total_price FROM cart c JOIN product p ON c.product_id = p.id LEFT JOIN image i ON p.id = i.product_id WHERE c.user_id = $1", id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Database query error",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var data Cartdata
		err := rows.Scan(&data.ProductID, &data.ProductName, &data.Stock, &data.Image, &data.Quantity, &data.Price, &data.TotalPrice)
		if err != nil {
			c.JSON(500, gin.H{
				"Error": "Failed to scan row",
			})
			return
		}
		totalQuantity += data.Quantity
		totalPrice += data.TotalPrice
		datas = append(datas, data)
	}

	if len(datas) == 0 {
		c.HTML(http.StatusOK, "cart.html", gin.H{
			"Message": "Cart is empty",
		})
		return
	}

	c.HTML(http.StatusOK, "cart.html", gin.H{
		"CartItems": datas,
		"TotalQuantity": totalQuantity,
		"TotalPrice":    totalPrice,
	})
}

func UpdateCart(c *gin.Context) {
	type data struct {
		ProductID uint `json:"product_id"`
		Quantity  uint `json:"quantity"`
	}
	var bindData data

	if err := c.ShouldBindJSON(&bindData); err != nil {
		c.JSON(400, gin.H{
			"Error": "Could not bind the data",
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

	var productPrice float64
	var productStock int
	err = db.QueryRow("SELECT price, stock FROM product WHERE id = $1", bindData.ProductID).Scan(&productPrice, &productStock)
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

	if bindData.Quantity > uint(productStock) {
		c.JSON(404, gin.H{
			"Message": "Out of Stock",
		})
		return
	}

	totalPrice := productPrice * float64(bindData.Quantity)

	result, err := db.Exec("UPDATE cart SET quantity = $1, total_price = $2 WHERE product_id = $3 AND user_id = $4", bindData.Quantity, totalPrice, bindData.ProductID, id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to update cart",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to retrieve rows affected",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(400, gin.H{
			"Error": "No rows were updated",
		})
		return
	}

	c.JSON(200, gin.H{
		"Message": "Quantity updated successfully",
	})
}

func RemoveFromCart(c *gin.Context) {
	type data struct {
		ProductID uint `json:"product_id"`
	}
	var bindData data

	if err := c.ShouldBindJSON(&bindData); err != nil {
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

	_, err = db.Exec("DELETE FROM cart WHERE product_id = $1 AND user_id = $2", bindData.ProductID, id)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Failed to remove from cart",
		})
		return
	}

	c.JSON(200, gin.H{
		"Message": "Removed from cart successfully",
	})
}

func EmptyCart(c *gin.Context) {
    id, err := strconv.Atoi(c.GetString("userid"))
    if err != nil {
        c.JSON(400, gin.H{
            "Error": "Error in string conversion",
        })
        return
    }

    db := config.DB

    // Delete all items from the cart for the current user
    _, err = db.Exec("DELETE FROM cart WHERE user_id = $1", id)
    if err != nil {
        c.JSON(500, gin.H{
            "Error": "Failed to empty cart",
        })
        return
    }
}

func CheckOut(c *gin.Context) {
    id, err := strconv.Atoi(c.GetString("userid"))
    if err != nil {
        c.JSON(400, gin.H{
            "Error": "Error while string conversion",
        })
        return
    }

    db := config.DB

    // Fetch default address of the user
    var userAddressData struct {
        Name     string
        Phoneno  string
        Houseno  string
        Area     string
        Landmark string
        City     string
        Pincode  string
        District string
        State    string
        Country  string
    }
    err = db.QueryRow("SELECT name, phone, house_no, area, landmark, city, pincode, district, state, country FROM address WHERE default_add = true AND user_id = $1", id).Scan(
        &userAddressData.Name,
        &userAddressData.Phoneno,
        &userAddressData.Houseno,
        &userAddressData.Area,
        &userAddressData.Landmark,
        &userAddressData.City,
        &userAddressData.Pincode,
        &userAddressData.District,
        &userAddressData.State,
        &userAddressData.Country,
    )
    if err == sql.ErrNoRows {
        c.JSON(404, gin.H{
            "Message": "Address does not exist",
        })
        return
    } else if err != nil {
        c.JSON(500, gin.H{
            "Error": err.Error(),
        })
        return
    }

    // Fetch total price from the cart
    var totalPrice float64
    err = db.QueryRow("SELECT SUM(total_price) FROM cart WHERE user_id = $1", id).Scan(&totalPrice)
    if err != nil {
        c.JSON(400, gin.H{
            "Error": "Cannot fetch total amount",
        })
        return
    }

    // Fetch cart items
    type CartItem struct {
        ProductName string
        Quantity    int
        Price       float64
        TotalPrice  float64
    }
    var cartItems []CartItem
    rows, err := db.Query("SELECT p.product_name, c.quantity, c.price, c.total_price FROM cart c JOIN product p ON c.product_id = p.id WHERE c.user_id = $1", id)
    if err != nil {
        c.JSON(500, gin.H{
            "Error": "Database query error",
        })
        return
    }
    defer rows.Close()

    for rows.Next() {
        var item CartItem
        if err := rows.Scan(&item.ProductName, &item.Quantity, &item.Price, &item.TotalPrice); err != nil {
            c.JSON(500, gin.H{
                "Error": "Error scanning cart item data",
            })
            return
        }
        cartItems = append(cartItems, item)
    }

    if err := rows.Err(); err != nil {
        c.JSON(500, gin.H{
            "Error": "Rows error",
        })
        return
    }

    c.HTML(http.StatusOK, "checkout.html", gin.H{
        "DefaultAddress": userAddressData,
        "TotalPrice":     totalPrice,
        "CartItems":      cartItems,
    })
}