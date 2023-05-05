package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	go SerialBackend()
	router := gin.Default()
	router.GET("/slcan/:id", getSlcanFrame)
	router.POST("/slcan", postSlcanFrame)
	router.PUT("/slcan/filter/:id", putSlcanFrame)
	router.DELETE("/slcan/filter/:id", delSlcanFrame)

	router.Run("localhost:8080")
}

func getSlcanFrame(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"slcan-svc": "illegal frame id"})
		return
	}
	f, err := backendDB.ReadFromSlcanDB(uint32(id))
	if err == nil {
		c.IndentedJSON(http.StatusOK, f)
	} else {
		c.IndentedJSON(http.StatusNoContent, gin.H{"slcan-svc": "requested frame not found"})
	}

	return
}

func postSlcanFrame(c *gin.Context) {
	var f slcanFrame

	if err := c.BindJSON(&f); err != nil {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"slcan-svc": "illegal frame data"})
		return
	}

	if err := backendDB.WriteToSerialBackend(f); err == nil {
		c.IndentedJSON(http.StatusCreated, f)
	} else {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"slcan-svc": "frame posted failed"})
	}

	return
}

func putSlcanFrame(c *gin.Context) {
	var f slcanFrame
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"slcan-svc": "illegal frame id"})
		return
	}

	if err := c.BindJSON(&f); err != nil {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"slcan-svc": "illegal frame data"})
		return
	}

	if uint32(id) != f.ID {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"slcan-svc": "url id & frame id mismatch"})
		return
	}

	err = backendDB.UpdateToSlcanDB(f)
	if err == nil {
		c.IndentedJSON(http.StatusOK, f)
	} else {
		c.IndentedJSON(http.StatusNoContent, gin.H{"slcan-svc": "requested frame not found"})
	}
}

func delSlcanFrame(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"slcan-svc": "illegal frame id"})
		return
	}
	err = backendDB.RemoveFromSlcanDB(uint32(id))
	if err == nil {
		c.IndentedJSON(http.StatusOK, nil)
	} else {
		c.IndentedJSON(http.StatusNoContent, gin.H{"slcan-svc": "requested frame not found"})
	}
}
