package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math/bits"
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

func getLink(c *gin.Context) {
	renderHTML := acceptsHTML(c)
	shortURL := c.Param("id")
	for _, link := range links {
		if link.ShortURL == shortURL {
			if renderHTML {
				c.Redirect(http.StatusFound, link.LongURL)
			} else {
				c.IndentedJSON(http.StatusOK, link.LongURL)
			}
			return
		}
	}
	if renderHTML {
		c.HTML(http.StatusNotFound, "missing.tmpl", nil)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "link not found"})
	}
}

func shortenLink(c *gin.Context) {
	longURL := c.Query("longURL")
	fmt.Println(longURL)
	urlHash := ""
	for _, link := range links {
		if link.LongURL == longURL {
			urlHash = link.ShortURL
		}
	}

	if urlHash == "" {
		id := links[len(links)-1].ID + 1
		urlHash = fmt.Sprintf("%x", getHash(id))
		newLink := Link{id, longURL, urlHash}
		links = append(links, newLink)
	}

	shortURL := fmt.Sprintf("%v/%v", ServerDomain, urlHash)
	fmt.Println(shortURL)
	if acceptsHTML(c) {
		c.HTML(http.StatusOK, "shorten.tmpl", gin.H{"shortURL": shortURL})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"shortURL": shortURL})
	}
}

func getHash(id uint32) uint32 {
	reversedPoly := bits.Reverse32(0xF4ACFB13)
	t := crc32.MakeTable(reversedPoly)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b, id)
	return crc32.Checksum(b, t)
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", getIndex)
	router.GET("/:id", getLink)
	router.GET("/shorten", shortenLink)

	router.Run(ServerDomain)
}
