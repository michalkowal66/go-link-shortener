package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"math/bits"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const ServerDomain = "localhost:8080"

var db *gorm.DB = nil

type Link struct {
	ID            uint32    `json:"id"`
	LongURL       string    `json:"longURL"`
	URLCode       string    `json:"urlCode"`
	CreatedOn     time.Time `json:"createdOn"`
	LastAccessed  time.Time `json:"lastAccessed"`
	TimesAccessed uint32    `json:"timesAccessed"`
}

func connectDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%v:%v@tcp(localhost:3306)/go_link_shortener?charset=utf8mb4&parseTime=True&loc=Local", os.Getenv("DBUSER"), os.Getenv("DBPASS"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
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

func postIndex(c *gin.Context) {
	rawURL := c.PostForm("rawURL")
	fmt.Println(rawURL)
	longURL := url.QueryEscape(rawURL)
	c.Redirect(http.StatusFound, fmt.Sprintf("/shorten?longURL=%s", longURL))
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
	var err error
	db, err = connectDB()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", getIndex)
	router.POST("/", postIndex)
	router.GET("/:id", getLink)
	router.GET("/shorten", shortenLink)

	router.Run(ServerDomain)
}
