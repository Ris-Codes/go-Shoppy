package models

import (
	"time"
)

type Payment struct {
	PaymentId     uint `json:"payment_id" form:"payment_id"`
	User          User
	UserId        uint
	PaymentMethod string `json:"payment_method" form:"payment_method"`
	Totalamount   uint   `json:"total_amount" form:"total_amount"`
	Status        string `json:"status" form:"status"`
}

type OrderDetails struct {
	Orderid     int `json:"oderid"`
	UserId      uint
	User        User
	AddressId   uint
	Address     Address
	PaymentId   uint
	Payment     Payment
	OderItemId  uint
	ProductId   uint
	Product     Product
	Quantity    uint
	TotalAmount uint   `json:"total_amount"`
	Status      string `json:"status"`
	CreatedAt   time.Time	
	UpdatedAt   time.Time
}

type RazorPay struct {
	UserID          uint   `json:"user_id" form:"user_id"`
	RazorPaymentId  string `json:"razorpaymentid" form:"razorpaymentid"`
	RazorPayOrderID string `json:"razorpayorderid" form:"razorpayorderid"`
	Signature       string `json:"signature" form:"signature"`
	AmountPaid      string `json:"amountpaid" form:"amountpaid"`
}