package redis

import (
	"testing"
	"time"

	"github.com/go-redis/redis"
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

	key := "tpredis"
	if err := c.Watch(func(tx *redis.Tx) error {
		n, err := tx.Get(key).Int64()
		if err != nil && err == redis.Nil {
			t.Errorf("err1-> %v", err)
			return err
		} else if err != nil && err != redis.Nil {
			t.Errorf("err2-> %v", err)
			return err
		}
		t.Logf("n-> %d", n)

		t.Logf("Start sleep.")
		time.Sleep(time.Duration(5) * time.Second)
		// 在redis客户端修改值，下面语句报错 redis: transaction failed

		_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
			// pipe handles the error case
			pipe.DecrBy(key, 2)
			return nil
		})
		return err
	}, key); err != nil {
		t.Errorf("err4-> %v", err)
	}
}
