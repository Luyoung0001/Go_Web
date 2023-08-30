package main

import (
	"Go_Web/session"
	"html/template"
	"net/http"
	"time"
)

var globalSessions *session.Manager

// init 初始化
func init() {
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	go globalSessions.GC()
}
func login(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	err := r.ParseForm()
	if err != nil {
		return
	}
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		w.Header().Set("Content-Type", "text/html")
		err := t.Execute(w, sess.Get("username"))
		if err != nil {
			return
		}
	} else {
		err := sess.Set("username", r.Form["username"])
		if err != nil {
			return
		}
		http.Redirect(w, r, "/", 302)
	}
}
func count(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	createtime := sess.Get("createtime")
	if createtime == nil {
		sess.Set("createtime", time.Now().Unix())
	} else if (createtime.(int64) + 360) < (time.Now().Unix()) {
		globalSessions.SessionDestroy(w, r)
		sess = globalSessions.SessionStart(w, r)
	}
	ct := sess.Get("countnum")
	if ct == nil {
		err := sess.Set("countnum", 1)
		if err != nil {
			return

		} else {
			err := sess.Set("countnum", ct.(int)+1)
			if err != nil {
				return
			}
		}
		t, _ := template.ParseFiles("count.gtpl")
		w.Header().Set("Content-Type", "text/html")
		err = t.Execute(w, sess.Get("countnum"))
		if err != nil {
			return
		}
	}
}

func main() {

}
