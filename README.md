<h1 align="center">rate-limiter</h1>
<h5 align="center">This package provides a rate limiter in Go (Golang) with a middleware adaptable for Gin web servers.</h5>

### Installation
- Download

    Type the following command in your terminal.
    ```bash
    go get github.com/go-redis/redis
    go get github.com/codeNino/ratelimiter
    ``` 

- Import
    ```go 
    import "github.com/go-redis/redis"
    import limiter "github.com/codeNino/ratelimiter"
    import "time"
    ```

---

### Quickstart

- Create a limiter object
    ```go
    // Set redis client
    rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})

    // set up rate limiter object
    cop := &limiter.RateLimiter{
        TotalLimit : 100, BurstLimit : 10,
	    MaxTime :  time.Hour * 24,   BurstPeriod : time.Hour * 2
	    Client : rdb,  TotalLimitPrefix : "TotalLimitPrefixForRandomService"
	    BurstLimitPrefix : "BurstLimitPrefixForRandomService"
    }

    ```

- Add different rate limiter type middleware to controlling each route. 
    ```go
    server := gin.Default()

    // create middleware for 

    server.POST("/ExamplePost1", limiter.IPLimiter(cop), func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello ExamplePost1")
	})

	server.GET("/ExampleGet1", limiter.UserLimiter(cop, "sample_user_id"), func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello ExampleGet1")
	})

	err = server.Run(":8080")
	if err != nil {
		log.Println("gin server error = ", err)
	}
    ```

---

### Response 
- When the consecutive requests and total request times is within limit, the request is allowed to continue.

- When limit is reached, the request is aborted and a `429` HTTP status code is sent.

<hr>

<hr>


### License

All source code is licensed under the [MIT License](./LICENSE).


