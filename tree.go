package kid

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Errors.
var (
	errNotFound         = errors.New("match not found")
	errMethodNotAllowed = errors.New("method is not allowed")
)

// Path parameters prefix and suffix.
const (
	paramPrefix     = "{"
	paramSuffix     = "}"
	plusParamPrefix = paramPrefix + "+"
	starParamPrefix = paramPrefix + "*"
)

type (
	// handlerMiddleware zips a handler and its middlewares to each other.
	handlerMiddleware struct {
		// handler is the route request handler.
		handler HandlerFunc

		// middlewares is route middlewares.
		middlewares []MiddlewareFunc

		// name is route name.
		name string
	}

	// Tree is a tree used for routing.
	Tree struct {
		// size is the number of nodes in the tree.
		size uint32

		// root of the tree.
		root *Node
	}

	// Node is a tree node.
	Node struct {
		// id of each node. It separates nodes from each other.
		id uint32

		// label of the node.
		label string

		children []*Node

		isParam bool
		isStar  bool

		// handlerMap maps HTTP methods to their handlers.
		handlerMap map[string]handlerMiddleware
	}

	// Params is the type of path parameters.
	Params map[string]string
)

// newNode returns a new node.
func newNode() Node {
	return Node{
		children:   make([]*Node, 0),
		handlerMap: make(map[string]handlerMiddleware),
	}
}

// newTree returns a new Tree.
func newTree() Tree {
	node := newNode()
	node.id = 1

	return Tree{
		size: 1,
		root: &node,
	}
}

// insert inserts a new node into the tree.
func (t *Tree) insertNode(path string, methods []string, middlewares []MiddlewareFunc, handler HandlerFunc) {
	if len(methods) == 0 {
		panic("providing at least one method is required")
	}

	panicIfNil(handler, "handler cannot be nil")

	path = cleanPath(path, false)

	segments := strings.Split(path, "/")[1:]

	currNode := t.root

	for i, segment := range segments {
		node := newNode()

		node.isParam = isParam(segment)
		node.isStar = isStar(segment)
		node.setLabel(segment)
		node.id = t.size + 1
		t.size++

		if i != len(segments)-1 {
			if node.isStar {
				panic("star path parameters can only be the last part of a path")
			}

			if child := currNode.getChild(node.label, node.isParam, node.isStar); child == nil {
				currNode.addChild(&node)
				currNode = &node
			} else {
				currNode = child
			}
		} else { // Only for the last iteration of the for loop.
			hm := handlerMiddleware{handler: handler, middlewares: middlewares, name: path}
			if child := currNode.getChild(node.label, node.isParam, node.isStar); child == nil {
				node.addHanlder(methods, hm)
				currNode.addChild(&node)
			} else {
				child.addHanlder(methods, hm)
			}
		}
	}
}

// doesMatch deterines if the path matches the node's label.
func (n Node) doesMatch(path []string, pos int) bool {
	if n.isStar {
		return true
	}

	if pos >= len(path) {
		return false
	}

	// Param matching.
	if n.isParam {
		return path[pos] != ""
	}

	// Exact matching.
	return path[pos] == n.label
}

// searchFinished returns true if the search has to be finished.
func (n Node) searchFinished(path []string, pos int) bool {
	if pos+1 == len(path) && len(n.handlerMap) > 0 {
		return true
	}
	return n.isStar
}

// getPathParam returns the path parameter.
func (n *Node) getPathParam(path []string, pos int) string {
	if n.isStar {
		return strings.Join(path[pos:], "/")
	}

	return path[pos]
}

// getChild returns the specified child of the node.
func (n Node) getChild(label string, isParam, isStar bool) *Node {
	for i := 0; i < len(n.children); i++ {
		if n.children[i].label == label && n.children[i].isParam == isParam && n.children[i].isStar == isStar {
			return n.children[i]
		}
	}

	return nil
}

// addChild adds the given node to the node's children.
func (n *Node) addChild(node *Node) {
	n.children = append(n.children, node)
}

// addHanlders add handlers to their methods.
func (n *Node) addHanlder(methods []string, hm handlerMiddleware) {
	for _, v := range methods {
		if _, ok := n.handlerMap[v]; ok {
			panic(fmt.Sprintf("handler is already registered for method %s and node %+v.", v, n))
		}

		n.handlerMap[v] = hm
	}
}

// setLabel sets the node's appropriate label.
func (n *Node) setLabel(label string) {
	n.label = label
	if n.isParam {
		if n.isStar {
			n.label = label[2 : len(label)-1]
		} else {
			n.label = label[1 : len(label)-1]
		}
	}
}

// isParam determines if a label is a parameter.
func isParam(label string) bool {
	if strings.HasPrefix(label, paramPrefix) && strings.HasSuffix(label, paramSuffix) {
		return true
	}
	return false
}

// isStar checks if a parameter is a star parameter.
func isStar(label string) bool {
	if isParam(label) && label[1] == '*' {
		return true
	}
	return false
}

// searchDFS searches the tree with the DFS search algorithm.
func (t Tree) searchDFS(path []string) (map[string]handlerMiddleware, Params, bool) {
	stack := []*Node{t.root}
	visitedMap := map[uint32]bool{}
	params := make(Params)
	var pos int

	// Search while stack is not empty.
SearchLoop:
	for len(stack) != 0 {
		// Accessing last element.
		node := stack[len(stack)-1]

		if !visitedMap[node.id] {
			visitedMap[node.id] = true

			if node.isParam {
				params[node.label] = node.getPathParam(path, pos)
			}

			if node.searchFinished(path, pos) {
				return node.handlerMap, params, true
			}
		}

		for _, child := range node.children {
			if !visitedMap[child.id] && child.doesMatch(path, pos+1) {
				// Push child to the stack.
				stack = append(stack, child)
				pos++
				continue SearchLoop
			}
		}

		if node.isParam {
			delete(params, node.label)
		}

		// Pop from stack.
		stack = stack[:len(stack)-1]
		pos--
	}

	return nil, params, false
}

// search searches the Tree and tries to match the path to a handler if possible.
func (t Tree) search(path, method string) (handlerMiddleware, Params, error) {
	segments := strings.Split(path, "/")

	hmMap, params, found := t.searchDFS(segments)

	if !found {
		return handlerMiddleware{}, params, errNotFound
	}

	if hm, ok := hmMap[method]; ok {
		return hm, params, nil
	}

	return handlerMiddleware{}, params, errMethodNotAllowed
}

// cleanPath normalizes the path.
//
// If soft is false it also removes duplicate slashes.
func cleanPath(s string, soft bool) string {
	if s == "" {
		return "/"
	}

	if s[0] != '/' {
		s = "/" + s
	}

	if soft {
		return s
	}

	// Removing repeated slashes.
	var buff bytes.Buffer
	for i := 0; i < len(s); i++ {
		if i != 0 && s[i] == '/' && s[i-1] == '/' {
			continue
		}
		buff.WriteByte(s[i])
	}

	return buff.String()
}
