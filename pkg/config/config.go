package config

type Config struct {
	Database struct {
		ConnectionString   string
		MaxIdleConnections int
		MaxOpenConnections int
	}
	Server struct {
		Address string
	}
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		config = &Config{
			Database: struct {
				ConnectionString   string
				MaxIdleConnections int
				MaxOpenConnections int
			}{
				ConnectionString:   "./database/database.db",
				MaxIdleConnections: 3,
				MaxOpenConnections: 100,
			},
			Server: struct{ Address string }{
				Address: "127.0.0.1:8000",
			},
		}
	}

	return config
}
