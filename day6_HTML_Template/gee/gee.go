package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
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

	htmlTemplates *template.Template // html 模板集合
	funcMap       template.FuncMap   // 自定义函数
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.routerGroup = &routerGroup{engine: engine}
	engine.groups = []*routerGroup{engine.routerGroup}
	return engine
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
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
	c.engine = engine
	engine.router.handle(c) // 调用对应处理函数
}

// filesystem 是文件系统，用来操作文件
// fileserver 是文件服务器，用来相应 http 请求
func (group *routerGroup) createStaticHandler(relative string, fs http.FileSystem) HandlerFunc {
	absolutePath := group.prefix + relative
	// http.FileServer(fs) http.StripPreix 都是实现了 http.Handler 接口的对象
	// http.StripPreix 会在 http.FileServer 的基础上将请求 URL 的路径去掉前缀 absolutePath
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.SetSatusCode(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// 在 group 组中，将 relative 路径，映射为 root，这样当访问 relative/*filepath 时，就会改为访问 root/*filepath
func (group *routerGroup) Static(relative string, root string) {
	handler := group.createStaticHandler(relative, http.Dir(root)) // http.Dir 实现了 http.FileSystem 接口，用来指明服务器从哪里提供文件
	patternURL := path.Join(relative, "/*filepath")
	group.GET(patternURL, handler)
}
