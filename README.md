# BidNow Live — War Room CTF (Demo Build)

A third, standalone CTF — same 5 vulnerability topics as ShopKart v2 and
Grand Vistara, but its own story, its own discovery flow, and its own
payloads for every single flag. Built as the **demo/walkthrough CTF** to
show clients how CTF play works before they attempt ShopKart v2 and
Vistara, without giving away either of those answers.

| # | ShopKart (Basic) | Grand Vistara (Advanced) | BidNow War Room (Demo) |
|---|---|---|---|
| 1 | Login form bypass (`admin' --`) | Search box extraction (`' OR '1'='1'--`) | Self-service paddle lookup (`' OR '1'='1' --`) |
| 2 | Credentials in HTML comment | Predictable base64 reset token | Client-editable role cookie (DevTools) |
| 3 | `<script>alert()</script>` | `<img src=x onerror=alert()>` | `<svg onload=alert()>` |
| 4 | Direct `/api/users` URL | robots.txt → directory chain | sitemap.xml → `/backups/*.csv` |
| 5 | Exposed `/.env` file | Hidden `?debug=1` parameter | Verbose crash/stack-trace on bad input |

Interface is also different on purpose: a dark "War Room" terminal console
(not the same card styling as the other two), and each solved flag shows a
**2-part debrief** (Technique Used / The Fix) instead of a single reveal
line or a 3-part breakdown.

---

## 1. Requirements

- Go 1.22+
- A C compiler (for the SQLite driver)

## 2. How to run it

```bash
cd bidnow-ctf
go mod tidy   # first run only
go run .
```

Open **http://localhost:8083**. Restarting resets all progress.

## 3. How to play

1. Land on the BidNow Live storefront, browse the lots.
2. Click the red **📡 War Room** button (top right) to enter the CTF layer.
3. Work through the 5 leads (any order) — each has a briefing, a target
   link (or a "find it yourself" note), progressive hints, and a flag box.
4. On a correct submission you get a 2-part debrief: what technique it was,
   and the one-line fix. Close all 5 to finish the investigation.

## 4. What's actually vulnerable (for the facilitator)

| # | Lead | Vulnerability class | Where |
|---|---|---|---|
| 1 | The Paddle Lookup | SQL Injection | `/paddle-status` |
| 2 | The Trusted Cookie | Broken Authentication | `bn_role` cookie after `/login` |
| 3 | The Condition Report | Stored XSS (SVG onload) | `/lot/1` report box |
| 4 | The Backup Trail | Sensitive Data Exposure | `/sitemap.xml` → `/backups/paddle-export-2026.csv` |
| 5 | The Crash Report | Security Misconfiguration | `/lot/<non-numeric>` verbose error |

See `// VULNERABLE ON PURPOSE` comments in `handlers.go`. Full answers in
`SOLUTIONS.md`.

## 5. Project structure

```
bidnow-ctf/
  main.go        - server bootstrap & routing
  db.go           - in-memory SQLite schema + seed data
  flags.go        - the 5 leads: story text, hints, flag, technique, fix
  handlers.go     - all HTTP handlers (every bug lives here)
  session.go      - cookie-based per-player progress tracking
  templates/      - all HTML (Go html/template)
  static/         - style.css (dark terminal theme) + app.js
```
