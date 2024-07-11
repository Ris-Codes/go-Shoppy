package controllers

import (
	"bytes"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/razorpay/razorpay-go"
)

func CashOnDelivery(c *gin.Context) {
	id, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	db := config.DB

	// Fetch cart data
	var cartExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM cart WHERE user_id = $1)", id).Scan(&cartExists)
	if err != nil || !cartExists {
		c.JSON(404, gin.H{
			"Message": "Cart is empty",
		})
		return
	}

	// Fetch total amount
	var totalAmount float64
	err = db.QueryRow("SELECT SUM(total_price) FROM cart WHERE user_id = $1", id).Scan(&totalAmount)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error fetching the total amount from the cart",
		})
		return
	}

	// Insert payment data
	var paymentId int
	err = db.QueryRow("INSERT INTO payment (user_id, payment_method, total_amount , status) VALUES ($1, $2, $3, $4) RETURNING id", id, "COD", totalAmount, "Pending").Scan(&paymentId)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": err.Error(),
		})
		return
	}

	// Fetch address data
	var addressId int
	err = db.QueryRow("SELECT id FROM address WHERE user_id = $1 AND default_add = true", id).Scan(&addressId)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Address not exist",
		})
		return
	}

	// Insert order data
	var orderId int
	err = db.QueryRow("INSERT INTO orders (user_id, total_amount, payment_id, address_id, order_status) VALUES ($1, $2, $3, $4, $5) RETURNING id", id, totalAmount, paymentId, addressId, "pending").Scan(&orderId)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Creating the order",
		})
		return
	}

	// Insert order item data
	rows, err := db.Query("SELECT product_id, quantity FROM cart WHERE user_id = $1", id)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error fetching items from cart",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var productID, quantity int
		err := rows.Scan(&productID, &quantity)
		if err != nil {
			c.JSON(400, gin.H{
				"Error": "Error scanning cart items",
			})
			return
		}

		_, err = db.Exec("INSERT INTO order_item (order_id, product_id, quantity) VALUES ($1, $2, $3)", orderId, productID, quantity)
		if err != nil {
			c.JSON(400, gin.H{
				"Error": "Creating the order item",
			})
			return
		}
	}

	c.Redirect(http.StatusSeeOther, "/user/codsuccess")

	EmptyCart(c)
}

func Razorpay(c *gin.Context) {
	id, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}

	db := config.DB

	// Fetch user data
	var userdata models.User
	err = db.QueryRow("SELECT id, firstname, lastname, email, phone FROM users WHERE id = $1", id).Scan(&userdata.ID, &userdata.FirstName, &userdata.LastName, &userdata.Email, &userdata.PhoneNumber)
	if err != nil {
		c.JSON(404, gin.H{
			"Error": err.Error(),
		})
		return
	}

	// Fetch total amount from cart
	var totalAmount float64
	err = db.QueryRow("SELECT SUM(total_price) FROM cart WHERE user_id = $1", id).Scan(&totalAmount)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": err.Error(),
		})
		return
	}

	// Sending payment details to Razorpay
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_SECRET"))
	data := map[string]interface{}{
		"amount":   totalAmount * 100, // amount in paise
		"currency": "INR",
		"receipt":  "some_receipt_id",
	}

	// Creating payment details for client order
	body, err := client.Order.Create(data, nil)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": err,
		})
		return
	}

	// To render the HTML page with user & payment details
	orderId := body["id"].(string)

	c.HTML(200, "app.html", gin.H{
		"user":       userdata,
		"totalprice": totalAmount,
		"orderid":    orderId,
	})
}

func RazorpaySuccess(c *gin.Context) {
	userIDStr := c.Query("userid")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "userid parameter is empty",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Invalid userid format: " + err.Error(),
		})
		return
	}

	db := config.DB

	// Fetching the payment details from Razorpay
	orderID := c.Query("order_id")
	paymentID := c.Query("payment_id")
	signature := c.Query("signature")
	totalAmount := c.Query("total")

	if orderID == "" || paymentID == "" || signature == "" || totalAmount == "" {
		c.JSON(400, gin.H{
			"Error": "Missing required parameters",
		})
		return
	}

	// Creating table razorpay using the data from Razorpay
	_, err = db.Exec("INSERT INTO razorpay (user_id, razorpay_paymentid, signature, razorpay_orderid, amount_paid) VALUES ($1, $2, $3, $4, $5)", userID, paymentID, signature, orderID, totalAmount)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error inserting into razorpay: " + err.Error(),
		})
		return
	}

	method := "RazorPay"
	status := "pending"

	// Converting to string total amount variable
	totalPrice, err := strconv.Atoi(totalAmount)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error in total amount conversion",
		})
		return
	}

	// Creating payment table
	err = db.QueryRow("INSERT INTO payment (user_id, payment_method, status, total_amount) VALUES ($1, $2, $3, $4) RETURNING id", userID, method, status, totalPrice).Scan(&paymentID)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error inserting into payment: " + err.Error(),
		})
		return
	}

	// Fetching address data
	var addressID int
	err = db.QueryRow("SELECT id FROM address WHERE user_id = $1 AND default_add = true",
		userID).Scan(&addressID)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error fetching address: " + err.Error(),
		})
		return
	}

	// Creating order table
	err = db.QueryRow("INSERT INTO orders (user_id, total_amount, payment_id, order_status, address_id) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userID, totalPrice, paymentID, "pending", addressID).Scan(&orderID)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error inserting into orders: " + err.Error(),
		})
		return
	}

	// Inserting cart items into order_item table
	rows, err := db.Query("SELECT product_id, quantity FROM cart WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Error fetching cart items: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var productID int
		var quantity float64
		err = rows.Scan(&productID, &quantity)
		if err != nil {
			c.JSON(400, gin.H{
				"Error": "Error scanning cart items: " + err.Error(),
			})
			return
		}
		_, err = db.Exec("INSERT INTO order_item (order_id, product_id, quantity) VALUES ($1, $2, $3)", orderID, productID, quantity)
		if err != nil {
			c.JSON(400, gin.H{
				"Error": "Error inserting into order_item: " + err.Error(),
			})
			return
		}
	}

	EmptyCart(c)

	// Redirect to success page with payment ID
	c.Redirect(http.StatusSeeOther, "/user/success")
}

// Update payment status by admin
func UpdatePaymentStatus(c *gin.Context) {
	orderID := c.PostForm("order_id")
	newStatus := c.PostForm("payment_status")

	db := config.DB
	_, err := db.Exec("UPDATE payment SET status = $1 WHERE id = (SELECT payment_id FROM orders WHERE id = $2)", newStatus, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.Redirect(http.StatusFound, "/admin/order_management")
}


// Templates for creating the pdf
func InvoiceF(c *gin.Context) {
	id, err := strconv.Atoi(c.GetString("userid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Invalid user ID",
		})
		return
	}

	type OrderItem struct {
		ID          int       `json:"id"`
		UserID      int       `json:"user_id"`
		TotalAmount float64   `json:"total_amount"`
		PaymentID   int       `json:"payment_id"`
		OrderStatus string    `json:"order_status"`
		AddressID   int       `json:"address_id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	db := config.DB
	var user models.User
	var payment models.Payment
	var orderItem OrderItem
	var address models.Address
	var products []models.Product

	// Fetching the latest order for the user
	err = db.QueryRow("SELECT * FROM orders WHERE user_id = $1 ORDER BY id DESC LIMIT 1", id).Scan(&orderItem.ID, &orderItem.UserID, &orderItem.TotalAmount, &orderItem.PaymentID, &orderItem.OrderStatus, &orderItem.AddressID, &orderItem.CreatedAt, &orderItem.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error fetching order details: " + err.Error(),
		})
		return
	}

	// Fetching the user data
	err = db.QueryRow("SELECT id, firstname, lastname, email, phone FROM users WHERE id = $1", id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error fetching user data: " + err.Error(),
		})
		return
	}

	// Fetching the address data
	err = db.QueryRow("SELECT id, name, phone, house_no, area, landmark, city, pincode, district, state, country FROM address WHERE id = $1", orderItem.AddressID).Scan(&address.Addressid, &address.Name, &address.Phoneno, &address.Houseno, &address.Area, &address.Landmark, &address.City, &address.Pincode, &address.District, &address.State, &address.Country)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error fetching address data: " + err.Error(),
		})
		return
	}

	// Fetching the payment data
	err = db.QueryRow("SELECT id, user_id, payment_method, total_amount, status FROM payment WHERE id = $1", orderItem.PaymentID).Scan(&payment.PaymentId, &payment.UserId, &payment.PaymentMethod, &payment.Totalamount, &payment.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error fetching payment data: " + err.Error(),
		})
		return
	}

	// Fetching products related to the order item
	rows, err := db.Query("SELECT p.id, p.product_name, p.description, p.stock, p.price, p.category_id, p.brand_id FROM product p INNER JOIN order_item oi ON p.id = oi.product_id WHERE oi.order_id = $1", orderItem.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error fetching products: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.ProductName, &p.Description, &p.Stock, &p.Price, &p.CategoryId, &p.BrandId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Error scanning product: " + err.Error(),
			})
			return
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error with product rows: " + err.Error(),
		})
		return
	}

	// Generating the invoice data
	type Item struct {
		Product     string `json:"product"`
		Price       uint   `json:"price"`
		Description string `json:"description"`
	}

	items := make([]Item, len(products))
	for i, p := range products {
		items[i] = Item{
			Product:     p.ProductName,
			Price:       p.Price,
			Description: p.Description,
		}
	}

	timeString := orderItem.CreatedAt.Format("02-01-2006")

	type Invoice struct {
		BusinessName  string
		Name          string
		Date          string
		Email         string
		OrderId       int
		PaymentMethod string
		Totalamount   float64
		Address       models.Address
		Items         []Item
	}

	businessName := "Go Shoppy"

	invoice := Invoice{
		BusinessName:  businessName,
		Name:          user.FirstName + " " + user.LastName,
		Date:          timeString,
		Email:         user.Email,
		OrderId:       orderItem.ID,
		PaymentMethod: payment.PaymentMethod,
		Totalamount:   orderItem.TotalAmount,
		Address: models.Address{
			Phoneno:  address.Phoneno,
			Houseno:  address.Houseno,
			Area:     address.Area,
			Landmark: address.Landmark,
			City:     address.City,
			Pincode:  address.Pincode,
			District: address.District,
			State:    address.State,
			Country:  address.Country,
		},
		Items: items,
	}

	// Create a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// Title
	pdf.CellFormat(190, 10, invoice.BusinessName, "", 0, "C", false, 0, "")
	pdf.Ln(10)
	pdf.Cell(190, 10, "Invoice")
	pdf.Ln(10)

	// Invoice Details Section
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Order ID: "+strconv.Itoa(invoice.OrderId))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Order Date: "+invoice.Date)
	pdf.Ln(8)
	pdf.Cell(40, 10, "Name: "+invoice.Name)
	pdf.Ln(8)
	pdf.Cell(40, 10, "Email: "+invoice.Email)
	pdf.Ln(8)

	// Address Section
	pdf.Cell(40, 10, "Billing Address:")
	pdf.Ln(8)
	pdf.Cell(40, 10, invoice.Address.Houseno+", "+invoice.Address.Area)
	pdf.Ln(8)
	pdf.Cell(40, 10, invoice.Address.Landmark+", "+invoice.Address.City+", "+invoice.Address.Pincode)
	pdf.Ln(8)
	pdf.Cell(40, 10, invoice.Address.Phoneno)
	pdf.Ln(8)
	pdf.Cell(40, 10, invoice.Address.District+", "+invoice.Address.State)
	pdf.Ln(8)
	pdf.Cell(40, 10, invoice.Address.Country)
	pdf.Ln(8)

	// Payment Method Section
	pdf.Cell(40, 10, "Payment method: "+invoice.PaymentMethod)
	pdf.Ln(8)

	// Products
	pdf.Cell(40, 10, "Items:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(70, 7, "Product", "1", 0, "", false, 0, "")
	pdf.CellFormat(50, 7, "Description", "1", 0, "", false, 0, "")
	pdf.CellFormat(25, 7, "Price", "1", 0, "", false, 0, "")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 12)
	for _, item := range invoice.Items {
		pdf.CellFormat(70, 7, item.Product, "1", 0, "", false, 0, "")
		pdf.CellFormat(50, 7, item.Description, "1", 0, "", false, 0, "")
		pdf.CellFormat(25, 7, strconv.Itoa(int(item.Price)), "1", 0, "", false, 0, "")
		pdf.Ln(7)
	}

	// Total Amount
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Total Amount: ")
	pdf.CellFormat(40, 10, strconv.FormatFloat(invoice.Totalamount, 'f', 2, 64), "", 0, "", false, 0, "")
	pdf.Ln(10)

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Error generating PDF: " + err.Error(),
		})
		return
	}

	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
