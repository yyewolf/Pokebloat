package utilities

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var Cache *cache.Cache

func init() {
	Cache = cache.New(5*time.Minute, 10*time.Minute)
}
