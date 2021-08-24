package redis

import (
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
)

func Test(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
    // redis://user:pass@redis_host:port/db
	redisURL := "redis://" + s.Addr()
	t.Logf(redisURL)
	const prefix = "test:prefix:"
	helper, err := NewHelper(redisURL, "testPool", nil, NewOptionsWithDefaultCodec(
		Prefix(prefix), Expiration(10 * time.Minute)))
	if err != nil {
		t.Fatal(err)
	}

	/***** ping *****/
	if err := helper.Ping(); err != nil {
		t.Fatal(err)
	}
	key := "key"
	dataToSave := map[string]int {
		"a": 1,
		"b": 2,
	}

	/***** save *****/
	if err := helper.Save(key, dataToSave, 10 * time.Minute); err != nil {
		t.Fatal(err)
	}

	/***** get *****/
	var dataToGet map[string]int
	if err := helper.Get(key, &dataToGet); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dataToSave, dataToGet) {
		t.Fatal("dataToGet from redis is not equal to dataToSave")
	}
	t.Logf("dataToGet: %v", dataToGet)

	/***** delete *****/
	if err := helper.Delete(key); err != nil {
		t.Fatal(err)
	}

	/***** get after delete *****/
	dataToGet = nil
	err = helper.Get(key, &dataToGet)
	if err != ErrNotFound {
		t.Fatal("key is still in redis after deleted")
	}

	/***** save with a 3s expiration *****/
	if err := helper.Save(key, dataToSave, 3 * time.Second); err != nil {
		t.Fatal(err)
	}
	// FastForward 3s, let the key in redis expired.
	// NOTE: DO NOT USE time.Sleep(). It not work. Ref: https://github.com/alicebob/miniredis/issues/149
	s.FastForward(3 * time.Second)
	err = helper.Get(key, &dataToGet)
	if err != ErrNotFound {
		t.Fatal("key is still in redis after expired")
	}

	/***** test prefix *****/
	// 1. save
	if err := helper.Save(key, dataToSave, 3 * time.Minute); err != nil {
		t.Fatal(err)
	}
	// 2. use redis command to get by "prefix + key"
	b, err := redis.Bytes(s.Get(prefix + key))
	if err != nil {
		t.Fatal(err)
	}
	// 3. Decode
	if err := DefaultCodec.Decode(b, &dataToGet); err != nil {
		t.Fatal(err)
	}
	// 4. isEqual
	if !reflect.DeepEqual(dataToSave, dataToGet) {
		t.Fatal("dataToGet from redis is not equal to dataToSave")
	}
}