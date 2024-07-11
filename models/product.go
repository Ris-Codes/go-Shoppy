package models

import "time"

type Category struct {
	ID           uint `json:"id"`
	CategoryName string
}

type Product struct {
	ID          uint   `json:"id" form:"id"`
	ProductName string `json:"productname" form:"product_name"`
	Description string `json:"description" form:"description"`
	Stock       uint   `json:"stock" form:"stock"`
	Price       uint   `json:"price" form:"price"`
	Category    Category
	CategoryId  uint `json:"category_id" form:"category_id"`
	Brand       Brand
	BrandId     uint `json:"brand_id" form:"brand_id"`
	Image       Image
}

type Brand struct {
	ID        uint   `json:"id" form:"id"`
	BrandName string `json:"brand_name" form:"brand_name"`
}

type Cart struct {
	ID         uint `json:"id"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Product    Product
	ProductId  uint
	Quantity   uint
	Price      uint
	TotalPrice uint
	Userid     uint
	User       User
}

type Image struct {
	ID        uint   `json:"id"`
	ProductId uint   `json:"product_id"`
	Image     string `json:"image"`
}

type Wishlist struct {
	ID        uint `json:"id"`
	Userid    uint
	User      User
	Product   Product
	ProductId uint `json:"product_id"`
	Category  Category
	Brand     Brand
	Image     Image
	InCart	  bool
}
