package e2e_test

import (
	"fmt"
	"net/http"

	"github.com/gocolly/colly/v2"
)

// loginScenario is private test state, created once per Godog scenario.
type loginScenario struct {
	baseURL   string
	email     string
	password  string
	collector *colly.Collector
	page      *page
}

func (s *loginScenario) openLogin() error {
	if err := s.collector.Visit(s.baseURL + "/login"); err != nil {
		return err
	}
	return s.expectLogin()
}

func (s *loginScenario) login(credentials string) error {
	action, fields, err := loginForm(s.page)
	if err != nil {
		return err
	}
	email, password := s.email, s.password
	switch credentials {
	case "正しい認証情報":
	case "誤ったパスワード":
		password += "-invalid"
	case "未登録のメールアドレス":
		email = "unknown@example.invalid"
	case "空のパスワード":
		password = ""
	default:
		return fmt.Errorf("未対応の認証条件: %q", credentials)
	}
	fields["user[email]"], fields["user[password]"] = email, password
	return s.collector.Post(action, fields)
}

func (s *loginScenario) expectAccount() error {
	return expectPage(s.page, http.StatusOK, "/account", "[data-testid=current-user]", s.email)
}

func (s *loginScenario) visitAccount() error {
	return s.collector.Visit(s.baseURL + "/account")
}

func (s *loginScenario) reopenAccount() error {
	if err := s.visitAccount(); err != nil {
		return err
	}
	return s.expectAccount()
}

func (s *loginScenario) expectAuthenticationError() error {
	return expectPage(s.page, http.StatusUnprocessableEntity, "/session", "[role=alert]", "メールアドレスまたはパスワードが違います")
}

func (s *loginScenario) expectLogin() error {
	return expectPage(s.page, http.StatusOK, "/login", "h1", "ログイン")
}

func (s *loginScenario) expectLoginRequired() error {
	if err := s.visitAccount(); err != nil {
		return err
	}
	return s.expectLogin()
}

func (s *loginScenario) loginWithInvalidCSRF(condition string) error {
	action, fields, err := loginForm(s.page)
	if err != nil {
		return err
	}
	fields["user[email]"], fields["user[password]"] = s.email, s.password
	switch condition {
	case "CSRFトークンなし":
		delete(fields, "authenticity_token")
	case "不正なCSRFトークン":
		fields["authenticity_token"] = "forged-token"
	default:
		return fmt.Errorf("未対応の CSRF 条件: %q", condition)
	}
	return s.collector.Post(action, fields)
}

func (s *loginScenario) expectCSRFError() error {
	return expectPage(s.page, http.StatusForbidden, "/session", "[role=alert]", "フォームの有効期限が切れています。もう一度ログインしてください")
}
