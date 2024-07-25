package app

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/ca-media-nantes/libgo/v2/logger"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/terui-ryota/admin/internal/app/admin-web/config"
	"github.com/terui-ryota/admin/internal/app/admin-web/handler"
	//"github.com/ca-media-nantes/pick/backend/admin/internal/app/admin-web/config"
	//"github.com/ca-media-nantes/pick/backend/admin/internal/app/admin-web/handler"
)

type app struct {
	applicationConfig config.Application
	//adminHandler      handler.AdminHandler
	authHandler      handler.AuthHandler
	offerItemHandler handler.OfferItemHandler
	systemHandler    handler.SystemHandler
}

func NewApp(
	applicationConfig config.Application,
	//adminHandler handler.AdminHandler,
	authHandler handler.AuthHandler,
	offerItemHandler handler.OfferItemHandler,
	systemHandler handler.SystemHandler,
) *app {
	return &app{
		applicationConfig: applicationConfig,
		//adminHandler:      adminHandler,
		authHandler:      authHandler,
		offerItemHandler: offerItemHandler,
		systemHandler:    systemHandler,
	}
}

func (a *app) Start() error {
	if a.applicationConfig.GinDebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(static.Serve("/static", static.LocalFile("static", false)))
	fm := sprig.FuncMap()
	fm["JSON"] = func(a any) template.JS {
		j, err := json.Marshal(a)
		if err != nil {
			logger.Default().Infof("json.Marshal: %w", err)
		}
		return template.JS(j) //nolint: gosec
	}
	fm["URL"] = func(s string) template.URL {
		return template.URL(s) //nolint: gosec
	}
	router.SetFuncMap(fm)
	router.Use(func(c *gin.Context) {
		if c.Keys == nil {
			c.Keys = make(map[string]any)
		}
		c.Keys["environment"] = os.Getenv("ENV")
		c.Keys["contextPath"] = a.applicationConfig.ContextPath
		c.Keys["path"] = c.Request.URL.Path
		c.Next()
	})

	files := make([]string, 0)
	if err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".html") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("filepath.Walk: %w", err)
	}
	logger.Default().Debugf("%s", files)
	router.LoadHTMLFiles(files...)

	// 未認証で実行可能なエンドポイント
	router.GET("/live", a.systemHandler.Live)
	//router.GET("/auth/login", a.authHandler.Login)
	//router.GET("/auth/callback", a.authHandler.LoginCallback)
	//router.GET("/auth/logout", a.authHandler.Logout)

	router.GET("/", a.authHandler.Root)

	// github.com/ca-media-nantes/pick/go-lib/pkg/common_metadata での共通項目を取得するために必要です
	// common_metadata 内で auth_handler の saveLoginUserToContext で設定した共通項目を取得できるようになります
	router.ContextWithFallback = true
	router.GET("/offer_item", a.offerItemHandler.ListOfferItemsHTML)
	router.GET("/offer_item/new", a.offerItemHandler.GetOfferItemNewHTML)
	router.GET("/offer_item/:offer_item_id", a.offerItemHandler.GetOfferItemDetailHTML)
	router.GET("/offer_item/:offer_item_id/edit", a.offerItemHandler.GetOfferItemEditHTML)
	router.GET("/offer_item/:offer_item_id/:stage", a.offerItemHandler.GetOfferItemStageDetailHTML)
	//router.GET("/offer_item/:offer_item_id/shipment/zip/:filename", a.offerItemHandler.DownloadShipmentZip)
	router.GET("/offer_item/:offer_item_id/shipment/preview/:filename", a.offerItemHandler.DownloadShipmentPreview)
	router.GET("api/offer_item/search", a.offerItemHandler.SearchOfferItemsJSON)
	router.POST("api/offer_item", a.offerItemHandler.CreateOfferItem)
	router.GET("api/offer_item/assignee_under_examination", a.offerItemHandler.ListAssigneesUnderExaminationJSON)
	router.GET("api/offer_item/:offer_item_id", a.offerItemHandler.GetOfferItem)
	router.PATCH("api/offer_item/:offer_item_id", a.offerItemHandler.UpdateOfferItem)
	router.DELETE("api/offer_item/:offer_item_id", a.offerItemHandler.DeleteOfferItem)
	router.PATCH("api/offer_item/:offer_item_id/invite", a.offerItemHandler.Invite)
	router.GET("api/offer_item/:offer_item_id/assignee", a.offerItemHandler.ListAssigneesJSON)

	router.POST("api/offer_item/:offer_item_id/stage/:stage/send_remind_mail", a.offerItemHandler.SendRemindMail)

	router.POST("api/offer_item/:offer_item_id/save_lottery_results", a.offerItemHandler.SaveLotteryResults)
	router.POST("api/offer_item/:offer_item_id/save_preexamination_results", a.offerItemHandler.SavePreExaminationResults)
	router.POST("api/offer_item/:offer_item_id/save_examination_results", a.offerItemHandler.SaveExaminationResults)
	router.POST("api/offer_item/:offer_item_id/finish_shipment", a.offerItemHandler.FinishShipment)
	router.POST("api/offer_item/:offer_item_id/save_payment_results", a.offerItemHandler.SavePaymentResults)

	router.PATCH("/offer_item/:offer_item_id/close", a.offerItemHandler.Close)

	//// 認証必須のエンドポイント
	//requiredAuth := router.Group("/").Use(a.authHandler.Authorize, handler.CheckPermission)
	//{
	//	requiredAuth.GET("/admin", a.adminHandler.IndexHTML)
	//	requiredAuth.GET("/admin/users", a.adminHandler.IndexUserHTML)
	//	requiredAuth.GET("/admin/users/:name", a.adminHandler.GetUserHTML)
	//
	//	requiredAuth.GET("/offer_item", a.offerItemHandler.ListOfferItemsHTML)
	//	requiredAuth.GET("/offer_item/new", a.offerItemHandler.GetOfferItemNewHTML)
	//	requiredAuth.GET("/offer_item/:offer_item_id", a.offerItemHandler.GetOfferItemDetailHTML)
	//	requiredAuth.GET("/offer_item/:offer_item_id/edit", a.offerItemHandler.GetOfferItemEditHTML)
	//	requiredAuth.GET("/offer_item/:offer_item_id/:stage", a.offerItemHandler.GetOfferItemStageDetailHTML)
	//	requiredAuth.GET("/offer_item/:offer_item_id/shipment/zip/:filename", a.offerItemHandler.DownloadShipmentZip)
	//	requiredAuth.GET("/offer_item/:offer_item_id/shipment/preview/:filename", a.offerItemHandler.DownloadShipmentPreview)
	//}
	//
	//// 認証必須の rest api のエンドポイント
	//apiRequiredAuth := router.Group("/api").Use(a.authHandler.Authorize, handler.CheckPermission)
	//{
	//	apiRequiredAuth.GET("/admin/users", a.adminHandler.ListJSON)
	//	apiRequiredAuth.POST("/admin/users", a.adminHandler.CreateUser)
	//	apiRequiredAuth.PUT("/admin/users/:name", a.adminHandler.UpdateUser)
	//	apiRequiredAuth.DELETE("/admin/users/:name", a.adminHandler.DeleteUser)
	//
	//	apiRequiredAuth.GET("/offer_item/search", a.offerItemHandler.SearchOfferItemsJSON)
	//	apiRequiredAuth.POST("/offer_item", a.offerItemHandler.CreateOfferItem)
	//	apiRequiredAuth.GET("/offer_item/assignee_under_examination", a.offerItemHandler.ListAssigneesUnderExaminationJSON)
	//	apiRequiredAuth.GET("/offer_item/:offer_item_id", a.offerItemHandler.GetOfferItem)
	//	apiRequiredAuth.PATCH("/offer_item/:offer_item_id", a.offerItemHandler.UpdateOfferItem)
	//	apiRequiredAuth.DELETE("/offer_item/:offer_item_id", a.offerItemHandler.DeleteOfferItem)
	//	apiRequiredAuth.PATCH("/offer_item/:offer_item_id/invite", a.offerItemHandler.Invite)
	//	apiRequiredAuth.GET("/offer_item/:offer_item_id/assignee", a.offerItemHandler.ListAssigneesJSON)
	//
	//	apiRequiredAuth.POST("/offer_item/:offer_item_id/stage/:stage/send_remind_mail", a.offerItemHandler.SendRemindMail)
	//
	//	apiRequiredAuth.POST("/offer_item/:offer_item_id/save_lottery_results", a.offerItemHandler.SaveLotteryResults)
	//	apiRequiredAuth.POST("/offer_item/:offer_item_id/save_preexamination_results", a.offerItemHandler.SavePreExaminationResults)
	//	apiRequiredAuth.POST("/offer_item/:offer_item_id/save_examination_results", a.offerItemHandler.SaveExaminationResults)
	//	apiRequiredAuth.POST("/offer_item/:offer_item_id/finish_shipment", a.offerItemHandler.FinishShipment)
	//	apiRequiredAuth.POST("/offer_item/:offer_item_id/save_payment_results", a.offerItemHandler.SavePaymentResults)
	//
	//	apiRequiredAuth.PATCH("/offer_item/:offer_item_id/close", a.offerItemHandler.Close)
	//
	//}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.applicationConfig.Port),
		Handler:           router,
		ReadHeaderTimeout: 1 * time.Minute,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
	}

	go func() {
		// サービスの接続
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Default().Errorf("listen: %s\n", err)
		}
	}()

	go a.serveMetrics()

	// シグナル割り込みを待ち、タイムアウト時間が5秒の graceful shutdown をする
	quit := make(chan os.Signal, 1)
	defer close(quit)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Default().Errorf("Server Shutdown:", err)
	}
	// ctx.Done() をキャッチする。5秒間のタイムアウト。
	<-ctx.Done()
	logger.Default().Infof("timeout of 5 seconds.")
	return nil
}

// serveMetrics メトリクス取得用のアプリケーションを配置します。
func (a *app) serveMetrics() {
	h := promhttp.Handler()
	http.Handle("/metrics", h)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.applicationConfig.MonitorPort),
		Handler:           h,
		ReadHeaderTimeout: 1 * time.Minute,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Default().Errorf("listen: %s\n", err)
	}
}
