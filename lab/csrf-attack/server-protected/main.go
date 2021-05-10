package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spongeprojects/magicconch"
	"html/template"
)

var tmpl = template.Must(template.New("index").Parse(`
<html>
<head>
<title>CSRF Victim</title>
</head>
<body>
<p>账户余额:</p>
<ul>
  {{range $user, $balance := .Users}}
    <li>{{$user}}: {{$balance}} 元</li>
  {{end}}
</ul>

<hr/>

{{if .LoggedIn}}
  <p><a href="/logout">[退出]</a> 你好, {{.Username}}</p>
  <form method="POST" action="/transfer">
    <input name="csrf_token" type="hidden" value="{{.CSRFToken}}"/>
    <input name="amount" type="number"/>
    <button type="submit">转出账户</button>
  </form>
{{else}}
  <p>登录:</p>
  <form method="POST" action="/login">
    <input name="username"/>
    <button type="submit">登录</button>
  </form>
{{end}}
</body>
</html>`))

var users = make(map[string]int)
var sessions = make(map[string]string)
var csrfSessions = make(map[string]string)

func main() {
	r := gin.Default()
	r.SetHTMLTemplate(tmpl)
	r.GET("/", func(c *gin.Context) {
		// 获取会话 ID
		sessionID, _ := c.Cookie("sessionID")
		// 根据会话 ID 查找用户
		username, loggedIn := sessions[sessionID]
		csrfToken := magicconch.StringRand(32)
		csrfSessions[sessionID] = csrfToken

		c.HTML(200, "index", gin.H{
			"LoggedIn":  loggedIn,
			"Username":  username,
			"Users":     users,
			"CSRFToken": csrfToken,
		})
	})
	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		users[username] = 1000
		// 生成并保存新会话
		sessionID := magicconch.StringRand(32)
		sessions[sessionID] = username
		// 保存会话 ID 到客户端
		c.SetCookie("sessionID", sessionID, 3600, "/", "", false, false)
		// 登录后回到首页（不能直接把 POST 请求重定向到首页）
		c.Header("Content-Type", "text/html")
		c.String(200, "<script>window.location.replace(\"/\");</script>")
	})
	r.POST("/transfer", func(c *gin.Context) {
		// 获取会话 ID
		sessionID, _ := c.Cookie("sessionID")
		// 根据会话 ID 查找用户
		username, loggedIn := sessions[sessionID]
		if !loggedIn {
			c.String(403, "你没有登录")
			return
		}
		amount := magicconch.StringToInt(c.PostForm("amount"))
		csrfToken := c.PostForm("csrf_token")
		if csrfSessions[sessionID] != csrfToken {
			c.String(403, "CSRF 攻击，抓到你啦！（如果你不是恶意用户，请刷新页面重试）")
			return
		}
		users[username] -= amount
		// 转账完成后回到首页（不能直接把 POST 请求重定向到首页）
		c.Header("Content-Type", "text/html")
		c.String(200, "<script>window.location.replace(\"/\");</script>")
	})
	r.GET("/logout", func(c *gin.Context) {
		// 清除会话
		c.SetCookie("sessionID", "", 0, "/", "", false, false)
		// 退出登录后回到首页
		c.Redirect(307, "/")
	})
	err := r.Run("0.0.0.0:8080")
	magicconch.Must(err)
}
