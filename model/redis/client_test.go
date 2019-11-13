package redis

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
)

func TestClient(t *testing.T) {
	cfg, err := ReadConfig("test_redis")
	if err != nil {
		t.Fatal("ReadConfig(\"test_redis\")", err)
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal("NewClient(\"test_redis\")", err)
	}

	m := NewModule("test")

	s, err := c.Set(m.Key("a_key"), "a_value", time.Second).Result()
	if err != nil {
		t.Fatal("c.Set().Result() error:", err)
	}
	t.Logf("c.Set().Result() result: %s", s)

	s, err = c.Get(m.Key("a_key")).Result()
	if err != nil {
		t.Fatal("c.Get().Result() error:", err)
	}
	t.Logf("c.Get().Result() result: %s", s)
	time.Sleep(2 * time.Second)

	s, err = c.Get(m.Key("a_key")).Result()
	if err == nil {
		t.Fatalf("[after 2s] c.Get().Result() result: %s", s)
	}
	t.Logf("[after 2s] c.Get().Result() is null ?: %v", err == redis.Nil)
}
