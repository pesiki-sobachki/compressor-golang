package config

type HTTP struct {
	Address string `conf:"http_address" env:"HTTP_ADDRESS" default:":8080"`
}

type Storage struct {
	Path string `conf:"storage_path" env:"STORAGE_PATH" default:"./storage"`
}

type Log struct {
	Level string `conf:"log_level" env:"LOG_LEVEL" default:"info"`
}

type Config struct {
	HTTP    HTTP
	Storage Storage
	Log     Log
}
