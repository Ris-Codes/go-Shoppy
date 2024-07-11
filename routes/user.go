package routes

import (
	"github.com/Ris-Codes/go-Shoppy/controllers"
	middleware "github.com/Ris-Codes/go-Shoppy/middleware"
	"github.com/gin-gonic/gin"
)

func UserRouts(c *gin.Engine) {
	c.GET("/", middleware.UserAuth, controllers.UserHome)
	User := c.Group("/user")
	{

		//User rountes >>
		User.POST("/login", controllers.UserLogin)
		User.GET("/login", func(c *gin.Context) {
			c.HTML(200, "user_login.html", nil)
		})
		User.POST("/signup", controllers.UserSignUP)
		User.GET("/signup", func(c *gin.Context) {
			c.HTML(200, "user_signup.html", nil)
		})
		User.POST("/signup/otpvalidate", controllers.OtpValidation)
		User.GET("/signup/otpvalidate", func(c *gin.Context) {
			c.HTML(200, "otpvalidate.html", nil)
		})
		User.GET("/logout", middleware.UserAuth, controllers.UserSignout)
		User.GET("/uservalidate", middleware.UserAuth, controllers.UserValidate)

		//User profile routes
		User.GET("/viewprofile", middleware.UserAuth, controllers.ShowUserDetails)
		User.POST("/addaddress", middleware.UserAuth, controllers.AddAddress)
		User.GET("/editaddresspage/:id", middleware.UserAuth, controllers.EditAddressPage)
		User.POST("/editaddress/:id", middleware.UserAuth, controllers.EditUserAddress)
		User.GET("/setdefault/:id", middleware.UserAuth, controllers.SetDefaultAddress)
		User.GET("/deleteaddress/:id", middleware.UserAuth, controllers.DeleteUserAddress)
		User.GET("/showaddresses", middleware.UserAuth, controllers.ShowAddress)
		User.GET("/userchangepassword", middleware.UserAuth, func(c *gin.Context) {
			c.HTML(200, "user_changepassword.html", nil)
		})
		User.POST("/changepassword", middleware.UserAuth, controllers.UserChangePassword)
		User.POST("/editprofile", middleware.UserAuth, controllers.EditUserProfilebyUser)
		User.GET("/editprofile", middleware.UserAuth, controllers.EditUserProfilePage)

		// Wishlist
		User.GET("/wishlist", middleware.UserAuth, controllers.Wishlist)
		User.POST("/addtowishlist", middleware.UserAuth, controllers.AddToWishlist)
		User.POST("/removefromwishlist", middleware.UserAuth, controllers.RemoveFromWishlist)

		//User product managment
		User.GET("/viewproducts", middleware.UserAuth, controllers.ViewProducts)

		//User carts routes
		User.GET("/viewcart", middleware.UserAuth, controllers.ViewCart)
		User.POST("/addtocart", middleware.UserAuth, controllers.AddToCart)
		User.POST("/cart/update", middleware.UserAuth, controllers.UpdateCart)
		User.POST("/cart/remove", middleware.UserAuth, controllers.RemoveFromCart)
		User.POST("/cart/empty", middleware.UserAuth, controllers.EmptyCart)
		User.GET("/cart/checkout", middleware.UserAuth, controllers.CheckOut)

		// Order managements by user
		User.GET("/showorder", middleware.UserAuth, controllers.ShowOrder)

		//Forgot Password
		User.POST("/forgotpassword/changepassword", middleware.UserAuth, controllers.ChangePassword)

		//payments route
		User.GET("/payment", middleware.UserAuth, func(c *gin.Context) {
			c.HTML(200, "payment.html", nil)
		})
		User.GET("/payment/cashondelivery", middleware.UserAuth, controllers.CashOnDelivery)
		User.GET("/codsuccess", middleware.UserAuth, func(c *gin.Context) {
			c.HTML(200, "codsuccess.html", nil)
		})
		User.GET("/payment/razorpay", middleware.UserAuth, controllers.Razorpay)
		User.GET("/payment/success", middleware.UserAuth, controllers.RazorpaySuccess)
		User.GET("/success", middleware.UserAuth, func(c *gin.Context) {
			c.HTML(200, "success.html", nil)
		})

		//invoice
		User.GET("/invoice", middleware.UserAuth, controllers.InvoiceF)
	}
}
