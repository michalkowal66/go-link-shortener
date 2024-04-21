package main

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

const ServerDomain = "localhost:8080"

type Link struct {
	ID       uint32 `json:"id"`
	LongURL  string `json:"longURL"`
	ShortURL string `json:"shortURL"`
}

var links = []Link{
	{1, "https://www.youtube.com/watch?v=7jMlFXouPk8", "2af5e8c4"},
}

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
