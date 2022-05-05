package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"youtube-board/my"

	"github.com/gorilla/sessions"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// db variable.
var dbName = "data.sqlite3"

// session variable.
var sesName = "ytboard-session"
var cs = sessions.NewCookieStore([]byte("secret-key-1234"))

// login check.
func checkLogin(w http.ResponseWriter, rq *http.Request) *my.User {
	ses, _ := cs.Get(rq, sesName)
	if ses.Values["login"] == nil || !ses.Values["login"].(bool) {
		http.Redirect(w, rq, "/login", http.StatusFound)
	}
	ac := ""
	if ses.Values["account"] != nil {
		ac = ses.Values["account"].(string)
	}

	var user my.User
	v2_db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db, _ := v2_db.DB()
	defer db.Close()
	v2_db.Where("account = ?", ac).First(&user)

	return &user
}

// Template for no-template.
func notemp() *template.Template {
	tmp, _ := template.New("index").Parse("no page.")
	return tmp
}

// get target Template.
func page(fname string) *template.Template {
	tmps, _ := template.ParseFiles("templates/"+fname+".html", "templates/head.html", "templates/foot.html")
	return tmps
}

// top page handler
func index(w http.ResponseWriter, rq *http.Request) {
	user := checkLogin(w, rq)

	v2_db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db, _ := v2_db.DB()
	defer db.Close()

	var pl []my.Post
	v2_db.Where("group_id > 0").Order("created_at desc").Limit(10).Find(&pl)
	var gl []my.Group
	v2_db.Order("created_at desc").Limit(10).Find(&gl)

	item := struct {
		Title   string
		Message string
		Name    string
		Account string
		Plist   []my.Post
		Glist   []my.Group
	}{
		Title:   "Index",
		Message: "this is top page.",
		Name:    user.Name,
		Account: user.Account,
		Plist:   pl,
		Glist:   gl,
	}
	er := page("index").Execute(w, item)
	if er != nil {
		log.Fatal(er)
	}
}

// top page handler.
func post(w http.ResponseWriter, rq *http.Request) {
	user := checkLogin(w, rq)

	pid := rq.FormValue("pid")
	v2_db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db, _ := v2_db.DB()
	defer db.Close()

	if rq.Method == "POST" {
		msg := rq.PostFormValue("message")
		pId, _ := strconv.Atoi(pid)
		cmt := my.Comment{
			UserId:  int(user.Model.ID),
			PostId:  pId,
			Message: msg,
		}
		v2_db.Create(&cmt)
	}

	var pst my.Post
	var cmts []my.CommentJoin

	v2_db.Where("id = ?", pid).First(&pst)
	v2_db.Table("comments").Select("comments.*, users.id, users.name").Joins("join users on users.id =comments.user_id").Where("comments.post_id = ?", pid).Order("created_at desc").Find(&cmts)

	item := struct {
		Title   string
		Message string
		Name    string
		Account string
		Post    my.Post
		Clist   []my.CommentJoin
	}{
		Title:   "Post",
		Message: "post id=" + pid,
		Name:    user.Name,
		Account: user.Account,
		Post:    pst,
		Clist:   cmts,
	}

	er := page("post").Execute(w, item)
	if er != nil {
		log.Fatal(er)
	}
}

// home handler
func home(w http.ResponseWriter, rq *http.Request) {
	user := checkLogin(w, rq)

	v2_db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db, _ := v2_db.DB()
	defer db.Close()

	if rq.Method == "POST" {
		switch rq.PostFormValue("form") {
		case "post":
			ad := rq.PostFormValue("address")
			ad = strings.TrimSpace(ad)
			if strings.HasPrefix(ad, "https://youtu.be/") {
				ad = strings.TrimPrefix(ad, "https://youtu.be/")
			}

			pt := my.Post{
				UserId:  int(user.Model.ID),
				Address: ad,
				Message: rq.PostFormValue("message"),
			}
			v2_db.Create(&pt)
		case "group":
			gp := my.Group{
				UserId:  int(user.Model.ID),
				Name:    rq.PostFormValue("name"),
				Message: rq.PostFormValue("message"),
			}
			v2_db.Create(&gp)
		}
	}

	var pts []my.Post
	var gps []my.Group

	v2_db.Where("user_id=?", user.ID).Order("created_at desc").Limit(10).Find(&pts)
	v2_db.Where("user_id=?", user.ID).Order("created_at desc").Limit(10).Find(&gps)

	item := struct {
		Title   string
		Message string
		Name    string
		Account string
		Plist   []my.Post
		Glist   []my.Group
	}{
		Title:   "Home",
		Message: "User account=\"" + user.Account + "\",",
		Name:    user.Name,
		Account: user.Account,
		Plist:   pts,
		Glist:   gps,
	}
	er := page("home").Execute(w, item)
	if er != nil {
		log.Fatal(er)
	}
}

// group handler.
func group(w http.ResponseWriter, rq *http.Request) {
	user := checkLogin(w, rq)

	gid := rq.FormValue("gid")
	v2_db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db, _ := v2_db.DB()
	defer db.Close()

	if rq.Method == "POST" {
		ad := rq.PostFormValue("address")
		ad = strings.TrimSpace(ad)
		if strings.HasPrefix(ad, "https://youtu.be/") {
			ad = strings.TrimPrefix(ad, "https://youtu.be/")
		}
		gId, _ := strconv.Atoi(gid)
		pt := my.Post{
			UserId:  int(user.Model.ID),
			Address: ad,
			Message: rq.PostFormValue("message"),
			GroupId: gId,
		}
		v2_db.Create(&pt)
	}

	var grp my.Group
	var pts []my.Post

	v2_db.Where("id = ?", gid).First(&grp)
	if err := v2_db.Order("created_at desc").Model(&grp).Association("Post").Append(&pts).Error; err != nil {
		log.Println(err)
	}

	item := struct {
		Title   string
		Message string
		Name    string
		Account string
		Group   my.Group
		Plist   []my.Post
	}{
		Title:   "Group",
		Message: "Group id=" + gid,
		Name:    user.Name,
		Account: user.Account,
		Group:   grp,
		Plist:   pts,
	}

	er := page("group").Execute(w, item)
	if er != nil {
		log.Fatal(er)
	}
}

// login handler.
func login(w http.ResponseWriter, rq *http.Request) {
	item := struct {
		Title   string
		Message string
		Account string
	}{
		Title:   "Login",
		Message: "type your account & password:",
		Account: "",
	}

	if rq.Method == "GET" {
		er := page("login").Execute(w, item)
		if er != nil {
			log.Fatal(er)
		}
		return
	}

	if rq.Method == "POST" {
		v2_db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		db, _ := v2_db.DB()
		defer db.Close()

		usr := rq.PostFormValue("account")
		pass := rq.PostFormValue("pass")
		item.Account = usr

		// check account and password
		var re int64
		var user my.User

		v2_db.Where("account = ? and password = ?", usr, pass).Find(&user).Count(&re)

		if re <= 0 {
			item.Message = "Wrong account or password."
			page("login").Execute(w, item)
			return
		}

		// logined.
		ses, _ := cs.Get(rq, sesName)
		ses.Values["login"] = true
		ses.Values["account"] = usr
		ses.Values["name"] = user.Name
		ses.Save(rq, w)
		http.Redirect(w, rq, "/", http.StatusFound)
	}
	er := page("login").Execute(w, item)
	if er != nil {
		log.Fatal(er)
	}
}

// logout handler
func logout(w http.ResponseWriter, rq *http.Request) {
	ses, _ := cs.Get(rq, sesName)
	ses.Values["login"] = nil
	ses.Values["account"] = nil
	ses.Save(rq, w)
	http.Redirect(w, rq, "/login", http.StatusFound)
}

// main program
func main() {
	// my.Migrate()

	// index handling.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index(w, r)
	})
	// home handling.
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		home(w, r)
	})
	// post handling.
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		post(w, r)
	})
	// group handling.
	http.HandleFunc("/group", func(w http.ResponseWriter, r *http.Request) {
		group(w, r)
	})
	// ligin handling.
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		login(w, r)
	})
	// logout handling.
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		logout(w, r)
	})

	http.ListenAndServe("", nil)
}
