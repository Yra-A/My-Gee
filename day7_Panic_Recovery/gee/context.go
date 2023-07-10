package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request

	// request info
	Path   string
	Method string
	Params map[string]string

	// response info
	StatusCode int

	// middleware
	handlers []HandlerFunc // 要执行的中间件
	index    int           // 当前执行到第几个中间件

	//engine pointer
	engine *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

// 执行其他中间件或用户定义的 handler
// 如果别的中间件里还有 Next()，会继续执行下去，直到最后一个执行完毕，然后从后往前以此执行 Next() 后面的内容
func (c *Context) Next() {
	c.index++
	for ; c.index < len(c.handlers); c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
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

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

// 写入HTML
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.SetSatusCode(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
