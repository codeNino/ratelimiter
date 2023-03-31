package ratelimiter

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetIPAddress(c *gin.Context) (string, error) {
	//Get IP from the X-REAL-IP header
	ip := c.Request.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//Get IP from X-FORWARDED-FOR header
	ips := c.Request.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}
	return "", fmt.Errorf("No valid ip found")
}

// enforce rate limit based on ip address of client making the request
func IPLimiter(limiter RateLimiter) gin.HandlerFunc {

	return func(c *gin.Context) {
		ip_addr, err := GetIPAddress(c)

		if err != nil {
			log.Println(err)
			c.AbortWithError(401, errors.New("IP Adress Unknown"))
		}

		if !limiter.AllowRequest(ip_addr) {
			c.AbortWithError(429, errors.New("Too Many Requests Made, Retry Later"))
		}

		c.Next()
	}
}

// enforce rate limit based on the user making the request
func UserLimiter(limiter RateLimiter, user_id string) gin.HandlerFunc {

	return func(c *gin.Context) {

		if !limiter.AllowRequest(user_id) {
			c.AbortWithError(429, errors.New("Too Many Requests Made, Retry Later"))
		}

		c.Next()
	}
}

// enforce rate limit based on ip address of client and user making the request
func IPUserLimiter(limiter RateLimiter, user_id string) gin.HandlerFunc {

	return func(c *gin.Context) {
		ip_addr, err := GetIPAddress(c)

		if err != nil {
			log.Println(err)
			c.AbortWithError(401, errors.New("IP Adress Unknown"))
		}

		if !limiter.AllowRequest(ip_addr, user_id) {
			c.AbortWithError(429, errors.New("Too Many Requests Made, Retry Later"))
		}

		c.Next()
	}
}
