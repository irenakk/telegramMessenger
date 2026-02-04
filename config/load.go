package config

func LoadConfig() Config {
	var c Config
	c.TelegramToken = ""
	c.UserServiceURL = "http://localhost:8082/api/v1"
	return c
}
