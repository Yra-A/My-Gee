package gee

import (
	"net/http"
)

type HandlerFunc func(*Context)

type Engine struct {
	router *router
}

func New() *Engine {
	return &Engine{router: newRouter()}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// add GET request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// add POST request
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// run and start a http server
func (engine *Engine) RUN(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// 实现 http.Handler 接口的 ServeHTTP 方法
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req) // 用响应和请求，创建一个Context
	engine.router.handle(c) // 调用对应处理函数
}
