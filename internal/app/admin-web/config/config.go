package config

import (
	"go.uber.org/zap"
)

type AdminWebConfig struct {
	Application         Application `yaml:"application"`
	Account             Account     `yaml:"account"`
	GRPCServers         GRPCServers `yaml:"grpc_servers"`
	AmebaBlog           AmebaBlog
	PermanOpenIDConnect PermanOpenIDConnect `yaml:"perman_open_id_connect"`
	Databases           Databases           `yaml:"databases"`
	//RedisConfig         definitions.BasicRedisConfig `yaml:"redis"`
	//TracingConfig       tracing.Config               `yaml:"tracing"`
	ZapConfig         zap.Config         `yaml:"zap"`
	ZipPasswordConfig *ZipPasswordConfig `yaml:"zip_password_config"`
}

type Application struct {
	Port         uint   `yaml:"port"`
	MonitorPort  uint   `yaml:"monitor_port"`
	GinDebugMode bool   `yaml:"gin_debug_mode"`
	ContextPath  string `yaml:"context_path"`
	Domain       string `yaml:"domain"`
}

type Account struct {
	ClientID string `yaml:"client_id"`
}

type AmebaBlog struct {
	APIURL string `yaml:"api_url"`
}

type GRPCServers struct {
	AffiliatorGRPCServer    AffiliatorGRPCServer    `yaml:"affiliator"`
	AffiliateItemGRPCServer AffiliateItemGRPCServer `yaml:"affiliate_item"`
	CommerceGRPCServer      CommerceGRPCServer      `yaml:"commerce"`
	IdmapGRPCServer         IdmapGRPCServer         `yaml:"idmap"`
	MeasurementGRPCServer   MeasurementGRPCServer   `yaml:"measurement"`
	NotificationGRPCServer  NotificationGRPCServer  `yaml:"notification"`
	OfferItemGRPCServer     OfferItemGRPCServer     `yaml:"offer_item"`
	OfferItemV2GRPCServer   OfferItemV2GRPCServer   `yaml:"offer_item_v2"`
	SignupGRPCServer        SignupGRPCServer        `yaml:"signup"`
}

type GRPCServer struct {
	Host string `yaml:"host"`
	Port uint   `yaml:"port"`
}

type ZipPasswordConfig struct {
	ShipmentZipPassoword string `yaml:"shipment_zip_password"`
}

type (
	AffiliateItemGRPCServer GRPCServer
	AffiliatorGRPCServer    GRPCServer
	CommerceGRPCServer      GRPCServer
	IdmapGRPCServer         GRPCServer
	MeasurementGRPCServer   GRPCServer
	NotificationGRPCServer  GRPCServer
	OfferItemGRPCServer     GRPCServer
	OfferItemV2GRPCServer   GRPCServer
	SignupGRPCServer        GRPCServer
)

type PermanOpenIDConnect struct {
	Issuer       string `yaml:"issuer"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL  string `yaml:"redirect_url"`
}

type Databases struct {
	Primary Database `yaml:"primary"`
	Replica Database `yaml:"replica"`
}

type Database struct {
	Name        string `yaml:"dbname"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	MaxIdleConn int    `yaml:"max_idle_conn"`
	MaxOpenConn int    `yaml:"max_open_conn"`
	Debug       bool   `yaml:"debug"`
}
