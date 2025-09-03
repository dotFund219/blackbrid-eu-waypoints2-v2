package rabbitmq

type Config struct {
	URL       string `yaml:"url"`
	QueueName string `yaml:"queuename"`
	Durable   bool   `yaml:"durable"`
	AutoAck   bool   `yaml:"autoack"`
}
