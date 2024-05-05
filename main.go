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

// TODO: improve error handling, add template for 5xx error

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
	longURL := url.QueryEscape(rawURL)
	c.Redirect(http.StatusFound, fmt.Sprintf("/shorten?longURL=%s", longURL))
}

func getLink(c *gin.Context) {
	renderHTML := acceptsHTML(c)
	urlCode := c.Param("id")

	var link Link
	res := db.Where("url_code = ?", urlCode).First(&link)

	if res.Error == nil {
		link.TimesAccessed += 1
		link.LastAccessed = time.Now()

		res := db.Save(&link)
		if res.Error != nil {
			if acceptsHTML(c) {
				c.Status(http.StatusInternalServerError)
			} else {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "internal error, unable to access link information"})
			}
			return
		}

		if renderHTML {
			c.Redirect(http.StatusFound, link.LongURL)
		} else {
			c.IndentedJSON(http.StatusOK, gin.H{"longURL": link.LongURL})
		}
	} else {
		if renderHTML {
			c.HTML(http.StatusNotFound, "missing.tmpl", nil)
		} else {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "link not found"})
		}
	}
}

func shortenLink(c *gin.Context) {
	longURL := c.Query("longURL")
	curr_time := time.Now()

	var link Link
	res := db.Where("long_url = ?", longURL).First(&link)

	if res.Error != nil {
		link = Link{LongURL: longURL, CreatedOn: curr_time, LastAccessed: curr_time, TimesAccessed: 1}

		res := db.Create(&link)
		if res.Error != nil {
			if acceptsHTML(c) {
				c.Status(http.StatusInternalServerError)
			} else {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "internal error, unable to create short link"})
			}
			return
		}

		link.URLCode = fmt.Sprintf("%x", getHash(link.ID))
		res = db.Save(&link)
		if res.Error != nil {
			if acceptsHTML(c) {
				c.Status(http.StatusInternalServerError)
			} else {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "internal error, unable to update short link"})
			}
			return
		}
	}

	shortURL := fmt.Sprintf("%v/%v", ServerDomain, link.URLCode)
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
	router.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.tmpl", nil)
	})
	router.GET("/api", func(c *gin.Context) {
		c.HTML(http.StatusOK, "api.tmpl", nil)
	})
	router.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact.tmpl", nil)
	})

	router.Run(ServerDomain)
}
