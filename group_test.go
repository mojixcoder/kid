package kid

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func registerHandlers(g Group) {
	g.Get("/path", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	g.Post("/path", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	g.Patch("/path", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	g.Put("/path", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	g.Delete("/path", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	g.Connect("/path", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	g.Trace("/path", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	g.Options("/path", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	g.Head("/path", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})

	g.Any("/any", func(c *Context) error {
		return c.JSON(http.StatusOK, Map{"method": c.Request().Method})
	})
}

func TestNewGroup(t *testing.T) {
	k := New()

	g := newGroup(k, "/v1")
	assert.Equal(t, k, g.kid)
	assert.Equal(t, "/v1", g.prefix)
	assert.Nil(t, g.middlewares)

	g = newGroup(k, "/v1", nil)
	assert.NotNil(t, g.middlewares)
	assert.Len(t, g.middlewares, 1)
}

func TestGroup_combineMiddlewares(t *testing.T) {
	k := New()

	g := newGroup(k, "/v1")

	middlewares := g.combineMiddlewares(nil)
	assert.Nil(t, middlewares)
	assert.Len(t, middlewares, 0)

	middlewares = g.combineMiddlewares([]MiddlewareFunc{nil})
	assert.NotNil(t, middlewares)
	assert.Len(t, middlewares, 1)

	g.middlewares = []MiddlewareFunc{nil, nil}

	middlewares = g.combineMiddlewares(nil)
	assert.NotNil(t, middlewares)
	assert.Len(t, middlewares, 2)

	middlewares = g.combineMiddlewares([]MiddlewareFunc{nil})
	assert.NotNil(t, middlewares)
	assert.Len(t, middlewares, 3)
}

func TestGroup_Add(t *testing.T) {
	k := New()
	g := newGroup(k, "/v1")

	assert.PanicsWithValue(t, "handler cannot be nil", func() {
		g.Add("/", nil, []string{http.MethodGet, http.MethodPost})
	})

	g.Add("/test", func(c *Context) error {
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
		{req: httptest.NewRequest(http.MethodPost, "/v1/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodPost},
		{req: httptest.NewRequest(http.MethodGet, "/v1/test", nil), res: httptest.NewRecorder(), expectedMethod: http.MethodGet},
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

func TestGroup_Methods(t *testing.T) {
	k := New()
	g := newGroup(k, "/v1")

	registerHandlers(g)

	target := "/v1/path"
	any := "/v1/any"

	testCases := []struct {
		req            *http.Request
		res            *httptest.ResponseRecorder
		expectedMethod string
	}{
		{req: httptest.NewRequest(http.MethodGet, target, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodGet},
		{req: httptest.NewRequest(http.MethodPost, target, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodPost},
		{req: httptest.NewRequest(http.MethodPut, target, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodPut},
		{req: httptest.NewRequest(http.MethodPatch, target, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodPatch},
		{req: httptest.NewRequest(http.MethodDelete, target, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodDelete},
		{req: httptest.NewRequest(http.MethodOptions, target, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodOptions},
		{req: httptest.NewRequest(http.MethodConnect, target, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodConnect},
		{req: httptest.NewRequest(http.MethodTrace, target, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodTrace},
		{req: httptest.NewRequest(http.MethodHead, target, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodHead},

		{req: httptest.NewRequest(http.MethodGet, any, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodGet},
		{req: httptest.NewRequest(http.MethodDelete, any, nil), res: httptest.NewRecorder(), expectedMethod: http.MethodDelete},
	}

	for _, testCase := range testCases {
		t.Run(testCase.expectedMethod, func(t *testing.T) {
			k.ServeHTTP(testCase.res, testCase.req)

			assert.Equal(t, http.StatusOK, testCase.res.Code)
			assert.Equal(t, "application/json", testCase.res.Header().Get("Content-Type"))
			assert.Equal(t, fmt.Sprintf("{\"method\":%q}\n", testCase.expectedMethod), testCase.res.Body.String())
		})
	}
}

func TestGroup_Group(t *testing.T) {
	k := New()

	g := newGroup(k, "/v1")

	nestedG := g.Group("/api")
	assert.Equal(t, k, nestedG.kid)
	assert.Equal(t, "/v1/api", nestedG.prefix)
	assert.Nil(t, nestedG.middlewares)

	nestedG = g.Group("/api", nil)
	assert.NotNil(t, nestedG.middlewares)
	assert.Len(t, nestedG.middlewares, 1)
}

func TestGroup_Add_NestedGroups(t *testing.T) {
	k := New()
	g := newGroup(k, "/v1")
	nestedG := g.Group("/api")

	g.Add("/test", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": c.Request().Method})
	}, []string{http.MethodPost})

	nestedG.Add("/{var}", func(c *Context) error {
		return c.JSON(http.StatusCreated, Map{"message": c.Param("var")})
	}, []string{http.MethodPost})

	testCases := []struct {
		req                   *http.Request
		res                   *httptest.ResponseRecorder
		name, expectedMessage string
	}{
		{
			name:            "group",
			req:             httptest.NewRequest(http.MethodPost, "/v1/test", nil),
			res:             httptest.NewRecorder(),
			expectedMessage: "{\"message\":\"POST\"}\n",
		},
		{
			name:            "nested_group",
			req:             httptest.NewRequest(http.MethodPost, "/v1/api/test", nil),
			res:             httptest.NewRecorder(),
			expectedMessage: "{\"message\":\"test\"}\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			k.ServeHTTP(testCase.res, testCase.req)

			assert.Equal(t, http.StatusCreated, testCase.res.Code)
			assert.Equal(t, "application/json", testCase.res.Header().Get("Content-Type"))
			assert.Equal(t, testCase.expectedMessage, testCase.res.Body.String())
		})
	}
}
