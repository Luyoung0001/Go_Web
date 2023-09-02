package main

import (
	_ "Go_Web/memory"
	"Go_Web/session"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
)

// 全局 sessions 管理器
var globalSessions *session.Manager

// init 初始化

func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 600)
	go globalSessions.GC()
}
func login(c *gin.Context) {
	sess := globalSessions.SessionStart(c.Writer, c.Request)
	err := c.Request.ParseForm()
	if err != nil {
		return
	}
	if c.Request.Method == "GET" {
		t, err := template.ParseFiles("/Users/luliang/GoLand/Go_Web/template/login.html")
		if err != nil {
			fmt.Println(err)
		}
		c.Writer.Header().Set("Content-Type", "text/html")
		err = t.Execute(c.Writer, sess.Get("username"))
		if err != nil {
			return
		}
	} else {
		err := sess.Set("username", c.Request.Form["username"])
		if err != nil {
			return
		}
		http.Redirect(c.Writer, c.Request, "/", http.StatusPermanentRedirect)
	}
}
func count(c *gin.Context) {
	sess := globalSessions.SessionStart(c.Writer, c.Request)
	ct := sess.Get("countnum")
	if ct == nil {
		err := sess.Set("countnum", 1)
		if err != nil {
			return
		}
	} else {
		// 更新
		err := sess.Set("countnum", ct.(int)+1)
		if err != nil {
			return
		}
	}
	t, err := template.ParseFiles("template/count.html")
	if err != nil {
		fmt.Println(err)
	}
	c.Writer.Header().Set("Content-Type", "text/html")
	err = t.Execute(c.Writer, sess.Get("countnum"))
	if err != nil {
		return
	}
}

func main() {
	r := gin.Default()
	r.GET("/count", count)
	r.GET("/login", login)
	err := r.Run(":9000")
	if err != nil {
		return
	}

}
