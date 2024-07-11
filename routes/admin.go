package routes

import (
	"net/http"

	"github.com/Ris-Codes/go-Shoppy/controllers"
	middleware "github.com/Ris-Codes/go-Shoppy/middleware"
	"github.com/gin-gonic/gin"
)

func AdminRouts(c *gin.Engine) {
	admin := c.Group("/admin")
	{
		//Admin rounts
		admin.POST("/login", controllers.AdminLogin)
		admin.GET("/login", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin_login.html", nil)
		})
		admin.POST("/registernewadmin", controllers.AdminSignup)
		admin.GET("/registernewadmin", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin_signup.html", nil)
		})
		admin.GET("/logout", middleware.AdminAuth, controllers.AdminSignout)
		admin.GET("/profile", middleware.AdminAuth, controllers.AdminProfile)
		admin.GET("/adminvalidate", middleware.AdminAuth, controllers.ValidateAdmin)
		admin.GET("/adminpanel", middleware.AdminAuth, func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin_panel.html", nil)
		})

		//User management routes
		admin.GET("/user/viewuser", middleware.AdminAuth, controllers.ViewAllUser)
		admin.GET("/user/searchuser", middleware.AdminAuth, controllers.AdminSearchUser)
		admin.POST("/user/edituserprofile/:id", middleware.AdminAuth, controllers.EditUserProfileByadmin)
		admin.POST("/user/blockuser", middleware.AdminAuth, controllers.AdminBlockUser)
		admin.GET("/user/getuserprofile", middleware.AdminAuth, controllers.GetUserProfile)
		
		//specification management routes
		admin.POST("/brand/editbrand/:id", middleware.AdminAuth, controllers.EditBrand)
		admin.GET("/brand", middleware.AdminAuth, controllers.ViewBrand)
		
		//product management
		admin.POST("/addbrand", middleware.AdminAuth, controllers.AddBrands)
		admin.POST("/addcategories", middleware.AdminAuth, controllers.AddCategories)
		admin.GET("/viewproducts", middleware.AdminAuth, controllers.ViewProducts)
		admin.GET("/searchproducts", middleware.AdminAuth, controllers.SearchProductByAdmin)
		admin.POST("/addproduct", middleware.AdminAuth, controllers.AddProduct)
		admin.GET("/editproduct/:id", middleware.AdminAuth ,controllers.EditProductPage)
		admin.POST("/editproduct/:id", middleware.AdminAuth, controllers.EditProduct)
		admin.POST("/product/addimage", middleware.AdminAuth, controllers.AddImages)

		//Order management
		admin.GET("/order_management", middleware.AdminAuth, controllers.AdminOrderManagement)
		admin.POST("/update_order", middleware.AdminAuth, controllers.UpdateOrderStatus)
		admin.POST("/update_payment_status", middleware.AdminAuth, controllers.UpdatePaymentStatus)

		//Salse Report
		admin.GET("/salesreport", middleware.AdminAuth, controllers.SalesReport)
		admin.GET("/salesreport/download/excel", middleware.AdminAuth, controllers.DownloadExcel)
		admin.GET("/salesreport/download/pdf", middleware.AdminAuth, controllers.DownloadPDF)
	}
}
