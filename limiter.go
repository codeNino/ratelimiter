package ratelimiter

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// bucket is the limit bucket
type bucket struct {
	Attempts  int       `json:"attempts"`    // number of attempts  made (number of requests made to API)
	BlockTill time.Time `json:"block_until"` // service unavailable when max attempts reached till
}

func (t *bucket) marshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *bucket) unmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	return nil
}

//  RateLimiter is the limiter instance
type RateLimiter struct {
	TotalLimit       int           // maximum allowed requests over all
	BurstLimit       int           // maximum allowed consecutive requests in a short burst
	MaxTime          time.Duration // period for maximum allowed requests
	BurstPeriod      time.Duration // period for short bursts
	Client           *redis.Client
	TotalLimitPrefix string // prefix for total limit key in memory cache
	BurstLimitPrefix string // prefix for bursts limit key in memory cache
}

// note total requests made with user parameters and update per new request made
func (limiter *RateLimiter) UpdateTotalRequests(user_params ...string) {
	key_arr := append(user_params, limiter.TotalLimitPrefix)
	key := strings.Join(key_arr, "_")
	var bck bucket
	err := bck.unmarshalBinary([]byte(limiter.Client.Get(key).Val()))
	if err != nil {
		limit := &bucket{Attempts: 1, BlockTill: time.Now().Add(limiter.MaxTime)}
		jsonified, _ := limit.marshalBinary()
		limiter.Client.Set(key, jsonified, time.Hour*24).Val()
	} else {
		updated_limit := &bucket{Attempts: bck.Attempts + 1,
			BlockTill: bck.BlockTill}
		jsonified, _ := updated_limit.marshalBinary()
		limiter.Client.Set(key, jsonified, bck.BlockTill.Sub(time.Now())).Val()
	}
}

// check if total requests made within specified limit before accepting new user request
func (limiter *RateLimiter) AllowWithinTotalRequests(user_params ...string) bool {
	key_arr := append(user_params, limiter.TotalLimitPrefix)
	key := strings.Join(key_arr, "_")
	var bck bucket
	err := bck.unmarshalBinary([]byte(limiter.Client.Get(key).Val()))
	if err != nil {
		return false
	} else {
		if bck.Attempts >= limiter.TotalLimit {
			if time.Now().After(bck.BlockTill) {
				limiter.Client.Del(key)
				return true
			}
			return false
		}
		return true
	}
}

// note consecutive requests in short bursts made with user parameters and update per new request made
func (limiter *RateLimiter) UpdateConsecutiveRequests(user_params ...string) {
	key_arr := append(user_params, limiter.BurstLimitPrefix)
	key := strings.Join(key_arr, "_")
	var bck bucket
	err := bck.unmarshalBinary([]byte(limiter.Client.Get(key).Val()))
	if err != nil {
		limit := &bucket{Attempts: 1, BlockTill: time.Now().Add(limiter.BurstPeriod)}
		jsonified, _ := limit.marshalBinary()
		limiter.Client.Set(key, jsonified, time.Hour*1)
	} else {
		updated_limit := &bucket{Attempts: bck.Attempts + 1,
			BlockTill: bck.BlockTill}
		jsonified, _ := updated_limit.marshalBinary()
		limiter.Client.Set(key, jsonified, bck.BlockTill.Sub(time.Now()))
	}
}

// check if consecutive requests made within specified limit before accepting new user request
func (limiter *RateLimiter) AllowConsecutiveRequest(user_params ...string) bool {
	key_arr := append(user_params, limiter.BurstLimitPrefix)
	key := strings.Join(key_arr, "_")
	var bck bucket
	err := bck.unmarshalBinary([]byte(limiter.Client.Get(key).Val()))
	if err != nil {
		return false
	} else {
		if bck.Attempts >= limiter.BurstLimit {
			if time.Now().After(bck.BlockTill) {
				limiter.Client.Del(key)
				return true
			}
			return false
		}
		return true

	}
}

// note consecutive and total requests in made with user parameters and update each as per new request made
func (limiter *RateLimiter) NoteRequest(user_params ...string) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		limiter.UpdateConsecutiveRequests(user_params...)
		wg.Done()
	}()
	go func() {
		limiter.UpdateTotalRequests(user_params...)
		wg.Done()
	}()
	wg.Wait()
}

// check if consecutive and total requests made within specified limit before accepting new user request
func (limiter *RateLimiter) AllowRequest(user_params ...string) bool {
	var allowTotal, allowConsec bool
	var wg sync.WaitGroup
	wg.Add(2)
	go func(result *bool) {
		*result = limiter.AllowConsecutiveRequest(user_params...)
		wg.Done()
	}(&allowConsec)
	go func(result *bool) {
		*result = limiter.AllowWithinTotalRequests(user_params...)
		wg.Done()
	}(&allowTotal)
	wg.Wait()
	if allowConsec && allowTotal {
		return true
	}
	return false

}
