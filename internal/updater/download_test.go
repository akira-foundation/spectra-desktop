package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDownload_WritesBodyToTempFile(t *testing.T) {
	want := []byte("hello update payload")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(want)
	}))
	t.Cleanup(srv.Close)

	path, err := download(context.Background(), srv.URL, ".zip", nil)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	if !strings.HasSuffix(path, ".zip") {
		t.Fatalf("temp file %q missing .zip suffix", path)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("body mismatch: %q vs %q", got, want)
	}
}

func TestDownload_Sha256Match(t *testing.T) {
	payload := []byte("payload-for-hash")
	sum := sha256.Sum256(payload)
	wantHex := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	path, err := download(context.Background(), srv.URL, ".bin", nil)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	gotSum := sha256.Sum256(got)
	if hex.EncodeToString(gotSum[:]) != wantHex {
		t.Fatalf("sha256 mismatch: got %s want %s", hex.EncodeToString(gotSum[:]), wantHex)
	}
}

func TestDownload_ProgressInvoked(t *testing.T) {
	payload := make([]byte, 256*1024)
	for i := range payload {
		payload[i] = byte(i % 251)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "262144")
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	var calls atomic.Int64
	var lastDownloaded, lastTotal atomic.Int64
	progress := func(downloaded, total int64) {
		calls.Add(1)
		lastDownloaded.Store(downloaded)
		lastTotal.Store(total)
	}

	path, err := download(context.Background(), srv.URL, ".bin", progress)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	if calls.Load() == 0 {
		t.Fatal("progress was never called")
	}
	if lastDownloaded.Load() != int64(len(payload)) {
		t.Fatalf("final downloaded = %d, want %d", lastDownloaded.Load(), len(payload))
	}
	if lastTotal.Load() != int64(len(payload)) {
		t.Fatalf("final total = %d, want %d", lastTotal.Load(), len(payload))
	}
}

func TestDownload_Non200ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	_, err := download(context.Background(), srv.URL, ".zip", nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "http 404") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDownload_PartialDownloadErrorsAndCleansUp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10000")
		_, _ = w.Write([]byte("partial"))
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, err := hj.Hijack()
			if err == nil {
				_ = conn.Close()
			}
		}
	}))
	t.Cleanup(srv.Close)

	_, err := download(context.Background(), srv.URL, ".zip", nil)
	if err == nil {
		t.Fatal("expected error for truncated response")
	}
}

func TestDownload_InvalidURLReturnsError(t *testing.T) {
	if _, err := download(context.Background(), "http://127.0.0.1:1/x", ".zip", nil); err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestStageAssets_DownloadsZipAndSig(t *testing.T) {
	payload := []byte("zip-bytes")
	sig := []byte("sig-bytes")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".sig") {
			_, _ = w.Write(sig)
			return
		}
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	info := &UpdateInfo{URL: srv.URL + "/app.zip"}
	zipPath, sigPath, err := stageAssets(context.Background(), info, nil)
	if err != nil {
		t.Fatalf("stageAssets: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(zipPath)
		_ = os.Remove(sigPath)
	})

	gotZip, _ := os.ReadFile(zipPath)
	gotSig, _ := os.ReadFile(sigPath)
	if string(gotZip) != string(payload) {
		t.Fatalf("zip body = %q, want %q", gotZip, payload)
	}
	if string(gotSig) != string(sig) {
		t.Fatalf("sig body = %q, want %q", gotSig, sig)
	}
	if !strings.HasSuffix(sigPath, ".sig") {
		t.Fatalf("sig path %q missing .sig suffix", sigPath)
	}
	if !strings.HasSuffix(zipPath, ".zip") {
		t.Fatalf("zip path %q missing .zip suffix", zipPath)
	}
}

func TestStageAssets_SigFailureRemovesZip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".sig") {
			http.Error(w, "no sig", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte("payload"))
	}))
	t.Cleanup(srv.Close)

	info := &UpdateInfo{URL: srv.URL + "/app.zip"}
	_, _, err := stageAssets(context.Background(), info, nil)
	if err == nil {
		t.Fatal("expected stageAssets to fail when sig is missing")
	}
}
