package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request

	Path       string
	Method     string
	StatusCode int
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

// 获取POST表单中的数据
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// 获取URL查询参数
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// 设置Context的状态码
func (c *Context) SetSatusCode(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code) // 写入头文件
}

// 设置响应报文头部中的值
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// 向响应报文中写入格式化字符串
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.SetSatusCode(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// 向响应报文中写入JSON
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.SetSatusCode(code)
	encoder := json.NewEncoder(c.Writer) // 构造一个encoder，其会将obj编码为json并写入c.Writer
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500) // 返回一个带有错误信息的响应报文
	}
}

// 写入数据
func (c *Context) Data(code int, data []byte) {
	c.SetSatusCode(code)
	c.Writer.Write(data)
}

// 写入HTML
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Data(code, []byte(html))
}
