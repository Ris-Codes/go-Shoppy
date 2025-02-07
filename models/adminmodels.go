package models

type Admin struct {
	ID          uint   `json:"id"`
	Firstname   string `json:"first_name"`
	Lastname    string `json:"last_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber int    `json:"phone"`
	IsAdmin     bool   `JSON:"isadmin"`
}
