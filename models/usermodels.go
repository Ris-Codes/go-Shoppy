package models

import "time"

type User struct {
	ID          int    `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber int    `json:"phone"`
	IsAdmin     bool   `json:"isadmin"`
	Otp         string `json:"otp"`
	Isblocked   bool   `json:"isblocked"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Address struct {
	Addressid  int `json:"addressid"`
	User       User
	Userid     int    `json:"uid"`
	Name       string `json:"name" form:"name"`
	Phoneno    string `json:"phoneno" form:"phoneno"`
	Houseno    string `json:"houseno" form:"houseno"`
	Area       string `json:"area" form:"area"`
	Landmark   string `json:"landmark" form:"landmark"`
	City       string `json:"city" form:"city"`
	Pincode    string `json:"pincode" form:"pincode"`
	District   string `json:"district" form:"district"`
	State      string `json:"state" form:"state"`
	Country    string `json:"country" form:"country"`
	Defaultadd bool   `json:"defaultadd" form:"defaultadd"`
}
