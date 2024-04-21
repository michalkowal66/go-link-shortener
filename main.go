package main

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

const ServerDomain = "localhost:8080"

func acceptsHTML(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	acceptedMimes := strings.Split(accept, ",")
	return slices.Contains(acceptedMimes, "text/html")
}

func getIndex(c *gin.Context) {
	if acceptsHTML(c) {
		c.HTML(http.StatusOK, "index.tmpl", nil)
	} else {
		c.IndentedJSON(http.StatusNotAcceptable, nil)
	}
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", getIndex)

	router.Run(ServerDomain)
}
