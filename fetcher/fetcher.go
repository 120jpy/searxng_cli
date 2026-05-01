package fetcher

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	md "github.com/JohannesKaufmann/html-to-markdown"
)

func FetchURLs(urls []string, timeoutSec, concurrency int) map[string]string {
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
	return result
}

func fetchPage(browser *rod.Browser, rawURL string, timeoutSec int) string {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Sprintf("[error: invalid URL: %s]", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: parsed.String()})
	if err != nil {
		return fmt.Sprintf("[error: %s]", err)
	}
	defer page.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	page = page.Context(ctx)

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
			'aside', 'dialog',
			'nav', 'header', 'footer',
			'script', 'style', 'noscript', 'svg',
			'form', 'button', 'input', 'select', 'textarea'
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
