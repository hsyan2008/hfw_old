package hfw

import (
	"flag"
	_ "net/http/pprof"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/hsyan2008/go-logger/logger"
)

func init() {
	loadConfig()
	setLog()
	Init_db()
}
func Run() {
	startServe()
}

func loadConfig() {

	flag.StringVar(&ENVIRONMENT, "e", "dev", "set env,e.g dev test prod")
	flag.Parse()

	switch ENVIRONMENT {
	case "dev":
		fallthrough
	case "test":
		fallthrough
	case "prod":
		_, err := toml.DecodeFile(filepath.Join("config", ENVIRONMENT, "config.toml"), &Config)
		CheckErr(err)
	default:
		panic("未定义的环境")
	}

	//存为toml格式
	// var firstBuffer bytes.Buffer
	// e := toml.NewEncoder(&firstBuffer)
	// _ = e.Encode(Config)
	// _ = ioutil.WriteFile("config/"+ENVIRONMENT+"/config.toml", firstBuffer.Bytes(), 0644)

	//设置默认路由
	if Config.Route.Controller == "" {
		Config.Route.Controller = "index"
	}
	if Config.Route.Action == "" {
		Config.Route.Action = "index"
	}
}

//初始化log写入文件
func setLog() {
	lc := Config.Logger
	logger.SetLevelStr(lc.LogLevel)
	logger.SetConsole(lc.Console)
	logger.SetLogGoID(lc.LogGoID)

	if lc.LogFile != "" {
		if lc.LogType == "daily" {
			logger.SetRollingDaily(lc.LogFile)
		} else if lc.LogType == "roll" {
			logger.SetRollingFile(lc.LogFile, lc.LogMaxNum, lc.LogSize, lc.LogUnit)
		} else {
			logger.Warn("请设置log存储方式")
		}
	} else {
		logger.Warn("没有设置log目录和文件")
	}
}
