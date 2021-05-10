package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spongeprojects/magicconch"
	"html/template"
)

var tmpl = template.Must(template.New("index").Parse(`
<html>
<head>
<title>XSS Victim</title>
</head>
<body>
<p>注册用户列表:</p>
<ul>
  {{range .Users}}
    <li>{{.}}</li>
  {{end}}
</ul>

<hr/>

{{if .LoggedIn}}
  <p><a href="/logout">[退出]</a> 你好, {{.Username}}</p>
{{else}}
  <p>登录:</p>
  <form method="POST" action="/login">
    <input name="username"/>
    <button type="submit">登录</button>
  </form>
{{end}}
</body>
</html>`))

var users = make([]string, 0)
var sessions = make(map[string]string)

func main() {
	r := gin.Default()
	r.SetHTMLTemplate(tmpl)
	r.GET("/", func(c *gin.Context) {
		// 获取会话 ID
		sessionID, _ := c.Cookie("sessionID")
		// 根据会话 ID 查找用户
		username, loggedIn := sessions[sessionID]

		// 为了演示 XSS 攻击我们需要特地将 username 包装为 template.HTML,
		// 因为 html/template 包默认会帮我们对特殊字符进行转义
		var usersHTML []template.HTML
		for _, user := range users {
			usersHTML = append(usersHTML, template.HTML(user))
		}
		c.HTML(200, "index", gin.H{
			"LoggedIn": loggedIn,
			"Username": username,
			"Users":    usersHTML,
		})
	})
	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		if !magicconch.StringInSlice(username, users) {
			users = append(users, username)
		}
		// 生成并保存新会话
		sessionID := magicconch.StringRand(32)
		sessions[sessionID] = username
		// 保存会话 ID 到客户端
		c.SetCookie("sessionID", sessionID, 3600, "/", "", false, false)
		// 登录后回到首页（不能直接把 POST 请求重定向到首页）
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
