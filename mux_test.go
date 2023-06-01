package kid

import (
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

func TestNewRouter(t *testing.T) {
	router := newRouter()

	assert.NotNil(t, router.routes)
	assert.Empty(t, router.routes)
}

func TestMethodExists(t *testing.T) {
	assert.False(t, methodExists("GET", []string{"POST", "DELETE"}))
	assert.True(t, methodExists("GET", []string{"POST", "DELETE", "GET"}))
}

func TestCleanPath(t *testing.T) {
	slash := cleanPath("", true)

	assert.Equal(t, "/", slash)

	prefixSlash := cleanPath("test", true)

	assert.Equal(t, "/test", prefixSlash)

	cleanedPath := cleanPath("//api///v1////books/offer", false)

	assert.Equal(t, "/api/v1/books/offer", cleanedPath)
}

func TestRouter_add(t *testing.T) {
	router := newRouter()

	assert.PanicsWithValue(t, "providing at least one method is required", func() {
		router.add("/", testHandlerFunc, nil, nil)
	})

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		router.add("/", nil, []string{http.MethodGet}, nil)
	})

	assert.PanicsWithValue(t, "plus/star path parameters can only be the last part of a path", func() {
		router.add("/path/{+extraPath}/asd", testHandlerFunc, []string{http.MethodGet}, nil)
	})

	assert.PanicsWithValue(t, "plus/star path parameters can only be the last part of a path", func() {
		router.add("/path/{+extraPath}/test/test2", testHandlerFunc, []string{http.MethodGet}, nil)
	})

	router.add("/test/list/", testHandlerFunc, []string{http.MethodGet}, nil)

	router.add("/test/{var}/get", testHandlerFunc, []string{http.MethodGet, http.MethodPost}, []MiddlewareFunc{testMiddlewareFunc})

	router.add("/test/{+extraPath}", testHandlerFunc, []string{http.MethodPost}, nil)

	router.add("/path/{+extraPath}/", testHandlerFunc, []string{http.MethodDelete}, nil)

	assert.Equal(t, 4, len(router.routes))

	testCases := []struct {
		route Route
		name  string
	}{
		{
			name: "/test/list/",
			route: Route{
				methods:     []string{http.MethodGet},
				handler:     testHandlerFunc,
				segments:    []Segment{{isParam: false, tpl: "test"}, {isParam: false, tpl: "list"}, {isParam: false, tpl: ""}},
				middlewares: nil,
			},
		},
		{
			name: "/test/{var}/get",
			route: Route{
				methods:     []string{http.MethodGet, http.MethodPost},
				handler:     testHandlerFunc,
				segments:    []Segment{{isParam: false, tpl: "test"}, {isParam: true, tpl: "var"}, {isParam: false, tpl: "get"}},
				middlewares: []MiddlewareFunc{testMiddlewareFunc},
			},
		},
		{
			name: "/test/{+extraPath}",
			route: Route{
				methods:     []string{http.MethodPost},
				handler:     testHandlerFunc,
				segments:    []Segment{{isParam: false, isPlus: false, tpl: "test"}, {isParam: true, isPlus: true, tpl: "extraPath"}},
				middlewares: nil,
			},
		},
		{
			name: "/path/{+extraPath}",
			route: Route{
				methods:     []string{http.MethodDelete},
				handler:     testHandlerFunc,
				segments:    []Segment{{isParam: false, isPlus: false, tpl: "path"}, {isParam: true, isPlus: true, tpl: "extraPath"}},
				middlewares: nil,
			},
		},
	}

	for i := 0; i < len(testCases); i++ {
		testCase := testCases[i]
		t.Run(testCase.name, func(t *testing.T) {
			route := router.routes[i]

			assert.Equal(t, testCase.route.methods, route.methods)
			assert.Equal(t, testCase.route.segments, route.segments)
			assert.Equal(t, len(testCase.route.middlewares), len(route.middlewares))
			assert.True(t, funcsAreEqual(testCase.route.handler, route.handler))

			for i := 0; i < len(testCase.route.middlewares); i++ {
				expectedMiddlewareFunc := testCase.route.middlewares[i]
				middlewareFunc := route.middlewares[i]

				assert.True(t, funcsAreEqual(expectedMiddlewareFunc, middlewareFunc))
			}
		})
	}
}

func TestRouter_match(t *testing.T) {
	router := newRouter()

	router.add("/", testHandlerFunc, []string{http.MethodGet}, nil)
	router.add("/test/{var}/get", testHandlerFunc, []string{http.MethodGet, http.MethodPost}, nil)
	router.add("/test/{var}/get/{+plusPath}", testHandlerFunc, []string{http.MethodPut}, nil)
	router.add("/test/{var}/path/{*starPath}", testHandlerFunc, []string{http.MethodGet}, nil)

	firstRoute := router.routes[0]
	secondRoute := router.routes[1]
	plusRoute := router.routes[2]
	starRoute := router.routes[3]

	// Don't need to add starting slash in route's match method as they are skipped in router's find method.
	// For matching we should match relative paths, not abosulute paths.

	// Testing first route.
	params, err := firstRoute.match("", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(params))

	params, err = firstRoute.match("", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(params))

	params, err = firstRoute.match("a", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	params, err = firstRoute.match("", http.MethodPost)
	assert.ErrorIs(t, err, errMethodNotAllowed)
	assert.Nil(t, params)

	params, err = firstRoute.match("/", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	// Testing second route.
	params, err = secondRoute.match("test/hello/get", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "hello"}, params)

	params, err = secondRoute.match("test/123/get", http.MethodPost)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123"}, params)

	params, err = secondRoute.match("test/hello/get/", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	params, err = secondRoute.match("test/hello/get", http.MethodPut)
	assert.ErrorIs(t, err, errMethodNotAllowed)
	assert.Nil(t, params)

	params, err = secondRoute.match("test/hello/get2", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	params, err = secondRoute.match("test/hello/get/extra", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	params, err = secondRoute.match("test/hello/", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	params, err = secondRoute.match("test/hello", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	// Path varibales are required and cannot be empty.
	params, err = secondRoute.match("test//get", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	// Testing plus route.
	params, err = plusRoute.match("test/123/get/extra/path", http.MethodPut)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123", "plusPath": "extra/path"}, params)

	params, err = plusRoute.match("test/123/get/extra", http.MethodPut)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123", "plusPath": "extra"}, params)

	params, err = plusRoute.match("test/123/get/extra/path", http.MethodGet)
	assert.ErrorIs(t, err, errMethodNotAllowed)
	assert.Nil(t, params)

	params, err = plusRoute.match("test//get/extra/path", http.MethodPut)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	// At least one extra path is required
	params, err = plusRoute.match("test/123/get/", http.MethodPut)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	// Testing star route.
	params, err = starRoute.match("test/123/path/star/path", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123", "starPath": "star/path"}, params)

	params, err = starRoute.match("test/123/path/star", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123", "starPath": "star"}, params)

	params, err = starRoute.match("test/123/path/", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123", "starPath": ""}, params)

	params, err = starRoute.match("test/123/path", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123", "starPath": ""}, params)

	params, err = starRoute.match("test/123/path/star/path", http.MethodPost)
	assert.ErrorIs(t, err, errMethodNotAllowed)
	assert.Nil(t, params)

	params, err = starRoute.match("test//path/star/path", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)
}

func TestRouter_find(t *testing.T) {
	router := newRouter()

	router.add("/", testHandlerFunc, []string{http.MethodGet}, nil)
	router.add("/test/hi", testHandlerFunc, []string{http.MethodGet}, nil)
	router.add("/test/{var}", testHandlerFunc, []string{http.MethodGet, http.MethodPost}, []MiddlewareFunc{testMiddlewareFunc})

	route, params, err := router.find("/", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(params))
	assert.Equal(t, []string{http.MethodGet}, route.methods)
	assert.Equal(t, []Segment{{tpl: "", isParam: false}}, route.segments)
	assert.Nil(t, route.middlewares)
	assert.True(t, funcsAreEqual(testHandlerFunc, route.handler))

	_, params, err = router.find("", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(params))

	route, params, err = router.find("/test/123", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123"}, params)
	assert.Equal(t, []string{http.MethodGet, http.MethodPost}, route.methods)
	assert.Equal(t, []Segment{{tpl: "test", isParam: false}, {tpl: "var", isParam: true}}, route.segments)
	assert.Equal(t, 1, len(route.middlewares))
	assert.True(t, funcsAreEqual(testHandlerFunc, route.handler))
	assert.True(t, funcsAreEqual(testMiddlewareFunc, route.middlewares[0]))

	_, params, err = router.find("test/123", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123"}, params)

	route, params, err = router.find("/test/123", http.MethodPost)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "123"}, params)
	assert.Equal(t, []string{http.MethodGet, http.MethodPost}, route.methods)
	assert.Equal(t, []Segment{{tpl: "test", isParam: false}, {tpl: "var", isParam: true}}, route.segments)
	assert.Equal(t, 1, len(route.middlewares))
	assert.True(t, funcsAreEqual(testHandlerFunc, route.handler))
	assert.True(t, funcsAreEqual(testMiddlewareFunc, route.middlewares[0]))

	_, params, err = router.find("/test/123/", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	_, params, err = router.find("/test/123/", http.MethodPost)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	_, params, err = router.find("/test/123", http.MethodPut)
	assert.ErrorIs(t, err, errMethodNotAllowed)
	assert.Nil(t, params)

	_, params, err = router.find("/test/123/extra", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	_, params, err = router.find("/test", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	_, params, err = router.find("/test/", http.MethodGet)
	assert.ErrorIs(t, err, errNotFound)
	assert.Nil(t, params)

	// The first added methods have higher priority.
	route, params, err = router.find("/test/hi", http.MethodGet)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(params))
	assert.Equal(t, []string{http.MethodGet}, route.methods)
	assert.Equal(t, []Segment{{tpl: "test", isParam: false}, {tpl: "hi", isParam: false}}, route.segments)
	assert.Nil(t, route.middlewares)
	assert.True(t, funcsAreEqual(testHandlerFunc, route.handler))

	route, params, err = router.find("/test/hi", http.MethodPost)
	assert.NoError(t, err)
	assert.Equal(t, Params{"var": "hi"}, params)
	assert.Equal(t, []string{http.MethodGet, http.MethodPost}, route.methods)
	assert.Equal(t, []Segment{{tpl: "test", isParam: false}, {tpl: "var", isParam: true}}, route.segments)
	assert.Equal(t, 1, len(route.middlewares))
	assert.True(t, funcsAreEqual(testHandlerFunc, route.handler))
	assert.True(t, funcsAreEqual(testMiddlewareFunc, route.middlewares[0]))
}

func TestRouter_isPlus(t *testing.T) {
	router := newRouter()

	isPlus := router.isPlus("{+param}")
	assert.True(t, isPlus)

	isPlus = router.isPlus("{*param}")
	assert.False(t, isPlus)
}
