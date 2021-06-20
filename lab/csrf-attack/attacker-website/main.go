package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spongeprojects/magicconch"
	"html/template"
)

var tmpl = template.Must(template.New("index").Parse(`
<html>
<head>
<title>Welcome</title>
</head>
<body>
<p>点击按钮，免费获取 50000 元现金奖励:</p>
<form method="POST" action="http://localhost:8080/transfer">
  <input name="amount" type="hidden" value="99999"/>
  <button type="submit">马上获取</button>
</form>
</body>
</html>`))

func main() {
	r := gin.Default()
	r.SetHTMLTemplate(tmpl)
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index", nil)
	})
	err := r.Run("0.0.0.0:8081")
	magicconch.Must(err)
}
