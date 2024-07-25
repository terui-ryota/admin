package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terui-ryota/admin/pkg/logger"
)

func handle400JSON(c *gin.Context, err error) {
	if err != nil {
		logger.FromContext(c).Errorf("err: %w", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusBadRequest, "")
}

func handle5xxJSON(c *gin.Context, err error) {
	if err != nil {
		logger.FromContext(c).Errorf("err: %w", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusInternalServerError, "")
}

func handle400HTML(c *gin.Context, err error) {
	if err != nil {
		logger.FromContext(c).Errorf("err: %w", err)
	}
	data := c.Keys
	data["error"] = err
	c.HTML(http.StatusBadRequest, "error/4xx.html", data)
}

func handle404HTML(c *gin.Context, err error) {
	if err != nil {
		logger.FromContext(c).Errorf("err: %w", err)
	}
	data := c.Keys
	data["error"] = err
	c.HTML(http.StatusNotFound, "error/4xx.html", data)
}

func handle5xxHTML(c *gin.Context, err error) {
	if err != nil {
		logger.FromContext(c).Errorf("err: %w", err)
	}
	data := c.Keys
	data["error"] = err
	c.HTML(http.StatusInternalServerError, "error/5xx.html", data)
}
