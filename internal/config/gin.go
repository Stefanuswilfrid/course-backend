package config

//go get github.com/gin-gonic/gin
import "github.com/gin-gonic/gin"

func NewGin() *gin.Engine {
	return gin.Default()
}
