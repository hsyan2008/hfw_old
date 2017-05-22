package hfw

import (
	"sync"

	"github.com/pborman/uuid"
)

type sessionStore interface {
	Put(string, string, interface{}) error
	Get(string, string) (interface{}, error)
	IsExist(string, string) (bool, error)
	Del(string, string) error
	Destroy(string) error
	Rename(string, string) error
}

type Session struct {
	id    string
	newid string
	sh    sessionStore
}

var sessPool = sync.Pool{
	New: func() interface{} {
		return new(Session)
	},
}

func NewSession(c *Controller) *Session {
	// s := sessPool.Get().(*Session)
	s := new(Session)
	s.sh = NewSessRedisStore()
	s.newid = uuid.New()

	cookie, err := c.Request.Cookie(Config.Session.SessID)
	if err == nil {
		s.id = cookie.Value
	}

	return s
}

func (s *Session) IsExist(k string) bool {
	v, _ := s.sh.IsExist(s.id, k)
	return v
}

func (s *Session) Set(k string, v interface{}) {
	_ = s.sh.Put(s.id, k, v)
}

func (s *Session) Get(k string) interface{} {
	v, _ := s.sh.Get(s.id, k)
	return v
}

func (s *Session) Del(k string) {
	_ = s.sh.Del(s.id, k)
}

func (s *Session) Destroy() {
	_ = s.sh.Destroy(s.id)
}

func (s *Session) Rename() {
	_ = s.sh.Rename(s.id, s.newid)
}
