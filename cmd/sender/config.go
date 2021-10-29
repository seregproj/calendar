package main

type Config struct {
	Logger
	MessageBroker
}

type Logger struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	File  string `yaml:"file" env:"LOG_FILE"`
}

type MessageBroker struct {
	RBMQ
}

type RBMQ struct {
	DSN       string `yaml:"dsn" env:"RBMQ_DSN"`
	QueueName string `yaml:"queueName" env:"RBMQ_QUEUE_NAME"`
}

func NewConfig() Config {
	return Config{}
}
