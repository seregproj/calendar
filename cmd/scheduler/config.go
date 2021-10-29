package main

type Config struct {
	Logger
	Storage
	MessageBroker
	App
}

type Logger struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	File  string `yaml:"file" env:"LOG_FILE"`
}

type Storage struct {
	Type  string `yaml:"type" env:"STORAGE_TYPE" env-default:"memory"`
	PGSQL PGSQL
}

type PGSQL struct {
	DSN string `yaml:"dsn" env:"PGSQL_DSN"`
}

type MessageBroker struct {
	RBMQ
}

type RBMQ struct {
	DSN       string `yaml:"dsn" env:"RBMQ_DSN"`
	QueueName string `yaml:"queueName" env:"RBMQ_QUEUE_NAME"`
}

type App struct {
	Notifications
}

type Notifications struct {
	Limit int64 `yaml:"limit" env:"APP_NOTIFICATIONS_LIMIT"`
}

func NewConfig() Config {
	return Config{}
}
