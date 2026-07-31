package main

import (
	"fmt"
	"math/rand"
)

type template struct {
	service string
	level   string
	weight  int
	render  func(r *rand.Rand, v *vocab) string
}

// vocab holds the value pools a corpus draws from. The client IP pool is
// generated from the corpus seed instead of being a hand-written list, so a
// large corpus does not read as the same few addresses stamped over and over.
type vocab struct {
	users     []string
	clientIPs []string
}

func newVocab(r *rand.Rand) *vocab {
	ips := make([]string, 0, 512)
	for i := 0; i < 512; i++ {
		ips = append(ips, fmt.Sprintf("10.%d.%d.%d", 16+r.Intn(16), r.Intn(256), 1+r.Intn(254)))
	}
	return &vocab{users: userNames, clientIPs: ips}
}

var userNames = []string{
	"alice", "bob", "carol", "dave", "erin",
	"frank", "grace", "heidi", "ivan", "judy",
	"karen", "leo", "mona", "nate", "olga",
	"pete", "quinn", "rosa", "sam", "tina",
	"uma", "vince", "wendy", "xavier", "yuki",
	"zack", "aaron", "bella", "cody", "diana",
	"eddie", "fiona", "gary", "hana", "iris",
	"jack", "kyle", "lena", "marco", "nina",
}

var hosts = map[string][]string{
	"web":  {"web-1", "web-2"},
	"api":  {"api-1", "api-2"},
	"db":   {"db-1"},
	"auth": {"auth-1"},
}

var readPaths = []string{
	"/api/v1/orders", "/api/v1/items", "/api/v1/cart", "/api/v1/users", "/static/app.js",
}

var missPaths = []string{
	"/api/v1/coupons", "/favicon.ico", "/wp-login.php",
}

func pick(r *rand.Rand, list []string) string {
	return list[r.Intn(len(list))]
}

func between(r *rand.Rand, lo, hi int) int {
	return lo + r.Intn(hi-lo+1)
}

// The weights are the corpus identity: routine lines dominate, warnings are
// uncommon and errors are rare, so a free-text search returns a meaningful
// minority instead of noise.
var templates = []template{
	{"web", "info", 60, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("%s \"GET %s\" 200 %dms", pick(r, v.clientIPs), pick(r, readPaths), between(r, 3, 80))
	}},
	{"web", "info", 15, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("%s \"POST /api/v1/orders\" 201 %dms", pick(r, v.clientIPs), between(r, 15, 120))
	}},
	{"web", "info", 8, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("%s \"GET /healthz\" 200 %dms", pick(r, v.clientIPs), between(r, 1, 3))
	}},
	{"web", "warn", 4, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("slow request %s \"GET %s\" 200 %dms", pick(r, v.clientIPs), pick(r, readPaths), between(r, 1000, 4000))
	}},
	{"web", "warn", 2, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("%s \"GET %s\" 404 %dms", pick(r, v.clientIPs), pick(r, missPaths), between(r, 1, 5))
	}},
	{"web", "error", 1, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("upstream timeout %s \"POST /api/v1/orders\" 504 30000ms", pick(r, v.clientIPs))
	}},

	{"api", "info", 50, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("order %d created for user %s in %dms", between(r, 100000, 9999999), pick(r, v.users), between(r, 12, 90))
	}},
	{"api", "info", 12, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("cache miss for key items:%d, fetched from db in %dms", between(r, 1, 99999), between(r, 8, 40))
	}},
	{"api", "warn", 4, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("slow query on orders took %dms", between(r, 800, 3000))
	}},
	{"api", "warn", 2, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("retrying request to db-1:5432, attempt %d", between(r, 2, 3))
	}},
	{"api", "error", 1, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("connection refused to db-1:5432, dropping request for user %s", pick(r, v.users))
	}},
	{"api", "error", 1, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("request timeout after 5000ms for user %s", pick(r, v.users))
	}},

	{"db", "info", 20, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("connection authorized user=api database=shop pid=%d", between(r, 1000, 65000))
	}},
	{"db", "info", 10, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("checkpoint complete: wrote %d buffers in %d.%ds", between(r, 100, 2000), between(r, 4, 29), between(r, 0, 9))
	}},
	{"db", "info", 5, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("automatic vacuum of table \"orders\": removed %d dead tuples", between(r, 50, 3000))
	}},
	{"db", "warn", 3, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("duration: %dms statement: SELECT * FROM orders WHERE user_id = %d", between(r, 1000, 5000), between(r, 1, 20000))
	}},
	{"db", "error", 1, func(r *rand.Rand, v *vocab) string {
		return "deadlock detected while locking table \"orders\""
	}},

	{"auth", "info", 20, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("login succeeded for user %s from %s", pick(r, v.users), pick(r, v.clientIPs))
	}},
	{"auth", "info", 4, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("token refreshed for user %s", pick(r, v.users))
	}},
	{"auth", "warn", 2, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("login failed for user %s from %s: invalid password", pick(r, v.users), pick(r, v.clientIPs))
	}},
	{"auth", "warn", 1, func(r *rand.Rand, v *vocab) string {
		return fmt.Sprintf("token expired for user %s", pick(r, v.users))
	}},
	{"auth", "error", 1, func(r *rand.Rand, v *vocab) string {
		return "too many failed logins for user mallory, locking account for 15m"
	}},
}
