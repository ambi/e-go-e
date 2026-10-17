package e2e_test

import (
	"bytes"
	"context"
	"embed"
	"os"
	"strconv"
	"testing"

	"github.com/cucumber/godog"
)

//go:embed features/*.feature
var featureFiles embed.FS

func TestFeatures(t *testing.T) {
	concurrency := 4
	if value := os.Getenv("E2E_CONCURRENCY"); value != "" {
		var err error
		concurrency, err = strconv.Atoi(value)
		if err != nil || concurrency < 1 {
			t.Fatal("E2E_CONCURRENCY に 1 以上の整数を指定してください")
		}
	}
	format := os.Getenv("E2E_FORMAT")
	if format == "" {
		format = "progress"
		if testing.Verbose() {
			format = "pretty"
		}
	}
	baseURL, email, password := target(t)
	// Keep Go's concurrent subtest notifications out of the Gherkin text.
	var output bytes.Buffer
	defer func() { t.Log("\n" + output.String()) }()
	suite := godog.TestSuite{
		Name: "e-go-e",
		Options: &godog.Options{
			FS: featureFiles, Paths: []string{"features"},
			TestingT: t, DefaultContext: t.Context(),
			Concurrency: concurrency, Strict: true, NoColors: true,
			Format: format, Tags: os.Getenv("E2E_TAGS"), Output: &output,
		},
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			// Godog calls this initializer separately for every scenario,
			// including each row of a Scenario Outline.
			state := &loginScenario{baseURL: baseURL, email: email, password: password}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				state.collector, state.page = newCollector(ctx)
				return ctx, nil
			})
			sc.Step(`^ログイン画面を開いている$`, state.openLogin)
			sc.Step(`^(正しい認証情報|誤ったパスワード|未登録のメールアドレス|空のパスワード)でログインする$`, state.login)
			sc.Step(`^マイページに利用者のメールアドレスが表示される$`, state.expectAccount)
			sc.Step(`^ページを開き直してもログイン状態が維持される$`, state.reopenAccount)
			sc.Step(`^認証エラーメッセージが表示される$`, state.expectAuthenticationError)
			sc.Step(`^マイページを開くとログイン画面へ戻される$`, state.expectLoginRequired)
			sc.Step(`^未ログインの状態でマイページを開く$`, state.visitAccount)
			sc.Step(`^ログイン画面が表示される$`, state.expectLogin)
			sc.Step(`^(CSRFトークンなし|不正なCSRFトークン)で正しい認証情報を送信する$`, state.loginWithInvalidCSRF)
			sc.Step(`^フォームの送信が拒否される$`, state.expectCSRFError)
		},
	}
	if suite.Run() != 0 {
		t.Fatal("Godog のシナリオが失敗しました")
	}
}
