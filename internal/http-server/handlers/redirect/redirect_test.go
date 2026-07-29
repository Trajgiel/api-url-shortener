package redirect

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	resp "github.com/Trajgiel/api-url-shortener/internal/lib/api/response"
	"github.com/Trajgiel/api-url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

type urlGetterMock struct {
	getURL func(alias string) (string, error)
}

func (m *urlGetterMock) GetURL(alias string) (string, error) {
	return m.getURL(alias)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name          string
		alias         string
		getURL        func(alias string) (string, error)
		wantStatus    string
		wantErrSubstr string
		wantRedirect  string
	}{
		{
			name:  "success",
			alias: "google",
			getURL: func(alias string) (string, error) {
				return "https://google.com", nil
			},
			wantRedirect: "https://google.com",
		},
		{
			name:          "empty alias",
			alias:         "",
			wantStatus:    resp.StatusError,
			wantErrSubstr: "invalid request",
		},
		{
			name:  "url not found",
			alias: "unknown",
			getURL: func(alias string) (string, error) {
				return "", storage.ErrURLNotFound
			},
			wantStatus:    resp.StatusError,
			wantErrSubstr: "not found",
		},
		{
			name:  "unexpected storage error",
			alias: "google",
			getURL: func(alias string) (string, error) {
				return "", errors.New("unexpected error")
			},
			wantStatus:    resp.StatusError,
			wantErrSubstr: "internal error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getterMock := &urlGetterMock{getURL: tc.getURL}

			handler := New(discardLogger(), getterMock)

			req := httptest.NewRequest(http.MethodGet, "/"+tc.alias, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if tc.wantRedirect != "" {
				if rr.Code != http.StatusFound {
					t.Errorf("expected status %d, got %d", http.StatusFound, rr.Code)
				}

				if got := rr.Header().Get("Location"); got != tc.wantRedirect {
					t.Errorf("expected redirect location %q, got %q", tc.wantRedirect, got)
				}

				return
			}

			var got resp.Response
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal response body %q: %v", rr.Body.String(), err)
			}

			if got.Status != tc.wantStatus {
				t.Errorf("expected status %q, got %q", tc.wantStatus, got.Status)
			}

			if tc.wantErrSubstr != "" && got.Error != tc.wantErrSubstr {
				t.Errorf("expected error %q, got %q", tc.wantErrSubstr, got.Error)
			}
		})
	}
}
