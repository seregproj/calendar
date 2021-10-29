package main

type Config struct {
	Logger  Logger
	Server  ServerConf
	Storage Storage
}

type Logger struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	File  string `yaml:"file" env:"LOG_FILE"`
}

type ServerConf struct {
	GRPC GRPC
	HTTP HTTP
}

type HTTP struct {
	Host string `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port string `yaml:"port" env:"HTTP_PORT" env-default:"8888"`
}

type GRPC struct {
	Host string `yaml:"host" env:"GRPC_HOST" env-default:"0.0.0.0"`
	Port string `yaml:"port" env:"GRPC_PORT" env-default:"8889"`
}

type Storage struct {
	Type  string `yaml:"type" env:"STORAGE_TYPE" env-default:"memory"`
	PGSQL PGSQL
}

type PGSQL struct {
	DSN string `yaml:"dsn" env:"PGSQL_DSN"`
}

func NewConfig() Config {
	return Config{}
}
