package save

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	resp "github.com/Trajgiel/api-url-shortener/internal/lib/api/response"
	"github.com/Trajgiel/api-url-shortener/internal/storage"
)

type urlSaverMock struct {
	saveURL func(urlToSave string, alias string) (int64, error)
}

func (m *urlSaverMock) SaveURL(urlToSave string, alias string) (int64, error) {
	return m.saveURL(urlToSave, alias)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name          string
		body          string
		saveURL       func(urlToSave string, alias string) (int64, error)
		wantStatus    string
		wantErrSubstr string
		wantAlias     string
	}{
		{
			name: "success with custom alias",
			body: `{"url": "https://google.com", "alias": "google"}`,
			saveURL: func(urlToSave string, alias string) (int64, error) {
				return 1, nil
			},
			wantStatus: resp.StatusOk,
			wantAlias:  "google",
		},
		{
			name: "success without alias generates random one",
			body: `{"url": "https://google.com"}`,
			saveURL: func(urlToSave string, alias string) (int64, error) {
				if alias == "" {
					t.Errorf("expected generated alias to be passed to SaveURL, got empty string")
				}
				return 1, nil
			},
			wantStatus: resp.StatusOk,
		},
		{
			name:          "empty request body",
			body:          ``,
			wantStatus:    resp.StatusError,
			wantErrSubstr: "failed to decode request",
		},
		{
			name:          "invalid json",
			body:          `{"url":`,
			wantStatus:    resp.StatusError,
			wantErrSubstr: "failed to decode request",
		},
		{
			name:          "missing url",
			body:          `{"alias": "test"}`,
			wantStatus:    resp.StatusError,
			wantErrSubstr: "is a required field",
		},
		{
			name:          "invalid url",
			body:          `{"url": "not-a-url"}`,
			wantStatus:    resp.StatusError,
			wantErrSubstr: "is invalid url",
		},
		{
			name: "url already exists",
			body: `{"url": "https://google.com", "alias": "google"}`,
			saveURL: func(urlToSave string, alias string) (int64, error) {
				return 0, storage.ErrURLExists
			},
			wantStatus:    resp.StatusError,
			wantErrSubstr: "url already exists",
		},
		{
			name: "unexpected storage error",
			body: `{"url": "https://google.com", "alias": "google"}`,
			saveURL: func(urlToSave string, alias string) (int64, error) {
				return 0, errors.New("unexpected error")
			},
			wantStatus:    resp.StatusError,
			wantErrSubstr: "failed to save url",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			saver := &urlSaverMock{saveURL: tc.saveURL}

			handler := New(discardLogger(), saver)

			req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			var got Response
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal response body %q: %v", rr.Body.String(), err)
			}

			if got.Status != tc.wantStatus {
				t.Errorf("expected status %q, got %q", tc.wantStatus, got.Status)
			}

			if tc.wantErrSubstr != "" && !strings.Contains(got.Error, tc.wantErrSubstr) {
				t.Errorf("expected error to contain %q, got %q", tc.wantErrSubstr, got.Error)
			}

			if tc.wantAlias != "" && got.Alias != tc.wantAlias {
				t.Errorf("expected alias %q, got %q", tc.wantAlias, got.Alias)
			}
		})
	}
}
