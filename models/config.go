package models

type CoworkingConfig struct {
	Dsn            string `json:"dsn"`
	SecretKey      string `json:"secret_key"`
	AllowedOrigins string `json:"allowed_origins"`
}
