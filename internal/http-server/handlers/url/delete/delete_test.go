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
	deleteURL func(id int64) error
}

func (m *urlDeleterMock) DeleteURL(id int64) error {
	return m.deleteURL(id)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name          string
		id            string
		deleteURL     func(id int64) error
		wantStatus    string
		wantErrSubstr string
	}{
		{
			name: "success",
			id:   "1",
			deleteURL: func(id int64) error {
				return nil
			},
			wantStatus: resp.StatusOk,
		},
		{
			name:          "empty id",
			id:            "",
			wantStatus:    resp.StatusError,
			wantErrSubstr: "invalid request",
		},
		{
			name:          "non-numeric id",
			id:            "abc",
			wantStatus:    resp.StatusError,
			wantErrSubstr: "invalid request",
		},
		{
			name: "url not found",
			id:   "42",
			deleteURL: func(id int64) error {
				return storage.ErrURLNotFound
			},
			wantStatus:    resp.StatusError,
			wantErrSubstr: "not found",
		},
		{
			name: "unexpected storage error",
			id:   "1",
			deleteURL: func(id int64) error {
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

			req := httptest.NewRequest(http.MethodDelete, "/"+tc.id, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)
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
