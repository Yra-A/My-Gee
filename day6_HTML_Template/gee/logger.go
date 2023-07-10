package gee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		t := time.Now()                                                            // 处理请求开始的时间
		c.Next()                                                                   // 处理请求
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t)) //打印到日志中
	}
}
