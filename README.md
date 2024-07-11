# Go Shoppy - eCommerce Application
This is an eCommerce Web Application Created using Golang as Backend using Gin Framework and PostgreSQL as Database to manage the data.
## Framework Used
Gin-Gonic: This project is built using Gin framework. Its is a popular http web framework. 
#### <kbd> go get -u github.com/gin-gonic/gin </kbd>

## Database used:
PostgreSQL: PostgreSQL is a powerful, open source object-relational database. The data managment in this project is done using PostgreSQL. Raw sql queries are passed with the functions without the help of any ORMs. 

```go
import (
	"database/sql"
	_ "github.com/lib/pq"
)
```

## Templates
HTML Templates are used to render the data and pass the data from and to the database. Simple eCommerce frontent is provided in the `templates` directory

## External Packages Used
#### Razorpay
For Payment I have used the test case of Razorpay.
#### <kbd>github.com/razorpay/razorpay-go </kbd>

#### Gomail
Gomail is a simple and efficient package to send emails. It is well tested and documented.
#### <kbd> gopkg.in/mail.v2 </kbd>

#### JWT 
JSON Web Tokens are an open, industry standard RFC 7519 method for representing claims securely between two parties.
#### <kbd> github.com/golang-jwt/jwt/v4  </kbd>

#### Commands to run project:
#### <kbd> go run main.go </kbd>