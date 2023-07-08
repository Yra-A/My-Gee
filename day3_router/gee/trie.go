package gee

import (
	"strings"
)

type node struct {
	pattern  string  // 完整的待匹配路由
	part     string  // 当前节点的部分
	children []*node // 子节点
	isWild   bool    // 是否精确匹配，若 * 或 : 为 true
}

// 匹配到的第一个子节点
func (x *node) matchFirstChild(part string) *node {
	for _, child := range x.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 匹配到的所有子节点
func (x *node) matchChildren(part string) []*node {
	validNodes := make([]*node, 0)
	for _, child := range x.children {
		if child.part == part || child.isWild {
			validNodes = append(validNodes, child)
		}
	}
	return validNodes
}

// 往前缀路由树上插入一个节点
func (x *node) insertNode(pattern string, parts []string, height int) {
	if height == len(parts) {
		x.pattern = pattern
		return
	}
	part := parts[height]
	//fmt.Println(part)
	validChild := x.matchFirstChild(part)
	//fmt.Println(validChild)
	if validChild == nil {
		validChild = &node{
			part:   part,
			isWild: part[0] == '*' || part[0] == ':',
		}
		x.children = append(x.children, validChild)
	}
	validChild.insertNode(pattern, parts, height+1)
}

// 针对路径查找对应匹配节点
func (x *node) searchNode(parts []string, height int) *node {
	if height == len(parts) || strings.HasPrefix(x.part, "*") {
		if x.pattern == "" {
			return nil
		}
		return x
	}
	part := parts[height]
	validChildren := x.matchChildren(part)
	for _, child := range validChildren {
		if result := child.searchNode(parts, height+1); result != nil {
			return result
		}
	}
	return nil
}
