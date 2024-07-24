package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terui-ryota/admin/internal/app/admin-web/config"
)

const loginUserKey = "loginUser"

type LoginUserKey struct{}

type AuthHandler struct {
	applicationConfig config.Application
}

func NewAuthHandler(
	applicationConfig config.Application,
) *AuthHandler {
	return &AuthHandler{
		applicationConfig: applicationConfig,
	}
}

func (h *AuthHandler) Root(c *gin.Context) {
	data := c.Keys
	//if err != nil {
	//	if errors.Is(err, usecase.ErrUnAuthorized) {
	//		c.HTML(http.StatusOK, "public/index.html", data)
	//		return
	//	}
	//	handle5xxHTML(c, err)
	//	return
	//}
	//saveLoginUserToContext(c, *user)
	c.HTML(http.StatusOK, "public/index.html", data)
}
