package main

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"
)

type Memo struct {
	ID   int
	Text string
}

type User struct {
	Username string
	Password string
}

var memos []Memo
var users = map[string]string{}
var sessions = map[string]string{}
var tmpl = template.Must(template.ParseFiles("index.html", "login.html", "register.html", "edit.html", "users.html", "mypage.html"))
var mu sync.Mutex

func addMemo(text string) {
	mu.Lock()
	defer mu.Unlock()
	id := len(memos) + 1
	memo := Memo{ID: id, Text: text}
	memos = append(memos, memo)
}

func updateMemo(id int, text string) {
	mu.Lock()
	defer mu.Unlock()
	for i, memo := range memos {
		if memo.ID == id {
			memos[i].Text = text
			return
		}
	}
}

func deleteMemo(id int) {
	mu.Lock()
	defer mu.Unlock()
	for index, memo := range memos {
		if memo.ID == id {
			memos = append(memos[:index], memos[index+1:]...)
			return
		}
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		text := r.FormValue("text")
		addMemo(text)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl.ExecuteTemplate(w, "index.html", memos)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		idStr := r.FormValue("id")
		text := r.FormValue("text")
		id, err := strconv.Atoi(idStr)
		if err == nil {
			updateMemo(id, text)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var memo Memo
	for _, m := range memos {
		if m.ID == id {
			memo = m
			break
		}
	}

	tmpl.ExecuteTemplate(w, "edit.html", memo)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")
		if pwd, ok := users[username]; ok && pwd == password {
			sessionID := strconv.Itoa(len(sessions) + 1)
			sessions[sessionID] = username
			http.SetCookie(w, &http.Cookie{
				Name:  "session_id",
				Value: sessionID,
				Path:  "/",
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	tmpl.ExecuteTemplate(w, "login.html", nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")
		if _, exists := users[username]; !exists {
			users[username] = password
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	tmpl.ExecuteTemplate(w, "register.html", nil)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		deleteUser(username)
		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}

	tmpl.ExecuteTemplate(w, "users.html", users)
}

func deleteUser(username string) {
	mu.Lock()
	defer mu.Unlock()
	delete(users, username)
}

func mypageHandler(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	cookie, _ := r.Cookie("session_id")
	username := sessions[cookie.Value]

	tmpl.ExecuteTemplate(w, "mypage.html", username)
}

func isLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return false
	}
	_, ok := sessions[cookie.Value]
	return ok
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/users", deleteUserHandler)
	http.HandleFunc("/mypage", mypageHandler)
	http.ListenAndServe(":8080", nil)
}
