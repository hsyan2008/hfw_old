package hfw

var ENVIRONMENT string

var Config struct {
	Server   Server
	Logger   Logger
	Db       Db
	Cache    Cache
	Template Template
	Route    Route
	Session  struct {
		SessID     string
		Server     string
		Prefix     string
		Expiration int32
		Db         int
		Password   string
	}
}

type Server struct {
	Port         string
	ReadTimeout  int64
	WriteTimeout int64
}

type Logger struct {
	LogGoID   bool
	LogFile   string
	LogLevel  string
	Console   bool
	LogType   string
	LogMaxNum int32
	LogSize   int64
	LogUnit   string
}

type Db struct {
	Driver   string
	Username string
	Password string
	Protocol string
	Address  string
	Dbname   string
	Params   string
	Cache    string
}

type Cache struct {
	Servers []string
	Config  CacheConfig
}

type CacheConfig struct {
	Prefix     string
	Expiration int32
}
type Template struct {
	Static string
	Html   string
	Cache  bool
}
type Route struct {
	Controller string
	Action     string
}
