# mysql

A mysql ORM(Object Role Modeling) package with redis cache.

## Example

```go
import (
    "encoding/json"
    "testing"
    "time"

	"github.com/xiaoenai/tp-micro/v6/model/mysql"
    "github.com/xiaoenai/tp-micro/v6/model/redis"
)

type testTable struct {
    TestId      int64
    TestContent string
    Deleted     bool `db:"test_deleted"`
}

func (t *testTable) TableName() string {
    return "test_table"
}

func TestCacheDb(t *testing.T) {
    dbConf, err := mysql.ReadConfig("test_db")
    if err != nil {
        t.Fatal(err)
    }
    dbConf.Database = "test"
    redisConf, err := redis.ReadConfig("test_redis")
    if err != nil {
        t.Fatal(err)
    }
    db, err := mysql.Connect(dbConf, redisConf)
    if err != nil {
        t.Fatal(err)
    }
    _, err = db.Exec("CREATE TABLE IF NOT EXISTS `dbtest` (`test_id` INT(10) AUTO_INCREMENT, `test_content` VARCHAR(20), `test_deleted` TINYINT(2),  PRIMARY KEY(`test_id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='测试表'")
    if err != nil {
        t.Fatal(err)
    }
    c, err := db.RegCacheableDB(new(testTable), time.Second)
    if err != nil {
        t.Fatal(err)
    }
    obj := &testTable{
        TestId:      1,
        TestContent: "abc",
        Deleted:     false,
    }
    _, err = c.NamedExec("INSERT INTO dbtest (test_id,test_content,test_deleted)VALUES(:test_id,:test_content,:test_deleted) ON DUPLICATE KEY UPDATE test_id=:test_id", obj)
    if err != nil {
        t.Fatal(err)
    }

    // query and cache
    dest := &testTable{TestId: 1}
    err = c.CacheGet(dest)
    if err != nil {
        t.Fatal(err)
    }
    t.Logf("first:%#v", dest)

    // verify the cache
    cacheKey, _, err := c.CreateCacheKey(dest)
    t.Logf("cacheKey:%#v", cacheKey)
    key := cacheKey.Key
    if err != nil {
        t.Fatal(err)
    }
    b, err := c.Cache.Get(key).Bytes()
    if err != nil {
        t.Fatal(err)
    }
    var v1 = new(testTable)
    err = json.Unmarshal(b, v1)
    if err != nil {
        t.Fatal(err)
    }
    t.Logf("cache of before expiring: %+v", v1)

    time.Sleep(2 * time.Second)

    b, err = c.Cache.Get(key).Bytes()
    if err == nil {
        var v2 = new(testTable)
        err = json.Unmarshal(b, v2)
        if err != nil {
            t.Fatal(err)
        }
        t.Fatalf("expired but not deleted: %#v", v2)
    }
    t.Logf("expired cache error: %v", err)
}
```

```sh
go test -v -run=TestCacheDb
```