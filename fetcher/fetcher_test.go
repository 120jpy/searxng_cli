package fetcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCollectImageRefs(t *testing.T) {
	bodies := map[string]string{
		"https://example.com/page": "Some text ![alt](https://example.com/img.jpg) more text",
	}
	refs := collectImageRefs(bodies)
	if len(refs) != 1 {
		t.Fatalf("got %d refs, want 1", len(refs))
	}
	if refs[0].imageURL != "https://example.com/img.jpg" {
		t.Fatalf("imageURL = %q", refs[0].imageURL)
	}
	if refs[0].altText != "alt" {
		t.Fatalf("altText = %q", refs[0].altText)
	}
}

func TestCollectImageRefsRelativeURL(t *testing.T) {
	bodies := map[string]string{
		"https://example.com/page": "![](/images/foo.png)",
	}
	refs := collectImageRefs(bodies)
	if len(refs) != 1 {
		t.Fatalf("got %d refs, want 1", len(refs))
	}
	if refs[0].imageURL != "https://example.com/images/foo.png" {
		t.Fatalf("relative URL not resolved: %q", refs[0].imageURL)
	}
}

func TestCollectImageRefsDedup(t *testing.T) {
	bodies := map[string]string{
		"https://a.com/page1": "![](https://example.com/img.jpg)",
		"https://b.com/page2": "![](https://example.com/img.jpg)",
	}
	refs := collectImageRefs(bodies)
	if len(refs) != 1 {
		t.Fatalf("got %d refs, want 1 (dedup)", len(refs))
	}
}

func TestCollectImageRefsNoImages(t *testing.T) {
	bodies := map[string]string{
		"https://example.com": "just text no images",
	}
	refs := collectImageRefs(bodies)
	if len(refs) != 0 {
		t.Fatalf("got %d refs, want 0", len(refs))
	}
}

func TestDownloadAndReplace(t *testing.T) {
	imgContent := []byte("fake-image-binary-data")
	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(imgContent)
	}))
	defer imgServer.Close()

	bodies := map[string]string{
		"https://example.com/page": fmt.Sprintf("before ![alt](%s) after", imgServer.URL),
	}

	refs := collectImageRefs(bodies)
	if len(refs) != 1 {
		t.Fatalf("got %d refs, want 1", len(refs))
	}

	tempDir := t.TempDir()
	pathMap := downloadImages(refs, tempDir, 5, 1)
	replaceImageURLs(bodies, pathMap)

	body := bodies["https://example.com/page"]
	if !strings.Contains(body, tempDir) {
		t.Fatalf("expected local path in body, got: %s", body)
	}
	if strings.Contains(body, imgServer.URL) {
		t.Fatalf("remote URL should be replaced, got: %s", body)
	}
}

func TestDownloadFailure(t *testing.T) {
	bodies := map[string]string{
		"https://example.com/page": "![alt](http://not-a-real-server.example/nonexistent.jpg)",
	}

	refs := collectImageRefs(bodies)
	if len(refs) != 1 {
		t.Fatalf("got %d refs, want 1", len(refs))
	}

	tempDir := t.TempDir()
	pathMap := downloadImages(refs, tempDir, 1, 1)
	replaceImageURLs(bodies, pathMap)

	body := bodies["https://example.com/page"]
	if !strings.Contains(body, "[image failed:") {
		t.Fatalf("expected failure marker, got: %s", body)
	}
}

func TestExtFromContentType(t *testing.T) {
	tests := []struct {
		ct   string
		want string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/svg+xml", ".svg"},
		{"image/webp", ".webp"},
		{"image/x-icon", ".ico"},
		{"image/vnd.microsoft.icon", ".ico"},
		{"image/bmp", ".bmp"},
		{"application/octet-stream", ".bin"},
		{"", ".bin"},
	}
	for _, tt := range tests {
		got := extFromContentType(tt.ct)
		if got != tt.want {
			t.Errorf("extFromContentType(%q) = %q, want %q", tt.ct, got, tt.want)
		}
	}
}
