package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/models"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/tealeg/xlsx"
)

func SalesReport(c *gin.Context) {
	// Fetching the dates from the URL
	startDate := c.Query("startDate")
	endDateStr := c.Query("endDate")

	if startDate == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start date and end date are required"})
		return
	}

	// Converting the dates string to time.Time
	fromTime, err := time.Parse("02-01-2006", startDate)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid start Date"})
		return
	}
	toTime, err := time.Parse("02-01-2006", endDateStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid end Date"})
		return
	}

	var orderDetail []models.OrderDetails
	db := config.DB

	//fetch order details
	rows, err := db.Query("SELECT o.id AS order_id, o.created_at AS order_date, p.product_name, p.price, oi.quantity, o.total_amount, pay.payment_method, pay.status AS payment_status FROM orders o JOIN order_item oi ON o.id = oi.order_id JOIN product p ON oi.product_id = p.id JOIN payment pay ON o.payment_id = pay.id WHERE o.created_at BETWEEN $1 AND $2 ORDER BY o.created_at DESC", fromTime, toTime)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var order models.OrderDetails
		err := rows.Scan(&order.Orderid, &order.CreatedAt, &order.Product.ProductName, &order.Product.Price, &order.Quantity, &order.TotalAmount, &order.Payment.PaymentMethod, &order.Payment.Status)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error scanning data"})
			return
		}

		orderDetail = append(orderDetail, order)
	}

	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{"error": "Error iterating over rows"})
		return
	}

	// Create new Excel file
	f := excelize.NewFile()
	SheetName := "Sheet1"
	index := f.NewSheet(SheetName)

	// Set headers in Excel sheet
	f.SetCellValue(SheetName, "A1", "Order Date")
	f.SetCellValue(SheetName, "B1", "Order ID")
	f.SetCellValue(SheetName, "C1", "Product name")
	f.SetCellValue(SheetName, "D1", "Price")
	f.SetCellValue(SheetName, "E1", "Total Amount")
	f.SetCellValue(SheetName, "F1", "Payment method")
	f.SetCellValue(SheetName, "G1", "Payment Status")

	// Populate Excel with data
	for i, report := range orderDetail {
		row := i + 2
		f.SetCellValue(SheetName, fmt.Sprintf("A%d", row), report.CreatedAt.Format("02/01/2006"))
		f.SetCellValue(SheetName, fmt.Sprintf("B%d", row), report.Orderid)
		f.SetCellValue(SheetName, fmt.Sprintf("C%d", row), report.Product.ProductName)
		f.SetCellValue(SheetName, fmt.Sprintf("D%d", row), report.Product.Price)
		f.SetCellValue(SheetName, fmt.Sprintf("E%d", row), report.TotalAmount)
		f.SetCellValue(SheetName, fmt.Sprintf("F%d", row), report.Payment.PaymentMethod)
		f.SetCellValue(SheetName, fmt.Sprintf("G%d", row), report.Payment.Status)
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Save Excel file
	if err := f.SaveAs("./public/SalesReport.xlsx"); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save Excel file"})
		return
	}

	// Convert Excel to PDF
	CovertingExelToPdf(c)

	// Return HTML response or redirect
	c.HTML(200, "sales_report.html", gin.H{
		"Orders": orderDetail,
	})
}

func CovertingExelToPdf(c *gin.Context) {
	// Open the Excel file
	xlFile, err := xlsx.OpenFile("./public/SalesReport.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 14)

	// Convertig each cell in the Excel file to a PDF cell
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				//if there is any empty cell values skiping that
				if cell.Value == "" {
					continue
				}

				pdf.Cell(40, 10, cell.Value)
			}
			pdf.Ln(-1)
		}
	}

	// Save the PDF document
	err = pdf.OutputFileAndClose("./public/SalesReport.pdf")
	if err != nil {
		fmt.Println(err)
		return
	}

}

func DownloadExcel(c *gin.Context) {
	c.Header("Content-Disposition", "attachment; filename=SalesReport.xlsx")
	c.Header("Content-Type", "application/xlsx")
	c.File("./public/SalesReport.xlsx")
}

func DownloadPDF(c *gin.Context) {
	c.Header("Content-Disposition", "attachment; filename=SalesReport.pdf")
	c.Header("Content-Type", "application/pdf")
	c.File("./public/SalesReport.pdf")
}
