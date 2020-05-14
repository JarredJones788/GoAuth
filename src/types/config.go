package types

//MySQLConfig - mysql connection info
type MySQLConfig struct {
	Conn     string
	UserName string
	Password string
	DBName   string
}

//RedisConfig - redis connection info
type RedisConfig struct {
	Conn    string
	Timeout int
	Active  bool
}

//EmailConfig - email connection info
type EmailConfig struct {
	Email       string
	Password    string
	SMTPAddress string
	SMTPPort    int
}

//Config - runtime config
type Config struct {
	MySQL       MySQLConfig
	Redis       RedisConfig
	Email       EmailConfig
	ServerPort  string
	Host        string
	LogDuration float64
}
