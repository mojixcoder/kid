package kid

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/mojixcoder/kid/serializer"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	k := New()

	assert.NotNil(t, k)
	assert.Equal(t, newTree(), k.router)
	assert.Equal(t, 0, len(k.middlewares))
	assert.Equal(t, serializer.NewJSONSerializer(), k.jsonSerializer)
	assert.Equal(t, serializer.NewXMLSerializer(), k.xmlSerializer)
	assert.True(t, funcsAreEqual(defaultNotFoundHandler, k.notFoundHandler))
	assert.True(t, funcsAreEqual(defaultMethodNotAllowedHandler, k.methodNotAllowedHandler))
	assert.True(t, k.Debug())
}

func TestKid_Use(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "middleware cannot be nil", func() {
		k.Use(nil)
	})

	assert.Equal(t, 0, len(k.middlewares))

	k.Use(testMiddlewareFunc)

	assert.Equal(t, 1, len(k.middlewares))
}

func TestKid_Get(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Get("/", nil)
	})

	k.Get("/test", func(c *Context) {
		c.JSON(http.StatusOK, Map{"message": "ok"})
	})

	k.Get("/greet/{name}", func(c *Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, Map{"message": fmt.Sprintf("Hello %s", name)})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"ok\"}\n", res.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/greet/human", nil)
	res = httptest.NewRecorder()

	k.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"Hello human\"}\n", res.Body.String())
}

func TestKid_Post(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Post("/", nil)
	})

	k.Post("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"ok\"}\n", res.Body.String())
}

func TestKid_Put(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Put("/", nil)
	})

	k.Put("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": "put"})
	})

	req := httptest.NewRequest(http.MethodPut, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"put\"}\n", res.Body.String())
}

func TestKid_Delete(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Delete("/", nil)
	})

	k.Delete("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": "deleted"})
	})

	req := httptest.NewRequest(http.MethodDelete, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"deleted\"}\n", res.Body.String())
}

func TestKid_Patch(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Patch("/", nil)
	})

	k.Patch("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": "patch"})
	})

	req := httptest.NewRequest(http.MethodPatch, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"patch\"}\n", res.Body.String())
}

func TestKid_Trace(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Trace("/", nil)
	})

	k.Trace("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": "trace"})
	})

	req := httptest.NewRequest(http.MethodTrace, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"trace\"}\n", res.Body.String())
}

func TestKid_Connect(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Connect("/", nil)
	})

	k.Connect("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": "connect"})
	})

	req := httptest.NewRequest(http.MethodConnect, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"connect\"}\n", res.Body.String())
}

func TestKid_Options(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Options("/", nil)
	})

	k.Options("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": "options"})
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"options\"}\n", res.Body.String())
}

func TestKid_Head(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Head("/", nil)
	})

	k.Head("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": "head"})
	})

	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"head\"}\n", res.Body.String())
}

func TestKid_Add(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Add("/", nil, []string{http.MethodGet, http.MethodPost})
	})

	k.Add("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": c.Request().Method})
	}, []string{http.MethodGet, http.MethodPost})

	testCases := []struct {
		req            *http.Request
		res            *httptest.ResponseRecorder
		expectedMethod string
	}{
		{req: httptest.NewRequest(http.MethodPost, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodPost},
		{req: httptest.NewRequest(http.MethodGet, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodGet},
	}

	for _, testCase := range testCases {
		t.Run(testCase.expectedMethod, func(t *testing.T) {
			k.ServeHTTP(testCase.res, testCase.req)

			assert.Equal(t, http.StatusCreated, testCase.res.Code)
			assert.Equal(t, "application/json", testCase.res.Header().Get("Content-Type"))
			assert.Equal(t, fmt.Sprintf("{\"message\":%q}\n", testCase.expectedMethod), testCase.res.Body.String())
		})
	}
}

func TestKid_Any(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		k.Any("/", nil)
	})

	k.Any("/test", func(c *Context) {
		c.JSON(http.StatusCreated, Map{"message": c.Request().Method})
	})

	testCases := []struct {
		req            *http.Request
		res            *httptest.ResponseRecorder
		expectedMethod string
	}{
		{req: httptest.NewRequest(http.MethodHead, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodHead},
		{req: httptest.NewRequest(http.MethodDelete, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodDelete},
		{req: httptest.NewRequest(http.MethodPost, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodPost},
		{req: httptest.NewRequest(http.MethodPut, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodPut},
		{req: httptest.NewRequest(http.MethodPatch, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodPatch},
		{req: httptest.NewRequest(http.MethodGet, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodGet},
		{req: httptest.NewRequest(http.MethodTrace, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodTrace},
		{req: httptest.NewRequest(http.MethodConnect, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodConnect},
		{req: httptest.NewRequest(http.MethodOptions, "/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodOptions},
	}

	for _, testCase := range testCases {
		t.Run(testCase.expectedMethod, func(t *testing.T) {
			k.ServeHTTP(testCase.res, testCase.req)

			assert.Equal(t, http.StatusCreated, testCase.res.Code)
			assert.Equal(t, "application/json", testCase.res.Header().Get("Content-Type"))
			assert.Equal(t, fmt.Sprintf("{\"message\":%q}\n", testCase.expectedMethod), testCase.res.Body.String())
		})
	}
}

func TestKid_Group(t *testing.T) {
	k := New()
	g := k.Group("/v1")

	g.Get("/path", func(c *Context) {
		c.JSON(http.StatusOK, Map{"message": "group"})
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/path", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"group\"}\n", res.Body.String())
}

func TestKid_applyMiddlewaresToHandler(t *testing.T) {
	k := New()

	middlewares := []MiddlewareFunc{
		func(next HandlerFunc) HandlerFunc {
			return func(c *Context) {
				c.Set("key1", 10)
				next(c)
			}
		},
		func(next HandlerFunc) HandlerFunc {
			return func(c *Context) {
				c.Set("key2", 20)
				next(c)
			}
		},
	}

	handler := k.applyMiddlewaresToHandler(func(c *Context) {
		val1, _ := c.Get("key1")
		val2, _ := c.Get("key2")
		c.JSON(http.StatusOK, Map{"key1": val1, "key2": val2})
	}, middlewares...)

	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	res := httptest.NewRecorder()

	c := newContext(k)
	c.reset(req, res)

	handler(c)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"key1\":10,\"key2\":20}\n", res.Body.String())
}

func TestKid_ServeHTTP_NotFound(t *testing.T) {
	k := New()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"Not Found\"}\n", res.Body.String())
}

func TestKid_ServeHTTP_MethodnotAllowed(t *testing.T) {
	k := New()

	k.Get("/test", testHandlerFunc)

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"Method Not Allowed\"}\n", res.Body.String())
}

func TestKid_ServeHTTP_WriteStatusCodeIfNotWritten(t *testing.T) {
	k := New()

	k.Get("/test", func(c *Context) {
		c.Response().WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusCreated, res.Code)
}

func TestKid_Debug(t *testing.T) {
	k := New()
	k.debug = false

	assert.False(t, k.Debug())

	k.debug = true
	assert.True(t, k.Debug())
}

func TestKid_Run(t *testing.T) {
	k := New()

	k.Get("/", func(c *Context) {
		c.JSON(http.StatusOK, Map{"message": "healthy"})
	})

	go func() {
		assert.NoError(t, k.Run(":8080"))
	}()

	// Wait for the server to start
	time.Sleep(5 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080")
	assert.NoError(t, err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"healthy\"}\n", string(body))
}

func TestKid_Static(t *testing.T) {
	k := New()

	k.Static("/static/", "testdata/static")

	testCases := []struct {
		name               string
		req                *http.Request
		res                *httptest.ResponseRecorder
		expectedStatusCode int
		expectedContent    string
	}{
		{
			name:               "Serving main.html",
			req:                httptest.NewRequest(http.MethodGet, "/static/main.html", nil),
			res:                httptest.NewRecorder(),
			expectedStatusCode: http.StatusOK,
			expectedContent:    "main",
		},
		{
			name:               "Serving page.html in pages directory",
			req:                httptest.NewRequest(http.MethodGet, "/static/pages/page.html", nil),
			res:                httptest.NewRecorder(),
			expectedStatusCode: http.StatusOK,
			expectedContent:    "page",
		},
		{
			name:               "Serving pages/index.html",
			req:                httptest.NewRequest(http.MethodGet, "/static/pages/", nil),
			res:                httptest.NewRecorder(),
			expectedStatusCode: http.StatusOK,
			expectedContent:    "index",
		},
		{
			name:               "Non-existent",
			req:                httptest.NewRequest(http.MethodGet, "/static/doesn't-exist.html", nil),
			res:                httptest.NewRecorder(),
			expectedStatusCode: http.StatusNotFound,
			expectedContent:    "404 page not found\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			k.ServeHTTP(testCase.res, testCase.req)

			assert.Equal(t, testCase.expectedStatusCode, testCase.res.Code)
			assert.Equal(t, testCase.expectedContent, testCase.res.Body.String())
		})
	}
}

func TestKid_StaticFS(t *testing.T) {
	k := New()

	k.StaticFS("/static/", http.Dir("testdata/static"))

	testCases := []struct {
		name               string
		req                *http.Request
		res                *httptest.ResponseRecorder
		expectedStatusCode int
		expectedContent    string
	}{
		{
			name:               "Serving main.html",
			req:                httptest.NewRequest(http.MethodGet, "/static/main.html", nil),
			res:                httptest.NewRecorder(),
			expectedStatusCode: http.StatusOK,
			expectedContent:    "main",
		},
		{
			name:               "Serving page.html in pages directory",
			req:                httptest.NewRequest(http.MethodGet, "/static/pages/page.html", nil),
			res:                httptest.NewRecorder(),
			expectedStatusCode: http.StatusOK,
			expectedContent:    "page",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			k.ServeHTTP(testCase.res, testCase.req)

			assert.Equal(t, testCase.expectedStatusCode, testCase.res.Code)
			assert.Equal(t, testCase.expectedContent, testCase.res.Body.String())
		})
	}
}

func TestKid_printDebug(t *testing.T) {
	k := New()

	var w bytes.Buffer

	k.printDebug(&w, "hello %s\n", "Kid")
	assert.Equal(t, "[DEBUG] hello Kid\n", w.String())

	w.Reset()
	k.debug = false

	k.printDebug(&w, "hello %s\n", "Kid")
	assert.Empty(t, w.String())
}

func TestResolveAddress(t *testing.T) {
	goos := "windows"
	addr := resolveAddress([]string{}, goos)
	assert.Equal(t, "127.0.0.1:2376", addr)

	goos = "linux"
	addr = resolveAddress([]string{}, goos)
	assert.Equal(t, "0.0.0.0:2376", addr)

	addr = resolveAddress([]string{":2377", ":2378"}, goos)
	assert.Equal(t, ":2377", addr)
}

func TestGetPath(t *testing.T) {
	u, err := url.Parse("http://localhost/foo%25fbar")
	assert.NoError(t, err)

	assert.Empty(t, u.RawPath)
	assert.Equal(t, u.Path, getPath(u))

	u, err = url.Parse("http://localhost/foo%fbar")
	assert.NoError(t, err)

	assert.NotEmpty(t, u.RawPath)
	assert.Equal(t, u.RawPath, getPath(u))
}

func TestApplyOptions(t *testing.T) {
	k := New()

	assert.PanicsWithValue(t, "option cannot be nil", func() {
		k.ApplyOptions(nil)
	})

	k.ApplyOptions(
		WithDebug(true),
	)

	assert.True(t, k.Debug())
}

func TestPanicIfNil(t *testing.T) {
	assert.PanicsWithValue(t, "nil", func() {
		panicIfNil(nil, "nil")
	})

	assert.Panics(t, func() {
		var x *int
		panicIfNil(x, "")
	})

	assert.Panics(t, func() {
		var x []string
		panicIfNil(x, "")
	})

	assert.Panics(t, func() {
		var x map[string]string
		panicIfNil(x, "")
	})

	assert.Panics(t, func() {
		var x interface{}
		panicIfNil(x, "")
	})

	assert.Panics(t, func() {
		var x func()
		panicIfNil(x, "")
	})

	assert.Panics(t, func() {
		var x chan bool
		panicIfNil(x, "")
	})

	assert.Panics(t, func() {
		var x [1]int
		panicIfNil(x, "")
	})
}

func TestKid_NewContext(t *testing.T) {
	k := New()

	ctx := k.NewContext(nil, nil)

	assert.NotNil(t, ctx)
}
