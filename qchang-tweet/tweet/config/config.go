package config

type Config struct {
	ConnectionString string
	AuthSecretKey    string
}

func New() *Config {
	return &Config{
		ConnectionString: "postgres://admin:123456@localhost:5432/tweet?sslmode=disable",
		AuthSecretKey:    "1498b5467a63dffa2dc9d9e069caf075d16fc33fdd4c3b01bfadae6433767d93",
	}
}
