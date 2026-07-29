package tests

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

const (
	host          = "localhost:8082"
	basicAuthUser = "user"
	basicAuthPass = "password"
)

func newExpect(t *testing.T) *httpexpect.Expect {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://" + host,
		Client:   client,
		Reporter: httpexpect.NewAssertReporter(t),
	})
}

func TestURLShortener_SaveRedirectDelete_HappyPath(t *testing.T) {
	e := newExpect(t)

	const originalURL = "https://example.com/some-page"

	body := e.POST("/url").
		WithJSON(map[string]any{"url": originalURL}).
		WithBasicAuth(basicAuthUser, basicAuthPass).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	body.Value("status").String().IsEqual("OK")
	alias := body.Value("alias").String().NotEmpty().Raw()

	e.GET("/" + alias).
		Expect().
		Status(http.StatusFound).
		Header("Location").IsEqual(originalURL)

	e.DELETE("/url/{alias}").
		WithPath("alias", alias).
		WithBasicAuth(basicAuthUser, basicAuthPass).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("status").String().IsEqual("OK")

	e.GET("/" + alias).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("error").String().IsEqual("not found")
}

func TestURLShortener_Save_InvalidURL(t *testing.T) {
	e := newExpect(t)

	e.POST("/url").
		WithJSON(map[string]any{"url": "not-a-url"}).
		WithBasicAuth(basicAuthUser, basicAuthPass).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("error").String().Contains("invalid url")
}

func TestURLShortener_Save_NoAuth(t *testing.T) {
	e := newExpect(t)

	e.POST("/url").
		WithJSON(map[string]any{"url": "https://example.com"}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestURLShortener_Delete_NotFound(t *testing.T) {
	e := newExpect(t)

	e.DELETE("/url/{alias}").
		WithPath("alias", "nonexistent-alias-xyz").
		WithBasicAuth(basicAuthUser, basicAuthPass).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("error").String().IsEqual("not found")
}
