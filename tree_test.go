package kid

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testHandlerFunc HandlerFunc = func(c *Context) {}

	testMiddlewareFunc MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			next(c)
		}
	}
)

// funcsAreEqual checks if two functions have the same pointer value.
func funcsAreEqual(x, y any) bool {
	return reflect.ValueOf(x).Pointer() == reflect.ValueOf(y).Pointer()
}

func TestNewNode(t *testing.T) {
	node := newNode()

	assert.Empty(t, node.handlerMap)
	assert.Empty(t, node.children)
}

func TestNewTree(t *testing.T) {
	tree := newTree()

	assert.NotNil(t, tree.root)
	assert.EqualValues(t, 1, tree.size)
}

func TestNode_getChild(t *testing.T) {
	node := newNode()

	childNode := newNode()
	childNode.label = "test"

	node.addChild(&childNode)

	assert.Equal(t, &childNode, node.getChild("test", childNode.isParam, childNode.isStar))
	assert.Nil(t, node.getChild("test", !childNode.isParam, childNode.isStar))
	assert.Nil(t, node.getChild("test", childNode.isParam, !childNode.isStar))
	assert.Nil(t, node.getChild("test2", childNode.isParam, childNode.isStar))
}

func TestNode_addChild(t *testing.T) {
	node := newNode()

	childNode := newNode()
	childNode.label = "test"

	node.addChild(&childNode)
	assert.Len(t, node.children, 1)
}

func TestNode_addHanlder(t *testing.T) {
	node := newNode()

	node.addHanlder([]string{http.MethodGet, http.MethodPost}, handlerMiddleware{})

	assert.Len(t, node.handlerMap, 2)

	assert.PanicsWithValue(
		t,
		"handler is already registered for method GET and node &{id:0 label: children:[] isParam:false isStar:false handlerMap:map[GET:{handler:<nil> middlewares:[]} POST:{handler:<nil> middlewares:[]}]}.",
		func() {
			node.addHanlder([]string{http.MethodGet, http.MethodPost}, handlerMiddleware{})
		},
	)
}

func TestIsParam(t *testing.T) {
	assert.True(t, isParam("{param}"))

	assert.False(t, isParam("param"))

	assert.False(t, isParam("param}"))

	assert.False(t, isParam("{param"))
}

func TestIsStar(t *testing.T) {
	assert.True(t, isStar("{*param}"))

	assert.False(t, isStar("{param}"))
}

func TestTree_insertNode(t *testing.T) {
	tree := newTree()

	tree.insertNode("/test/path", []string{http.MethodGet}, nil, testHandlerFunc)

	assert.False(t, tree.root.isParam)
	assert.False(t, tree.root.isStar)
	assert.Equal(t, "", tree.root.label)
	assert.EqualValues(t, 1, tree.root.id)

	child := tree.root.getChild("test", false, false)
	assert.False(t, child.isParam)
	assert.False(t, child.isStar)
	assert.Equal(t, "test", child.label)
	assert.EqualValues(t, 2, child.id)

	_, ok := child.handlerMap[http.MethodGet]
	assert.False(t, ok)

	child2 := child.getChild("path", false, false)
	assert.False(t, child2.isParam)
	assert.False(t, child2.isStar)
	assert.Equal(t, "path", child2.label)
	assert.EqualValues(t, 3, child2.id)

	hm, ok := child2.handlerMap[http.MethodGet]
	assert.True(t, ok)
	assert.True(t, funcsAreEqual(hm.handler, testHandlerFunc))
	assert.Nil(t, hm.middlewares)

	tree.insertNode("/test", []string{http.MethodPost}, []MiddlewareFunc{testMiddlewareFunc}, testHandlerFunc)

	assert.False(t, tree.root.isParam)
	assert.False(t, tree.root.isStar)
	assert.Equal(t, "", tree.root.label)

	child = tree.root.getChild("test", false, false)
	assert.False(t, child.isParam)
	assert.False(t, child.isStar)
	assert.Equal(t, "test", child.label)
	assert.EqualValues(t, 2, child.id)

	hm, ok = child.handlerMap[http.MethodPost]
	assert.True(t, ok)
	assert.True(t, funcsAreEqual(hm.handler, testHandlerFunc))
	assert.Len(t, hm.middlewares, 1)
}

func TestNode_insert_Panics(t *testing.T) {
	tree := newTree()

	assert.PanicsWithValue(t, "providing at least one method is required", func() {
		tree.insertNode("/test", []string{}, nil, nil)
	})

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		tree.insertNode("/test", []string{http.MethodGet}, nil, nil)
	})

	assert.PanicsWithValue(t, "star path parameters can only be the last part of a path", func() {
		tree.insertNode("/{*starParam}/test", []string{http.MethodGet}, nil, testHandlerFunc)
	})
}

func TestNode_setLabel(t *testing.T) {
	n := newNode()

	n.isParam = false
	n.setLabel("static")
	assert.Equal(t, "static", n.label)

	n.isParam = true
	n.setLabel("{param}")
	assert.Equal(t, "param", n.label)

	n.isStar = true
	n.setLabel("{*starParam}")
	assert.Equal(t, "starParam", n.label)
}

func TestCleanPath(t *testing.T) {
	slash := cleanPath("", true)

	assert.Equal(t, "/", slash)

	prefixSlash := cleanPath("test", true)

	assert.Equal(t, "/test", prefixSlash)

	cleanedPath := cleanPath("//api///v1////books/offer", false)

	assert.Equal(t, "/api/v1/books/offer", cleanedPath)
}

func TestDFS(t *testing.T) {
	tree := newTree()

	tree.insertNode("/{path}/path1", []string{http.MethodGet}, nil, testHandlerFunc)
	tree.insertNode("/{path}/path2", []string{http.MethodPost}, nil, testHandlerFunc)
	tree.insertNode("/{path}/path2/", []string{http.MethodDelete}, nil, testHandlerFunc)
	tree.insertNode("/path/test", []string{http.MethodPut}, nil, testHandlerFunc)
	tree.insertNode("/path1/{*starParam}", []string{http.MethodPatch}, nil, testHandlerFunc)

	testCases := []struct {
		method         string
		path           []string
		expectedParams Params
		found          bool
	}{
		{method: http.MethodGet, path: []string{"", "param1", "path1"}, expectedParams: Params{"path": "param1"}, found: true},
		{method: http.MethodPost, path: []string{"", "param2", "path2"}, expectedParams: Params{"path": "param2"}, found: true},
		{method: http.MethodDelete, path: []string{"", "param", "path2", ""}, expectedParams: Params{"path": "param"}, found: true},
		{method: http.MethodPut, path: []string{"", "path", "test"}, expectedParams: Params{}, found: true},
		{method: http.MethodGet, path: []string{"", "path"}, expectedParams: Params{}, found: false},
		{method: http.MethodPatch, path: []string{"", "path1", "param1"}, expectedParams: Params{"starParam": "param1"}, found: true},
		{method: http.MethodPatch, path: []string{"", "path1", "param1", "param2"}, expectedParams: Params{"starParam": "param1/param2"}, found: true},
		{method: http.MethodPatch, path: []string{"", "path1"}, expectedParams: Params{"starParam": ""}, found: true},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			stack := []*Node{tree.root}
			vm := map[uint32]bool{}
			params := make(Params)

			hmMap, params, found := searchDFS(stack, vm, params, testCase.path, 0)
			hm, ok := hmMap[testCase.method]

			assert.Equal(t, testCase.found, found)
			assert.Equal(t, testCase.found, ok)
			assert.Equal(t, testCase.expectedParams, params)

			if testCase.found {
				assert.NotEmpty(t, hm)
			} else {
				assert.Empty(t, hm)
			}
		})
	}
}

func TestNode_getPathParam(t *testing.T) {
	node := newNode()
	path := []string{"", "param", "path"}

	assert.Equal(t, "param", node.getPathParam(path, 1))

	node.isStar = true

	assert.Equal(t, "param/path", node.getPathParam(path, 1))
}

func TestNode_searchFinished(t *testing.T) {
	node := newNode()
	path := []string{"", "param", "path"}

	assert.False(t, node.searchFinished(path, 2))
	assert.False(t, node.searchFinished(path, 1))

	node.handlerMap[http.MethodGet] = handlerMiddleware{}

	assert.True(t, node.searchFinished(path, 2))
	assert.False(t, node.searchFinished(path, 1))

	node.isStar = true

	assert.True(t, node.searchFinished(path, 2))
	assert.True(t, node.searchFinished(path, 1))
}

func TestNode_doesMatch(t *testing.T) {
	node := newNode()
	node.isStar = true
	node.label = "lbl"

	assert.True(t, node.doesMatch([]string{"0", "1", "2"}, 0))

	node.isStar = false

	assert.False(t, node.doesMatch([]string{"0", "1", "2"}, 3))

	node.isParam = true

	assert.True(t, node.doesMatch([]string{"0", "1", "2"}, 2))
	assert.False(t, node.doesMatch([]string{"0", "1", ""}, 2))

	node.isParam = false

	assert.True(t, node.doesMatch([]string{"0", "1", "lbl"}, 2))
	assert.False(t, node.doesMatch([]string{"0", "1", "invalid"}, 2))
}

func TestTree_search(t *testing.T) {
	tree := newTree()

	tree.insertNode("/path/test", []string{http.MethodGet}, nil, testHandlerFunc)

	testCases := []struct {
		name, method, path string
		expectedErr        error
		expectedParams     Params
	}{
		{name: "ok", method: http.MethodGet, path: "/path/test", expectedErr: nil, expectedParams: Params{}},
		{name: "not_found", method: http.MethodGet, path: "/path/test2", expectedErr: errNotFound, expectedParams: nil},
		{name: "method_not_allowed", method: http.MethodPost, path: "/path/test", expectedErr: errMethodNotAllowed, expectedParams: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hm, params, err := tree.search(tc.path, tc.method)

			assert.ErrorIs(t, err, tc.expectedErr)
			assert.Equal(t, tc.expectedParams, params)

			if tc.expectedErr == nil {
				assert.NotEmpty(t, hm)
			} else {
				assert.Empty(t, hm)
			}
		})
	}
}
