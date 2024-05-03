package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// setupRoutes initializes all the route handlers
func setupRoutes() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/protected", serveProtected)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username, password := r.Form.Get("username"), r.Form.Get("password")
	if correctPassword, ok := users[username]; ok && password == correctPassword {
		sessionToken := fmt.Sprintf("%s-%d", username, time.Now().Unix())
		sessions[sessionToken] = username
		http.SetCookie(w, &http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
			Path:  "/",
		})
		http.Redirect(w, r, "/protected", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func serveProtected(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username, ok := sessions[cookie.Value]
	if !ok {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	data := struct {
		Username string
		CanEdit  bool
		CanView  bool
	}{
		Username: username,
		CanEdit:  checkPermission(username, "edit"),
		CanView:  checkPermission(username, "view"),
	}

	tmpl, err := template.ParseFiles("static/protected.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}
