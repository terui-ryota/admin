package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (h *SystemHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, "==============ok==============")
}
