package kid

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/mojixcoder/kid/errors"
	"github.com/mojixcoder/kid/serializer"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	k := New()

	assert.NotNil(t, k)
	assert.Equal(t, newRouter(), k.router)
	assert.Equal(t, 0, len(k.middlewares))
	assert.Equal(t, serializer.NewJSONSerializer(), k.jsonSerializer)
	assert.Equal(t, serializer.NewXMLSerializer(), k.xmlSerializer)
	assert.True(t, funcsAreEqual(defaultErrorHandler, k.errorHandler))
	assert.True(t, funcsAreEqual(defaultNotFoundHandler, k.notFoundHandler))
	assert.True(t, funcsAreEqual(defaultMethodNotAllowedHandler, k.methodNotAllowedHandler))
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

	k.Get("/test", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"message": "ok"})
	})

	k.Get("/greet/{name}", func(c *Context) error {
		name := c.Param("name")
		return c.JSON(http.StatusOK, Map{"message": fmt.Sprintf("Hello %s", name)})
	})

	assert.Equal(t, 2, len(k.router.routes))
	assert.Equal(t, 1, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, http.MethodGet, k.router.routes[0].methods[0])

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

	k.Post("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": "ok"})
	})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 1, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, http.MethodPost, k.router.routes[0].methods[0])

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

	k.Put("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": "put"})
	})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 1, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, http.MethodPut, k.router.routes[0].methods[0])

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

	k.Delete("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": "deleted"})
	})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 1, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, http.MethodDelete, k.router.routes[0].methods[0])

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

	k.Patch("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": "patch"})
	})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 1, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, http.MethodPatch, k.router.routes[0].methods[0])

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

	k.Trace("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": "trace"})
	})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 1, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, http.MethodTrace, k.router.routes[0].methods[0])

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

	k.Connect("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": "connect"})
	})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 1, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, http.MethodConnect, k.router.routes[0].methods[0])

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

	k.Options("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": "options"})
	})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 1, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, http.MethodOptions, k.router.routes[0].methods[0])

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

	k.Head("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": "head"})
	})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 1, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, http.MethodHead, k.router.routes[0].methods[0])

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

	k.Add("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": c.Request().Method})
	}, []string{http.MethodGet, http.MethodPost})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 2, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t, []string{http.MethodGet, http.MethodPost}, k.router.routes[0].methods)

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

	k.Any("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": c.Request().Method})
	})

	assert.Equal(t, 1, len(k.router.routes))
	assert.Equal(t, 9, len(k.router.routes[0].methods))
	assert.Equal(t, 0, len(k.router.routes[0].middlewares))
	assert.Equal(t,
		[]string{
			http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodPatch, http.MethodDelete, http.MethodHead,
			http.MethodOptions, http.MethodConnect, http.MethodTrace,
		},
		k.router.routes[0].methods,
	)

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

func TestKid_applyMiddlewaresToHandler(t *testing.T) {
	k := New()

	middlewares := []MiddlewareFunc{
		func(next HandlerFunc) HandlerFunc {
			return func(c *Context) error {
				c.Set("key1", 10)
				return next(c)
			}
		},
		func(next HandlerFunc) HandlerFunc {
			return func(c *Context) error {
				c.Set("key2", 20)
				return next(c)
			}
		},
	}

	handler := k.applyMiddlewaresToHandler(func(c *Context) error {
		val1, _ := c.Get("key1")
		val2, _ := c.Get("key2")
		return c.JSON(http.StatusOK, Map{"key1": val1, "key2": val2})
	}, middlewares...)

	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	res := httptest.NewRecorder()

	c := newContext(k)
	c.reset(req, res)

	err := handler(c)

	assert.NoError(t, err)
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

func TestKid_ServeHTTP_ErrorReturnedByHandler(t *testing.T) {
	k := New()

	k.Get("/test", func(c *Context) error {
		return errors.NewHTTPError(http.StatusForbidden)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	k.ServeHTTP(res, req)

	assert.Equal(t, http.StatusForbidden, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "{\"message\":\"Forbidden\"}\n", res.Body.String())
}

func TestKid_ServeHTTP_WriteStatusCodeIfNotWritten(t *testing.T) {
	k := New()

	k.Get("/test", func(c *Context) error {
		c.Response().WriteHeader(http.StatusCreated)
		return nil
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

	k.Get("/", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"message": "healthy"})
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

func TestResolveAddress(t *testing.T) {
	addr := resolveAddress([]string{})

	assert.Equal(t, ":2376", addr)

	addr = resolveAddress([]string{":2377", "2378"})
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
