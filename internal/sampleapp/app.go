// Package sampleapp contains the small Gin application exercised by the E2E tests.
package sampleapp

import (
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"html/template"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*.html
var templates embed.FS

type session struct {
	csrf  string
	email string
}

// New creates an application with isolated, in-memory sample sessions.
func New() *gin.Engine {
	var mu sync.Mutex
	sessions := make(map[string]session)
	find := func(c *gin.Context) (string, session) {
		id, _ := c.Cookie("sample_session")
		mu.Lock()
		defer mu.Unlock()
		return id, sessions[id]
	}
	save := func(c *gin.Context, value session, oldID string) {
		id := rand.Text()
		mu.Lock()
		delete(sessions, oldID)
		sessions[id] = value
		mu.Unlock()
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("sample_session", id, 3600, "/", "", false, true)
	}

	app := gin.New()
	app.Use(gin.Recovery())
	app.SetHTMLTemplate(template.Must(template.ParseFS(templates, "templates/*.html")))
	app.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	app.GET("/login", func(c *gin.Context) {
		id, value := find(c)
		if value.csrf == "" {
			value.csrf = rand.Text()
			save(c, value, id)
		}
		c.HTML(http.StatusOK, "login.html", gin.H{"CSRF": value.csrf})
	})
	app.POST("/session", func(c *gin.Context) {
		id, value := find(c)
		if value.csrf == "" || subtle.ConstantTimeCompare([]byte(c.PostForm("authenticity_token")), []byte(value.csrf)) != 1 {
			c.HTML(http.StatusForbidden, "login.html", gin.H{"Error": "フォームの有効期限が切れています。もう一度ログインしてください"})
			return
		}
		// A fixed sample user keeps the demo independent of a database.
		if c.PostForm("user[email]") != "tester@example.com" || c.PostForm("user[password]") != "sample-password" {
			c.HTML(http.StatusUnprocessableEntity, "login.html", gin.H{
				"CSRF": value.csrf, "Error": "メールアドレスまたはパスワードが違います",
			})
			return
		}
		// Rotate the session ID on login and discard the pre-login session.
		save(c, session{csrf: rand.Text(), email: "tester@example.com"}, id)
		c.Redirect(http.StatusSeeOther, "/account")
	})
	app.GET("/account", func(c *gin.Context) {
		_, value := find(c)
		if value.email == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}
		c.HTML(http.StatusOK, "account.html", gin.H{"Email": value.email})
	})
	return app
}
