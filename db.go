package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// initDB creates an in-memory SQLite database and seeds it for the BidNow
// Live "War Room" console. A REAL SQL engine is used on purpose so the
// paddle-lookup bug in handlers.go is a genuine, exploitable SQL Injection
// - not a simulation.
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE staff (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
		password TEXT NOT NULL,
		role TEXT NOT NULL,
		display_name TEXT NOT NULL
	);

	CREATE TABLE paddles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		paddle_no TEXT NOT NULL,
		bidder_name TEXT NOT NULL,
		pin TEXT NOT NULL,
		private_note TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE lots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		current_bid TEXT NOT NULL,
		tagline TEXT NOT NULL,
		description TEXT NOT NULL,
		emoji TEXT NOT NULL
	);

	CREATE TABLE condition_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		lot_id INTEGER NOT NULL,
		author TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TEXT NOT NULL
	);

	CREATE TABLE reset_requests (
		code INTEGER PRIMARY KEY,
		paddle_no TEXT NOT NULL,
		used INTEGER NOT NULL DEFAULT 0
	);
	`
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("schema create: %v", err)
	}

	seed := `
	INSERT INTO staff (email, password, role, display_name) VALUES
		('director@bidnow.example', 'Auct10n#M@stervault!', 'admin', 'Auction Director'),
		('bidder@bidnow.example', 'demo1234', 'bidder', 'Demo Bidder');

	INSERT INTO paddles (paddle_no, bidder_name, pin, private_note) VALUES
		('PDL-1003', 'Confidential Consignor Account', '0000', 'BIDNOW{7c1a94ef2b6d0f83a5c1} -- this paddle backs the ₹22L necklace lot, PIN was never rotated after setup'),
		('PDL-1001', 'Demo Bidder', '4477', ''),
		('PDL-1002', 'Rohit Sharma', '9081', '');

	INSERT INTO lots (title, current_bid, tagline, description, emoji) VALUES
		('Antique Gold Vase', '12,40,000', 'Circa 1920, hand-etched gold leaf', 'A rare hand-etched vase from a private European estate collection, restored by certified conservators.', '🏺'),
		('Vintage Rolex Chronograph', '18,75,000', 'Full original set, box and papers', 'A 1968 chronograph in excellent condition, recently serviced with all original documentation intact.', '⌚'),
		('First Edition Manuscript', '6,20,000', 'Signed by the author, 1953 print run', 'One of only 200 surviving first-edition copies, signed on the inside cover.', '📜'),
		('Modernist Oil Painting', '9,90,000', 'Original canvas, 1974', 'A striking large-format oil painting from a celebrated modernist collection, never before auctioned.', '🖼️'),
		('Diamond Constellation Necklace', '22,10,000', 'GIA certified, 4.8 carat total', 'A one-of-a-kind necklace featuring a constellation pattern of certified diamonds.', '💎');

	INSERT INTO condition_reports (lot_id, author, content, created_at) VALUES
		(1, 'Kabir R.', 'Does the vase come with a certificate of authenticity?', '3 hours ago'),
		(1, 'Meera T.', 'Any visible restoration marks under UV light?', '1 hour ago');

	INSERT INTO reset_requests (code, paddle_no, used) VALUES
		(58213, 'PDL-1001', 0),
		(58214, 'PDL-1002', 0);
	`
	if _, err := db.Exec(seed); err != nil {
		log.Fatalf("seed insert: %v", err)
	}
}
