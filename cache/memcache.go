package cache

import (
	"hfw"

	"github.com/bradfitz/gomemcache/memcache"
)

type CacheStore struct {
	mc            *memcache.Client
	mc_prefix     string
	mc_expiration int32
}

func NewCacheStore() *CacheStore {
	cacheConfig := hfw.Config.Cache
	return &CacheStore{
		mc:            memcache.New(cacheConfig.Servers...),
		mc_prefix:     cacheConfig.Config.Prefix,
		mc_expiration: cacheConfig.Config.Expiration,
	}
}

func (s *CacheStore) Put(key string, value interface{}) (err error) {
	// key = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	Debug("Put cache key:", key)

	//注意必须有&号
	v, err := Gob.Marshal(&value)
	if err == nil {
		err = s.mc.Set(&memcache.Item{Key: s.mc_prefix + key, Value: v})
		if err != nil {
			Warn("Put cache key:", key, err)
		}
	} else {
		Warn("Put cache key:", key, err)
	}

	return
}

func (s *CacheStore) Get(key string) (value interface{}, err error) {
	// key = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	Debug("Get cache key:", key)

	it, err := s.mc.Get(s.mc_prefix + key)
	if err == nil {
		err = Gob.Unmarshal(it.Value, &value)
		if err != nil {
			Warn("Get cache key:", key, err)
		}
	} else {
		Warn("Get cache key:", key, err)
	}

	return
}

func (s *CacheStore) Del(key string) (err error) {
	// key = fmt.Sprintf("%x", md5.Sum([]byte(key)))
	Debug("Del cache key:", key)

	err = s.mc.Delete(s.mc_prefix + key)
	if err != nil {
		Warn("Del cache key:", key, err)
	}

	return
}

func (s *CacheStore) GetMulti() {
	// s.mc.GetMulti()
}
