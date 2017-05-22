package hfw

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

type sessRedisStore struct {
	c          redis.Conn
	prefix     string
	expiration int32
}

func NewSessRedisStore() *sessRedisStore {
	cacheConfig := Config.Session
	dialOption := []redis.DialOption{
		redis.DialConnectTimeout(1 * time.Second),
		redis.DialReadTimeout(1 * time.Second),
		redis.DialWriteTimeout(1 * time.Second),
		redis.DialDatabase(cacheConfig.Db),
	}
	if cacheConfig.Password != "" {
		dialOption = append(dialOption, redis.DialPassword(cacheConfig.Password))
	}
	c, err := redis.Dial("tcp", cacheConfig.Server, dialOption...)
	CheckErr(err)
	return &sessRedisStore{
		c:          c,
		prefix:     cacheConfig.Prefix + "sess_",
		expiration: cacheConfig.Expiration,
	}
}

func (s *sessRedisStore) IsExist(sessid, key string) (value bool, err error) {
	// key = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	// Debug("IsExist cache key:", sessid, key)

	value, err = redis.Bool(s.c.Do("HEXISTS", s.prefix+sessid, key))
	if err != nil {
		Warn("IsExist cache key:", sessid, key, err)
	}

	return
}

func (s *sessRedisStore) Put(sessid, key string, value interface{}) (err error) {
	// key = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	// Debug("Put cache key:", sessid, key, value)

	v, err := Gob.Marshal(&value)
	if err != nil {
		Warn("Put cache Gob Marshal key:", sessid, key, value, err)
	} else {
		_, err = s.c.Do("HSET", s.prefix+sessid, key, v)
		if err != nil {
			Warn("Put cache key:", sessid, key, v, err)
		}
	}

	return
}

func (s *sessRedisStore) Get(sessid, key string) (value interface{}, err error) {
	// key = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	// Debug("Get cache key:", sessid, key)

	v, err := s.c.Do("HGET", s.prefix+sessid, key)
	if err != nil {
		Warn("Get cache key:", sessid, key, err)
	} else {
		err = Gob.Unmarshal(v.([]byte), &value)
		if err != nil {
			Warn("Get cache Gob Unmarshal key:", sessid, key, err)
		}
	}

	return
}

func (s *sessRedisStore) Del(sessid, key string) (err error) {
	// key = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	// Debug("Del cache key:", sessid, key)

	_, err = s.c.Do("HDEL", s.prefix+sessid, key)
	if err != nil {
		Warn("Del cache key:", sessid, key, err)
	}

	return
}

func (s *sessRedisStore) Destroy(sessid string) (err error) {
	// key = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	// Debug("Del cache key:", sessid)

	_, err = s.c.Do("DEL", s.prefix+sessid)
	if err != nil {
		Warn("Del cache key:", sessid, err)
	}

	return
}

func (s *sessRedisStore) Rename(sessid, newid string) (err error) {
	// key = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	// Debug("Rename cache key:", sessid, "to key:", newid)

	defer func() {
		_ = s.c.Close()
	}()

	_, err = s.c.Do("RENAME", s.prefix+sessid, s.prefix+newid)
	if err != nil {
		Warn("Rename cache key:", sessid, "to key:", newid, err)
		return
	}
	if s.expiration > 0 {
		_, _ = s.c.Do("EXPIRE", s.prefix+newid, s.expiration)
	}

	return
}
