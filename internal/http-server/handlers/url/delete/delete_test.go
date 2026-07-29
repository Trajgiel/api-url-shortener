package delete

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

type urlDeleterMock struct {
	deleteURL func(alias string) error
}

func (m *urlDeleterMock) DeleteURL(alias string) error {
	return m.deleteURL(alias)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name          string
		alias         string
		deleteURL     func(alias string) error
		wantStatus    string
		wantErrSubstr string
	}{
		{
			name:  "success",
			alias: "google",
			deleteURL: func(alias string) error {
				return nil
			},
			wantStatus: resp.StatusOk,
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
			deleteURL: func(alias string) error {
				return storage.ErrURLNotFound
			},
			wantStatus:    resp.StatusError,
			wantErrSubstr: "not found",
		},
		{
			name:  "unexpected storage error",
			alias: "google",
			deleteURL: func(alias string) error {
				return errors.New("unexpected error")
			},
			wantStatus:    resp.StatusError,
			wantErrSubstr: "internal error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deleterMock := &urlDeleterMock{deleteURL: tc.deleteURL}

			handler := New(discardLogger(), deleterMock)

			req := httptest.NewRequest(http.MethodDelete, "/"+tc.alias, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

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
