package hfw

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/cachestore"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/hsyan2008/go-logger/logger"
)

var engine *xorm.Engine

func Init_db() {

	var err error
	dbConfig := Config.Db
	driver := dbConfig.Driver
	dbDsn := fmt.Sprintf("%s:%s@%s(%s)/%s%s",
		dbConfig.Username, dbConfig.Password, dbConfig.Protocol,
		dbConfig.Address, dbConfig.Dbname, dbConfig.Params)

	engine, err = xorm.NewEngine(driver, dbDsn)
	if err != nil {
		logger.Warn(err)
		panic(err)
	}

	if ENVIRONMENT != "prod" {
		engine.ShowSQL(true)
		engine.Logger().SetLevel(core.LOG_DEBUG)
	} else {
		engine.Logger().SetLevel(core.LOG_WARNING)
	}

	//连接池的空闲数大小
	engine.SetMaxIdleConns(100)
	//最大打开连接数
	engine.SetMaxOpenConns(1000)

	go keepalive()

	openCache(dbConfig.Cache)
}

func openCache(cacheType string) {
	var cacher *xorm.LRUCacher

	//开启缓存
	switch cacheType {
	case "memory":
		cacher = xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	case "memcache":
		cacher = xorm.NewLRUCacher(cachestore.NewMemCache(Config.Cache.Servers), 999999999)
	}
	if cacher != nil {
		//可以指定缓存有效时间，如下
		cacher.Expired = 86400 * time.Second
		//所有表开启缓存
		engine.SetDefaultCacher(cacher)
	}
}

//保持mysql连接活跃
func keepalive() {
	for {
		time.Sleep(3600 * time.Second)
		_ = engine.Ping()
	}
}
