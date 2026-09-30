package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var tmpl *template.Template

func loadTemplates() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

func exe(w http.ResponseWriter, name string, data any) {
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Println("template:", name, err)
	}
}

type Lot struct {
	ID          int
	Title       string
	CurrentBid  string
	Tagline     string
	Description string
	Emoji       string
}

type Report struct {
	ID        int
	LotID     int
	Author    string
	Content   template.HTML // deliberately unescaped - this IS the XSS bug
	CreatedAt string
}

// ---------- portal ----------

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	sess := getOrCreate(w, r)
	rows, err := db.Query(`SELECT id, title, current_bid, tagline, description, emoji FROM lots ORDER BY id`)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	defer rows.Close()
	var lots []Lot
	for rows.Next() {
		var l Lot
		if err := rows.Scan(&l.ID, &l.Title, &l.CurrentBid, &l.Tagline, &l.Description, &l.Emoji); err == nil {
			lots = append(lots, l)
		}
	}
	exe(w, "home", map[string]any{"Sess": sess, "Lots": lots})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreate(w, r)
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		email := r.FormValue("email")
		password := r.FormValue("password")
		var role, name string
		err := db.QueryRow(`SELECT role, display_name FROM staff WHERE email=? AND password=?`, email, password).
			Scan(&role, &name)
		if err != nil {
			exe(w, "login", map[string]any{"Sess": sess, "Error": "Invalid email or password."})
			return
		}
		mu.Lock()
		sess.LoggedInAs = email
		sess.DisplayName = name
		mu.Unlock()

		// VULNERABLE ON PURPOSE: a plain, readable, unsigned cookie carries
		// the caller's role. Anyone can open DevTools and edit this cookie
		// to any value - the server below trusts it blindly. Seeded
		// deliberately for the "Trusted Cookie" challenge (Broken Auth).
		http.SetCookie(w, &http.Cookie{
			Name:  "bn_role",
			Value: "bidder", // every fresh login always starts as plain "bidder"
			Path:  "/",
		})

		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}
	exe(w, "login", map[string]any{"Sess": sess})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreate(w, r)
	mu.Lock()
	sess.LoggedInAs = ""
	sess.DisplayName = ""
	mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "bn_role", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func accountHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreate(w, r)
	if sess.LoggedInAs == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// VULNERABLE ON PURPOSE: the decision to show privileged content is
	// made from the client-editable bn_role cookie, not from server-side
	// session/role state. Seeded for the "Trusted Cookie" challenge.
	roleCookie, _ := r.Cookie("bn_role")
	isAdmin := roleCookie != nil && roleCookie.Value == "admin"

	var flag string
	if isAdmin {
		flag = leadByID("l2").Flag
	}
	exe(w, "account", map[string]any{
		"Sess":    sess,
		"IsAdmin": isAdmin,
		"Flag":    flag,
	})
}

func paddleStatusHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreate(w, r)
	var result map[string]string
	var queried bool

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		paddleNo := r.FormValue("paddle_no")
		queried = true

		// VULNERABLE ON PURPOSE: raw string concatenation into SQL instead
		// of a parameterised query. Seeded deliberately for the "Paddle
		// Lookup" challenge (SQL Injection). Never build SQL like this in
		// a real application.
		query := fmt.Sprintf(
			"SELECT paddle_no, bidder_name, private_note FROM paddles WHERE paddle_no = '%s'",
			paddleNo,
		)
		row := db.QueryRow(query)
		var pn, name, note string
		if err := row.Scan(&pn, &name, &note); err == nil {
			result = map[string]string{"PaddleNo": pn, "Name": name, "Note": note}
		}
	}

	exe(w, "paddle_status", map[string]any{"Sess": sess, "Result": result, "Queried": queried})
}

func lotHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreate(w, r)
	idStr := strings.TrimPrefix(r.URL.Path, "/lot/")

	id, convErr := strconv.Atoi(idStr)
	if convErr != nil {
		// VULNERABLE ON PURPOSE: a misconfigured "developer mode" error
		// page that leaks internal details (including a sensitive token)
		// on a simple bad-input crash, instead of showing a generic error.
		// Seeded deliberately for the "Crash Report" challenge (Security
		// Misconfiguration).
		w.WriteHeader(http.StatusInternalServerError)
		exe(w, "crash", map[string]any{
			"Sess":      sess,
			"BadInput":  idStr,
			"ParseErr":  convErr.Error(),
			"DebugFlag": leadByID("l5").Flag,
		})
		return
	}

	var l Lot
	err := db.QueryRow(`SELECT id, title, current_bid, tagline, description, emoji FROM lots WHERE id = ?`, id).
		Scan(&l.ID, &l.Title, &l.CurrentBid, &l.Tagline, &l.Description, &l.Emoji)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		author := r.FormValue("author")
		content := r.FormValue("content")
		if author == "" {
			author = "Anonymous bidder"
		}
		if strings.TrimSpace(content) != "" {
			// VULNERABLE ON PURPOSE: report content is stored and later
			// rendered as raw HTML with no sanitisation/output-encoding.
			// Seeded deliberately for the "Condition Report" challenge
			// (Stored XSS via SVG onload).
			_, err := db.Exec(
				`INSERT INTO condition_reports (lot_id, author, content, created_at) VALUES (?, ?, ?, ?)`,
				l.ID, author, content, "just now",
			)
			if err != nil {
				log.Println("report insert error:", err)
			}
		}
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		return
	}

	rows, err := db.Query(`SELECT id, lot_id, author, content, created_at FROM condition_reports WHERE lot_id = ? ORDER BY id DESC`, l.ID)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	defer rows.Close()
	var reports []Report
	for rows.Next() {
		var rp Report
		var content string
		if err := rows.Scan(&rp.ID, &rp.LotID, &rp.Author, &content, &rp.CreatedAt); err == nil {
			rp.Content = template.HTML(content) // raw, unescaped - intentional bug
			reports = append(reports, rp)
		}
	}

	exe(w, "lot", map[string]any{
		"Sess":    sess,
		"Lot":     l,
		"Reports": reports,
		"Secret":  leadByID("l3").Flag,
	})
}

// ---------- deliberately exposed surfaces ----------

// sitemapHandler publishes a normal-looking sitemap that also links a
// leftover backup export. VULNERABLE ON PURPOSE: Sensitive Data Exposure,
// seeded for the "Backup Trail" challenge.
func sitemapHandler(w http.ResponseWriter, r *http.Request) {
	body := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>http://localhost:8083/</loc></url>
  <url><loc>http://localhost:8083/lot/1</loc></url>
  <url><loc>http://localhost:8083/lot/2</loc></url>
  <url><loc>http://localhost:8083/lot/3</loc></url>
  <url><loc>http://localhost:8083/lot/4</loc></url>
  <url><loc>http://localhost:8083/lot/5</loc></url>
  <url><loc>http://localhost:8083/backups/paddle-export-2026.csv</loc></url>
</urlset>`
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(body))
}

func backupFileHandler(w http.ResponseWriter, r *http.Request) {
	body := `paddle_no,bidder_name,pin,notes
PDL-1001,Demo Bidder,4477,
PDL-1002,Rohit Sharma,9081,
PDL-1003,Confidential Consignor Account,0000,"left in export by mistake - ` + leadByID("l4").Flag + `"
`
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	_, _ = w.Write([]byte(body))
}

// ---------- War Room (the CTF layer) ----------

func warroomHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreate(w, r)
	mu.Lock()
	foundCopy := make(map[string]bool, len(sess.Found))
	for k, v := range sess.Found {
		foundCopy[k] = v
	}
	mu.Unlock()

	exe(w, "warroom", map[string]any{
		"Sess":      sess,
		"Leads":     leads,
		"Found":     sess.foundCount(),
		"Total":     len(leads),
		"FoundJSON": foundCopy,
	})
}

type hintReq struct {
	ID string `json:"id"`
}
type hintResp struct {
	Hints []string `json:"hints"`
	Done  bool     `json:"done"`
}

func hintHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreate(w, r)
	var req hintReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	ld := leadByID(req.ID)
	if ld == nil {
		http.Error(w, "unknown lead", 404)
		return
	}
	mu.Lock()
	shown := sess.HintsShown[req.ID]
	if shown < len(ld.Hints) {
		shown++
		sess.HintsShown[req.ID] = shown
	}
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(hintResp{Hints: ld.Hints[:shown], Done: shown >= len(ld.Hints)})
}

type submitReq struct {
	ID   string `json:"id"`
	Flag string `json:"flag"`
}
type submitResp struct {
	Correct   bool   `json:"correct"`
	Message   string `json:"message"`
	Technique string `json:"technique,omitempty"`
	Fix       string `json:"fix,omitempty"`
	Found     int    `json:"found"`
	Total     int    `json:"total"`
	AllDone   bool   `json:"all_done"`
	LeadID    string `json:"lead_id"`
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreate(w, r)
	var req submitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	ld := leadByID(req.ID)
	if ld == nil {
		http.Error(w, "unknown lead", 404)
		return
	}
	submitted := strings.TrimSpace(req.Flag)
	resp := submitResp{LeadID: req.ID}

	if submitted == ld.Flag {
		mu.Lock()
		already := sess.Found[req.ID]
		sess.Found[req.ID] = true
		found := sess.foundCount()
		mu.Unlock()

		resp.Correct = true
		resp.Found = found
		resp.Total = len(leads)
		resp.AllDone = found == len(leads)
		if already {
			resp.Message = "Already on file - good confirmation though."
		} else {
			resp.Message = "Confirmed. Lead closed."
			resp.Technique = ld.Technique
			resp.Fix = ld.Fix
		}
	} else {
		resp.Correct = false
		resp.Message = "Doesn't match our records. Pull another thread."
		mu.Lock()
		resp.Found = sess.foundCount()
		resp.Total = len(leads)
		mu.Unlock()
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func progressHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreate(w, r)
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"found":       sess.Found,
		"hints_shown": sess.HintsShown,
		"total":       len(leads),
		"server_time": time.Now().Format(time.RFC3339),
	})
}
