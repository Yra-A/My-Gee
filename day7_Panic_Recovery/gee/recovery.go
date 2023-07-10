package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // 跳过前 3 个调用帧，将 pc 值保存在 pcs 中，返回了获取到的调用栈帧数

	var str strings.Builder
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)   // 获取每个 pc 指向的 func，fn 是一个 *Func
		file, line := fn.FileLine(pc) // 获取文件名，行号
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
	}
}
