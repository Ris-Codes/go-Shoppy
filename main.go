package main

import (
	"html/template"
	"path/filepath"

	"github.com/Ris-Codes/go-Shoppy/config"
	"github.com/Ris-Codes/go-Shoppy/initializer"
	"github.com/Ris-Codes/go-Shoppy/routes"
	"github.com/gin-gonic/gin"
)

func init() {
	initializer.LoadEnv()
	config.ConnectDB()
	R := createCustomRenderer()
	R.Static("/public", "./public")
}

var R = gin.Default()

func main() {
	routes.AdminRouts(R)
	routes.UserRouts(R)

	R.Run()
}

func createCustomRenderer() *gin.Engine {

	// Custom template functions
	funcMap := template.FuncMap{
		"seq": func(n int) []int {
			var seq []int
			for i := 1; i <= n; i++ {
				seq = append(seq, i)
			}
			return seq
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}

	// Load templates
	R.SetFuncMap(funcMap)
	R.LoadHTMLGlob(filepath.Join("templates", "*"))

	return R
}
