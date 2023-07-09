package gee

import (
	"log"
	"net/http"
	"strings"
)

type HandlerFunc func(*Context)

type routerGroup struct {
	prefix      string
	middlewares []HandlerFunc // 当前分组支持的中间件
	parent      *routerGroup  // 父分组
	engine      *Engine       // 所有分组共享一个 Engine
}

type Engine struct {
	*routerGroup
	router *router
	groups []*routerGroup // 存储了所有的分组
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.routerGroup = &routerGroup{engine: engine}
	engine.groups = []*routerGroup{engine.routerGroup}
	return engine
}

// 在当前分组下创建一个新的子分组
func (group *routerGroup) Group(prefix string) *routerGroup {
	newGroup := &routerGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: group.engine,
	}
	group.engine.groups = append(group.engine.groups, newGroup)
	return newGroup
}

func (group *routerGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *routerGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// add GET request
func (group *routerGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// add POST request
func (group *routerGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// run and start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// 实现 http.Handler 接口的 ServeHTTP 方法
// 匹配当前请求适用的所有中间件
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req) // 用响应和请求，创建一个Context
	c.handlers = middlewares
	engine.router.handle(c) // 调用对应处理函数
}
