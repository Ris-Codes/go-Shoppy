package controllers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
)

func ShowOrder(c *gin.Context) {
	userId, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	db := config.DB

	// Fetch user orders
	rows, err := db.Query("SELECT id, user_id, order_status FROM orders WHERE user_id = $1 ORDER BY id DESC", userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error fetching user orders: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	type OrderItemData struct {
		Product  string
		Price    float64
		Quantity int
	}

	type OrderData struct {
		OrderID int
		Status  string
		Items   []OrderItemData
	}
	var ordersData []OrderData

	// Iterate over each order
	for rows.Next() {
		var orderData OrderData
		err := rows.Scan(&orderData.OrderID, &userId, &orderData.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Error scanning user orders: " + err.Error(),
			})
			return
		}

		// Fetch items for each order
		itemRows, err := db.Query("SELECT product_id, quantity FROM order_item WHERE order_id = $1", orderData.OrderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Error fetching order items: " + err.Error(),
			})
			return
		}
		defer itemRows.Close()

		var items []OrderItemData
		for itemRows.Next() {
			var productId, quantity int
			err := itemRows.Scan(&productId, &quantity)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": "Error scanning order items: " + err.Error(),
				})
				return
			}

			var product models.Product
			err = db.QueryRow("SELECT id, product_name, price FROM product WHERE id = $1", productId).
				Scan(&product.ID, &product.ProductName, &product.Price)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": "Error fetching product details: " + err.Error(),
				})
				return
			}

			orderItem := OrderItemData{
				Product:  product.ProductName,
				Price:    float64(product.Price),
				Quantity: quantity,
			}

			items = append(items, orderItem)
		}

		if err := itemRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Error with order items rows: " + err.Error(),
			})
			return
		}

		orderData.Items = items
		ordersData = append(ordersData, orderData)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error with user orders rows: " + err.Error(),
		})
		return
	}

	// Render orders page using template
	c.HTML(http.StatusOK, "order.html", gin.H{
		"Orders": ordersData,
	})
}

func AdminOrderManagement(c *gin.Context) {
	db := config.DB

	rows, err := db.Query("SELECT o.id, u.firstname || ' ' || u.lastname AS user_name, 	p.product_name, oi.quantity, o.total_amount, o.order_status, pay.payment_method, pay.status AS payment_status, a.city || ', ' || a.state || ', ' || a.country AS address, o.created_at FROM orders o JOIN users u ON o.user_id = u.id JOIN order_item oi ON o.id = oi.order_id JOIN product p ON oi.product_id = p.id JOIN payment pay ON o.payment_id = pay.id JOIN address a ON o.address_id = a.id ORDER BY o.id DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type Product struct {
		ProductName string
		Quantity    int
	}

	type Order struct {
		OrderID       int
		UserName      string
		TotalAmount   float64
		OrderStatus   string
		PaymentMethod string
		PaymentStatus string
		Address       string
		CreatedAt     string
		Products      []Product
	}

	orderMap := make(map[int]*Order)
	for rows.Next() {
		var (
			orderID       int
			userName      string
			productName   string
			quantity      int
			totalAmount   float64
			orderStatus   string
			paymentMethod string
			paymentStatus string
			address       string
			createdAt     time.Time
		)
		err := rows.Scan(&orderID, &userName, &productName, &quantity, &totalAmount, &orderStatus, &paymentMethod, &paymentStatus, &address, &createdAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning data"})
			return
		}

		order, exists := orderMap[orderID]
		if !exists {
			order = &Order{
				OrderID:       orderID,
				UserName:      userName,
				TotalAmount:   totalAmount,
				OrderStatus:   orderStatus,
				PaymentMethod: paymentMethod,
				PaymentStatus: paymentStatus,
				Address:       address,
				CreatedAt:     createdAt.Format("02-01-2006 3:04PM"),
				Products:      []Product{},
			}
			orderMap[orderID] = order
		}
		order.Products = append(order.Products, Product{ProductName: productName, Quantity: quantity})
	}

	var orders []Order
	for _, order := range orderMap {
		orders = append(orders, *order)
	}

	// Sorting the orders slice by orderID in descending order
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].OrderID > orders[j].OrderID
	})

	c.HTML(http.StatusOK, "admin_order_management.html", gin.H{
		"Orders": orders,
	})
}

func UpdateOrderStatus(c *gin.Context) {
	orderID := c.PostForm("order_id")
	newStatus := c.PostForm("order_status")

	db := config.DB

	// Update order status
	_, err := db.Exec("UPDATE orders SET order_status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", newStatus, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// If the order status is updated to "Processing", update product stocks
	if newStatus == "Processing" {
		// Retrieve all products in the order
		rows, err := db.Query("SELECT product_id, quantity FROM order_item WHERE order_id = $1", orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order items"})
			return
		}
		defer rows.Close()

		type OrderItem struct {
			ProductID int
			Quantity  float64
		}

		var orderItems []OrderItem

		for rows.Next() {
			var oi OrderItem
			err := rows.Scan(&oi.ProductID, &oi.Quantity)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning order items"})
				return
			}
			orderItems = append(orderItems, oi)
		}

		// Update product stocks
		for _, item := range orderItems {
			_, err := db.Exec("UPDATE product SET stock = stock - $1 WHERE id = $2", item.Quantity, item.ProductID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
				return
			}
		}
	}

	c.Redirect(http.StatusFound, "/admin/order_management")
}