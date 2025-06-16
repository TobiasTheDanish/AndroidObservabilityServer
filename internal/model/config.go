package model

type Config struct {
	Port     int `goenv:"OBSERVE_API_PORT,default=8080"`
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Database string `goenv:"OBSERVE_DB_DATABASE,required"`
	Password string `goenv:"OBSERVE_DB_PASSWORD,required"`
	Username string `goenv:"OBSERVE_DB_USERNAME,required"`
	Port     string `goenv:"OBSERVE_DB_PORT,required"`
	Host     string `goenv:"OBSERVE_DB_HOST,required"`
	Schema   string `goenv:"OBSERVE_DB_SCHEMA,required"`
}
