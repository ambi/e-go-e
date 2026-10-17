package e2e_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ambi/e-go-e/internal/sampleapp"
	"github.com/gocolly/colly/v2"
)

// target starts a real HTTP listener for the sample unless a running app is
// specified. An unavailable explicit target fails the tests without fallback.
func target(t *testing.T) (baseURL, email, password string) {
	t.Helper()
	baseURL = os.Getenv("E2E_BASE_URL")
	email, password = "tester@example.com", "sample-password"
	if baseURL == "" {
		server := httptest.NewServer(sampleapp.New())
		t.Cleanup(server.Close)
		return server.URL, email, password
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		t.Fatal("E2E_BASE_URL に絶対 HTTP(S) URL を指定してください")
	}
	if email = os.Getenv("E2E_EMAIL"); email == "" {
		t.Fatal("E2E_BASE_URL を指定する場合は E2E_EMAIL が必要です")
	}
	if password = os.Getenv("E2E_PASSWORD"); password == "" {
		t.Fatal("E2E_BASE_URL を指定する場合は E2E_PASSWORD が必要です")
	}
	return strings.TrimRight(baseURL, "/"), email, password
}

// page records the most recent response, including the final redirect URL.
type page struct {
	status int
	path   string
	html   *colly.HTMLElement
}

func newCollector(ctx context.Context) (*colly.Collector, *page) {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.ParseHTTPErrorResponse(),
		colly.IgnoreRobotsTxt(),
		colly.StdlibContext(ctx),
	)
	c.SetRequestTimeout(10 * time.Second)
	current := &page{}
	c.OnResponse(func(r *colly.Response) {
		*current = page{status: r.StatusCode, path: r.Request.URL.Path}
	})
	c.OnHTML("html", func(e *colly.HTMLElement) { current.html = e })
	return c, current
}

func expectPage(current *page, status int, path, selector, text string) error {
	if current.status != status || current.path != path {
		return fmt.Errorf("画面の遷移: 期待 HTTP %d %s、実際 HTTP %d %s", status, path, current.status, current.path)
	}
	if current.html == nil {
		return fmt.Errorf("HTML の画面が返されませんでした")
	}
	matches := current.html.DOM.Find(selector)
	if matches.Length() != 1 {
		return fmt.Errorf("%q の要素数: 期待 1、実際 %d", selector, matches.Length())
	}
	got := strings.Join(strings.Fields(matches.Text()), " ")
	if got != text {
		return fmt.Errorf("%q の表示: 期待 %q、実際 %q", selector, text, got)
	}
	return nil
}

// loginForm is specific to the sample login page, not a general form engine.
// Colly supplies the HTML helpers and resolves the form's action URL.
func loginForm(current *page) (string, map[string]string, error) {
	if current.html == nil || current.html.DOM.Find("form#login").Length() != 1 {
		return "", nil, fmt.Errorf("ログインフォームが見つかりません")
	}
	fields := make(map[string]string)
	current.html.ForEach("form#login input[type=hidden][name]", func(_ int, e *colly.HTMLElement) {
		fields[e.Attr("name")] = e.Attr("value")
	})
	if fields["authenticity_token"] == "" {
		return "", nil, fmt.Errorf("ログイン画面の CSRF トークンが空です")
	}
	action := current.html.ChildAttr("form#login", "action")
	return current.html.Request.AbsoluteURL(action), fields, nil
}
