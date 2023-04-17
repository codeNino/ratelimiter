package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis"

	"github.com/gin-gonic/gin"
)

var rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})

// create a rate limiter instance
var cop = &RateLimiter{
	TotalLimit: 1, BurstLimit: 1,
	MaxTime: time.Hour * 1, BurstPeriod: time.Minute * 5,
	Client: rdb, TotalLimitPrefix: "TotalLimitPrefixForRandomService",
	BurstLimitPrefix: "BurstLimitPrefixForRandomService"}

func setup() {
	gin.SetMode(gin.TestMode)
}

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	os.Exit(exitCode)

}

// Helper function to create a router during testing
func getRouter(withTemplates bool) *gin.Engine {
	r := gin.Default()
	return r
}

func testHTTPResponse(t *testing.T, r *gin.Engine, req *http.Request, f func(w *httptest.ResponseRecorder) bool) {

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create the service and process the above request.
	r.ServeHTTP(w, req)

	if !f(w) {
		t.Fail()
	}
}

// Helper function to process a request and test its response
func testMiddlewareRequest(url string, t *testing.T, r *gin.Engine, expectedHTTPCode int) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-REAL-IP", "105.112.114.123")

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		return w.Code == expectedHTTPCode
	})
}

// Test that a GET request to the ExampleGet1 endpoint returns the HTTP code 200
func TestFirstGetExample1Call(t *testing.T) {

	rdb.FlushDB()

	r := getRouter(true)

	r.GET("/ExampleGet1", UserLimiter(*cop, "dummy-username"), func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello ExampleGet1")
	})

	testMiddlewareRequest("/ExampleGet1", t, r, http.StatusOK)

}

func TestSubsequentGetExample1Call(t *testing.T) {

	r := getRouter(true)

	r.GET("/ExampleGet1", UserLimiter(*cop, "dummy-username"), func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello ExampleGet1")
	})

	testMiddlewareRequest("/ExampleGet1", t, r, http.StatusTooManyRequests)

}
