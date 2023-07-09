package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node // 每个 method 对应的trie树根节点
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// pasrse 路径，只允许有一个 *
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, s := range vs {
		if s != "" {
			parts = append(parts, s)
			if s[0] == '*' {
				break
			}
		}
	}
	return parts
}

// 添加路由，每个 method 对应着一棵前缀树的根
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insertNode(pattern, parts, 0)
	r.handlers[key] = handler // handler 函数映射
}

// 获取路径匹配到节点，以及路径上的参数映射，eg：{name: Yra}
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	pathParts := parsePattern(path)
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	goalNode := root.searchNode(pathParts, 0)
	params := make(map[string]string)
	if goalNode != nil {
		patternParts := parsePattern(goalNode.pattern)
		for i, key := range patternParts {
			if key[0] == ':' {
				params[key[1:]] = pathParts[i]
			}
			if key[0] == '*' && len(key) > 1 {
				params[key[1:]] = strings.Join(pathParts[i:], "/") // 将后面所有内容整合，并用 '/' 分隔，并且不进行继续的匹配
				break
			}
		}
		return goalNode, params
	}
	return nil, nil
}

// 调用处理函数
func (r *router) handle(c *Context) {
	x, params := r.getRoute(c.Method, c.Path)
	if x != nil {
		c.Params = params
		key := c.Method + "-" + x.pattern
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
