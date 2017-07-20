package api

// based on -  github.com/donnpebe/go-redis-timeseries
import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
)

// ErrNotFound is returned when the store cannot find the key specified
var ErrNotFound = errors.New("record not found")

// timeSeries is use to save time series data to redis
type timeSeries struct {
	prefix     string
	timestep   time.Duration
	expiration time.Duration
	pool       *redis.Pool
}

// NewTimeSeries create new timeSeries
func NewTimeSeries(prefix string, timestep time.Duration, exp time.Duration, pool *redis.Pool) *timeSeries {
	return &timeSeries{
		prefix:     prefix,
		timestep:   timestep,
		expiration: exp,
		pool:       pool,
	}
}

// Add data to timeseries db
func (t *timeSeries) Add(data interface{}, tm time.Time) (err error) {

	var dataBytes []byte

	if dataBytes, err = json.Marshal(data); err != nil {
		return
	}
	conn := t.pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	conn.Send("ZADD", t.key(tm), tm.UnixNano(), dataBytes)

	if t.expiration > 0 {
		sc := redis.NewScript(2, "local ex = redis.pcall('zcard', KEYS[1]) \n if ex == 1 then return redis.call('expire', KEYS[1], KEYS[2]) end")
		sc.Send(conn, t.key(tm), int64(t.expiration.Seconds()))
	}

	_, err = conn.Do("EXEC")

	return
}

// FetchRange fetch data from the begin time to end time
func (t *timeSeries) FetchRange(begin, end time.Time, dest interface{}) (err error) {
	if begin.After(end) {
		return errors.New("Begin time value must be less than end time value")
	}

	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr || d.IsNil() {
		return errors.New("Fetch value must be non-nil pointer")
	}
	d = d.Elem()
	if d.Kind() != reflect.Slice {
		return errors.New("Fetch value must be pointer to slice")
	}

	typ := d.Type().Elem()

	isPtr := false
	if typ.Kind() == reflect.Ptr {
		isPtr = true
		typ = typ.Elem()
	}

	tmb := begin.UnixNano()
	tme := end.UnixNano()

	numOfKey := (tme - tmb) / int64(t.timestep.Nanoseconds())

	conn := t.pool.Get()
	defer conn.Close()

	for i := int64(0); i <= numOfKey; i++ {
		key := t.key(begin.Add(time.Duration(time.Duration(i) * t.timestep)))
		conn.Send("ZRANGEBYSCORE", key, tmb, tme)
	}
	conn.Flush()

	dumpData := make([][]string, numOfKey+1)
	rcCount := 0
	for i := int64(0); i <= numOfKey; i++ {
		dumpData[i], err = redis.Strings(conn.Receive())
		if err != nil {

			log.Println(err, conn.Err())
			return
		}
		rcCount += len(dumpData[i])
	}

	ensureLen(d, rcCount)
	i := 0
	for _, v := range dumpData {
		for _, r := range v {
			d := d.Index(i)
			var val interface{}
			if isPtr {
				if d.IsNil() {
					d.Set(reflect.New(typ))
				}
				val = d.Interface()
			} else {
				val = d.Addr().Interface()
			}

			json.Unmarshal([]byte(r), val)
			i++
		}
	}

	return
}

func (t *timeSeries) key(tm time.Time) string {
	tmi := tm.UnixNano()
	return fmt.Sprintf("%s::%d", t.prefix, t.normalizeTimeInt64(tmi))
}

func (t *timeSeries) normalizeTimeInt64(tm int64) int64 {
	return tm - (tm % int64(t.timestep.Nanoseconds()))
}

func ensureLen(d reflect.Value, n int) {
	if n > d.Cap() {
		d.Set(reflect.MakeSlice(d.Type(), n, n))
	} else {
		d.SetLen(n)
	}
}
