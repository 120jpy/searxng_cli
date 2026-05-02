package fetcher

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	md "github.com/JohannesKaufmann/html-to-markdown"
)

func FetchURLs(urls []string, timeoutSec, concurrency int) (map[string]string, func()) {
	result := make(map[string]string, len(urls))
	var mu sync.Mutex
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	launchURL := launcher.New().MustLaunch()
	browser := rod.New().ControlURL(launchURL).MustConnect()
	defer browser.Close()

	for _, u := range urls {
		wg.Add(1)
		sem <- struct{}{}
		go func(rawURL string) {
			defer wg.Done()
			defer func() { <-sem }()

			body := fetchPage(browser, rawURL, timeoutSec)
			mu.Lock()
			result[rawURL] = body
			mu.Unlock()
		}(u)
	}
	wg.Wait()

	cleanup := func() {}
	imageRefs := collectImageRefs(result)
	if len(imageRefs) > 0 {
		tempDir := filepath.Join(os.TempDir(), "searxng-cli-images", randomDirName())
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return result, cleanup
		}
		cleanup = func() { os.RemoveAll(tempDir) }

		pathMap := downloadImages(imageRefs, tempDir, timeoutSec, concurrency)
		replaceImageURLs(result, pathMap)
	}

	return result, cleanup
}

func fetchPage(browser *rod.Browser, rawURL string, timeoutSec int) string {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Sprintf("[error: invalid URL: %s]", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return fmt.Sprintf("[error: %s]", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	page = page.Context(ctx)
	defer page.Close()

	if err := page.Navigate(parsed.String()); err != nil {
		return fmt.Sprintf("[error: navigate: %s]", err)
	}
	if err := page.WaitLoad(); err != nil {
		return fmt.Sprintf("[error: waitload: %s]", err)
	}

	page.WaitRequestIdle(500*time.Millisecond, nil, nil, nil)()

	html, err := page.Eval(`() => {
		const selectors = [
			'[class*="ad"]', '[id*="ad"]',
			'[class*="sidebar"]', '[id*="sidebar"]',
			'[class*="nav"]', '[id*="nav"]',
			'[class*="menu"]', '[id*="menu"]',
			'[class*="footer"]', '[id*="footer"]',
			'[class*="cookie"]', '[id*="cookie"]',
			'[class*="widget"]', '[id*="widget"]',
			'[class*="social"]', '[id*="social"]',
			'[class*="share"]', '[id*="share"]',
			'[class*="tracking"]', '[id*="tracking"]',
			'[class*="analytics"]', '[id*="analytics"]',
			'[class*="pixel"]', '[id*="pixel"]',
			'[class*="banner"]', '[id*="banner"]',
			'[class*="skip"]', '[id*="skip"]',
			'[class*="accessibility"]', '[id*="accessibility"]',
			'[class*="popup"]', '[id*="popup"]',
			'[class*="overlay"]', '[id*="overlay"]',
			'[class*="modal"]', '[id*="modal"]',
			'aside', 'dialog',
			'nav', 'header', 'footer',
			'script', 'style', 'noscript', 'svg', 'canvas',
			'form', 'button', 'input', 'select', 'textarea',
			'iframe',
			'img[src*="pixel"]', 'img[src*="track"]',
			'img[src*="bat.bing"]', 'img[src*="doubleclick"]',
			'img[height="1"]', 'img[width="1"]',
		];
		for (const sel of selectors) {
			for (const el of document.querySelectorAll(sel)) {
				if (el.tagName === 'BODY') continue;
				if (el.querySelector('article, main, [role="main"]')) {
					el.removeAttribute('class');
					el.removeAttribute('id');
					continue;
				}
				el.remove();
			}
		}
		for (const el of document.querySelectorAll('*')) {
			const s = window.getComputedStyle(el);
			if (el.tagName === 'BODY') continue;
			if (s.display === 'none' || s.visibility === 'hidden') {
				if (el.querySelector('article, main, [role="main"]')) {
					el.removeAttribute('class');
					el.removeAttribute('id');
					continue;
				}
				el.remove();
			}
		}
		return document.body.innerHTML;
	}`)
	if err != nil {
		return fmt.Sprintf("[error: %s]", err)
	}

	rawHTML := html.Value.Str()
	if strings.TrimSpace(rawHTML) == "" {
		return "[error: empty page]"
	}

	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(rawHTML)
	if err != nil {
		return fmt.Sprintf("[error: markdown conversion: %s]", err)
	}

	return strings.TrimSpace(markdown)
}

var imageRe = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

type imageRef struct {
	pageURL  string
	imageURL string
	altText  string
}

func collectImageRefs(bodies map[string]string) []imageRef {
	var refs []imageRef
	seen := make(map[string]bool)
	for pageURL, body := range bodies {
		base, err := url.Parse(pageURL)
		if err != nil {
			continue
		}
		matches := imageRe.FindAllStringSubmatch(body, -1)
		for _, m := range matches {
			altText := m[1]
			imgURL := m[2]
			parsed, err := url.Parse(imgURL)
			if err != nil {
				continue
			}
			resolved := base.ResolveReference(parsed)
			absURL := resolved.String()
			if seen[absURL] {
				continue
			}
			seen[absURL] = true
			refs = append(refs, imageRef{
				pageURL:  pageURL,
				imageURL: absURL,
				altText:  altText,
			})
		}
	}
	return refs
}

func downloadImages(refs []imageRef, tempDir string, timeoutSec, concurrency int) map[string]string {
	pathMap := make(map[string]string, len(refs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

	for _, ref := range refs {
		wg.Add(1)
		sem <- struct{}{}
		go func(ref imageRef) {
			defer wg.Done()
			defer func() { <-sem }()

			resp, err := client.Get(ref.imageURL)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return
			}

			contentType := resp.Header.Get("Content-Type")
			ext := extFromContentType(contentType)
			hash := sha256.Sum256([]byte(ref.imageURL))
			filename := hex.EncodeToString(hash[:]) + ext
			path := filepath.Join(tempDir, filename)

			f, err := os.Create(path)
			if err != nil {
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, resp.Body); err != nil {
				f.Close()
				os.Remove(path)
				return
			}

			mu.Lock()
			pathMap[ref.imageURL] = path
			mu.Unlock()
		}(ref)
	}
	wg.Wait()
	return pathMap
}

func replaceImageURLs(bodies map[string]string, pathMap map[string]string) {
	for pageURL, body := range bodies {
		base, _ := url.Parse(pageURL)
		replaced := imageRe.ReplaceAllStringFunc(body, func(match string) string {
			parts := imageRe.FindStringSubmatch(match)
			if len(parts) < 3 {
				return match
			}
			altText := parts[1]
			imgURL := parts[2]

			parsed, err := url.Parse(imgURL)
			if err != nil {
				return match
			}
			resolved := base.ResolveReference(parsed)
			absURL := resolved.String()

			if localPath, ok := pathMap[absURL]; ok {
				return fmt.Sprintf("![%s](%s)", altText, localPath)
			}
			return fmt.Sprintf("[image failed: %s]", absURL)
		})
		bodies[pageURL] = replaced
	}
}

func extFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "image/jpeg"):
		return ".jpg"
	case strings.Contains(ct, "image/png"):
		return ".png"
	case strings.Contains(ct, "image/gif"):
		return ".gif"
	case strings.Contains(ct, "image/svg"):
		return ".svg"
	case strings.Contains(ct, "image/webp"):
		return ".webp"
	case strings.Contains(ct, "image/x-icon"), strings.Contains(ct, "image/vnd.microsoft.icon"):
		return ".ico"
	case strings.Contains(ct, "image/bmp"):
		return ".bmp"
	default:
		return ".bin"
	}
}

func randomDirName() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
