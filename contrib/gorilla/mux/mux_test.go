package mux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"

	"github.com/stretchr/testify/assert"
)

func TestHttpTracer(t *testing.T) {
	for _, ht := range []struct {
		code         int
		method       string
		url          string
		resourceName string
		errorStr     string
	}{
		{
			code:         http.StatusOK,
			method:       "GET",
			url:          "/200",
			resourceName: "GET /200",
		},
		{
			code:         http.StatusNotFound,
			method:       "GET",
			url:          "/not_a_real_route",
			resourceName: "GET unknown",
		},
		{
			code:         http.StatusMethodNotAllowed,
			method:       "POST",
			url:          "/405",
			resourceName: "POST unknown",
		},
		{
			code:         http.StatusInternalServerError,
			method:       "GET",
			url:          "/500",
			resourceName: "GET /500",
			errorStr:     "500: Internal Server Error",
		},
	} {
		t.Run(http.StatusText(ht.code), func(t *testing.T) {
			assert := assert.New(t)
			mt := mocktracer.Start()
			defer mt.Stop()
			codeStr := strconv.Itoa(ht.code)

			// Send and verify a request
			r := httptest.NewRequest(ht.method, ht.url, nil)
			w := httptest.NewRecorder()
			router().ServeHTTP(w, r)
			assert.Equal(ht.code, w.Code)
			assert.Equal(codeStr+"!\n", w.Body.String())

			spans := mt.FinishedSpans()
			assert.Equal(1, len(spans))

			s := spans[0]
			assert.Equal("http.request", s.OperationName())
			assert.Equal("my-service", s.Tag(ext.ServiceName))
			assert.Equal(codeStr, s.Tag(ext.HTTPCode))
			assert.Equal(ht.method, s.Tag(ext.HTTPMethod))
			assert.Equal(ht.url, s.Tag(ext.HTTPURL))
			assert.Equal(ht.resourceName, s.Tag(ext.ResourceName))
			if ht.errorStr != "" {
				assert.Equal(ht.errorStr, s.Tag(ext.Error).(error).Error())
			}
		})
	}
}

func router() http.Handler {
	mux := NewRouter(WithServiceName("my-service"))
	mux.Handle("/200", okHandler())
	mux.Handle("/500", errorHandler(http.StatusInternalServerError))
	mux.Handle("/405", okHandler()).Methods("GET")
	mux.NotFoundHandler = errorHandler(http.StatusNotFound)
	mux.MethodNotAllowedHandler = errorHandler(http.StatusMethodNotAllowed)
	return mux
}

func errorHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("%d!", code), code)
	})
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("200!\n"))
	})
}
