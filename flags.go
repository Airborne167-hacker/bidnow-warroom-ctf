package main

// Lead holds everything the War Room console needs to render and grade one flag.
type Lead struct {
	ID          string
	Number      int
	CodeName    string
	Category    string
	Briefing    string
	TargetLabel string
	TargetURL   string
	Hints       []string
	Flag        string
	Technique   string // "what technique this was" - shown after correct submit
	Fix         string // "the one-line fix" - shown after correct submit
}

var leads = []Lead{
	{
		ID:          "l1",
		Number:      1,
		CodeName:    "The Paddle Lookup",
		Category:    "SQL Injection",
		Briefing:    "Bidders can check their own bid status by typing their paddle number into the self-service lookup tool. Our informant says this tool trusts whatever you type a little too much. Can you make it show you a paddle that isn't yours?",
		TargetLabel: "Open the Paddle Status Lookup",
		TargetURL:   "/paddle-status",
		Hints: []string{
			"The lookup tool takes one thing you type and drops it straight into a database question: \u201Cshow me the paddle whose number is exactly this.\u201D What if your input could change what that question actually asks?",
			"Try ending your input early with a single quote and adding a condition that's always true, then commenting out whatever comes after \u2014 something like:  ' OR '1'='1' --",
			"A few exact inputs worth trying in the Paddle Number field:  ' OR '1'='1' --   or   ' OR 1=1 --   or   ' OR 'a'='a' --",
		},
		Flag:      "BIDNOW{7c1a94ef2b6d0f83a5c1}",
		Technique: "SQL Injection \u2014 the lookup tool built its database question by gluing your raw input directly into the query text. A tautology (a condition that's always true) made the query match every row instead of just one, so the very first (and most sensitive) paddle in the table came back \u2014 not yours.",
		Fix:       "Use parameterised queries / prepared statements so user input is always treated as data, never as part of the query's logic.",
	},
	{
		ID:          "l2",
		Number:      2,
		CodeName:    "The Trusted Cookie",
		Category:    "Broken Authentication",
		Briefing:    "After logging in as a normal bidder, the portal remembers who you are using something stored right in your own browser. Our informant thinks the server is trusting that stored value a little too much \u2014 as if the browser could never lie to it.",
		TargetLabel: "Log in as the demo bidder",
		TargetURL:   "/login",
		Hints: []string{
			"Log in with the demo bidder account shown on the login page. Once you land on My Account, open your browser's developer tools (F12) and look at Application/Storage \u2192 Cookies for this site.",
			"You'll find a cookie that just says which role you are \u2014 in plain readable text, not encrypted or signed in any way. What happens if a value like that can be edited by whoever holds the browser?",
			"Edit that cookie's value directly in DevTools to say  admin  instead of  bidder , then refresh the My Account page.",
		},
		Flag:      "BIDNOW{2e6b81fa4d0c973e15af}",
		Technique: "Broken Authentication via client-side trust \u2014 the server decided what page/content to show based on a plain, unsigned cookie the browser itself controls. Anyone can edit a cookie's value in DevTools in five seconds; the server had no way to tell a real admin from an edited cookie.",
		Fix:       "Never make authorization decisions from a value the client can edit. Keep role/permission checks in server-side session state, and sign or encrypt any cookie that must carry identity information.",
	},
	{
		ID:          "l3",
		Number:      3,
		CodeName:    "The Condition Report",
		Category:    "Cross-Site Scripting (XSS)",
		Briefing:    "Bidders can leave a condition report / question on any lot before bidding, and every visitor sees it exactly as typed. Our informant thinks the report box doesn't check what kind of content you're actually submitting.",
		TargetLabel: "Open the Antique Gold Vase lot page",
		TargetURL:   "/lot/1",
		Hints: []string{
			"A normal report is just text. But this box doesn't seem to check if you typed text \u2014 or an image-like tag that the browser will try to render and run.",
			"Try submitting a report whose entire content is exactly:  <svg onload=alert('hi')>  \u2014 then look at the actual reports list on the page, not the form.",
			"Once your own popup shows 'hi', prove real impact: submit a report containing exactly  <svg onload=alert(BIDNOW_SECRET)>",
		},
		Flag:      "BIDNOW{f4a209c86e1b7d052fca}",
		Technique: "Stored Cross-Site Scripting via an SVG tag's onload event \u2014 the report content was saved and later re-rendered as raw HTML with no sanitisation. Browsers will happily execute JavaScript inside an SVG's onload attribute, not just inside <script> tags, which is why filters that only block <script> still miss this.",
		Fix:       "Encode all user-supplied content on output (escape HTML special characters) and, ideally, strip or allow-list HTML tags/attributes for anything a user can submit and others will view.",
	},
	{
		ID:          "l4",
		Number:      4,
		CodeName:    "The Backup Trail",
		Category:    "Sensitive Data Exposure",
		Briefing:    "Search engines and crawlers use a small file every website publishes to know what pages exist. Our informant thinks that file quietly points to something that was never meant to be public \u2014 a leftover backup from setting up tonight's auction.",
		TargetLabel: "No page for this one \u2014 start with the site's sitemap",
		TargetURL:   "",
		Hints: []string{
			"Most websites publish a file at  /sitemap.xml  listing every page they want indexed. It's meant for search engines, but nothing stops a person from reading it too.",
			"Open  /sitemap.xml  in your browser and read through every URL listed \u2014 one of them doesn't look like a normal page.",
			"One listed entry points to a file under  /backups/ \u2014 open that URL directly and read through its contents.",
		},
		Flag:      "BIDNOW{9b13ac57f0e28d641c9a}",
		Technique: "Sensitive Data Exposure \u2014 a backup export file was left inside the public web folder and, worse, was linked from the site's own sitemap, making it trivially discoverable. No login, no scanner, just reading a file search engines were told to crawl.",
		Fix:       "Never store backups, exports, or config files inside a publicly-served folder, and audit sitemap.xml / robots.txt regularly for anything that shouldn't be there.",
	},
	{
		ID:          "l5",
		Number:      5,
		CodeName:    "The Crash Report",
		Category:    "Security Misconfiguration",
		Briefing:    "Every application breaks sometimes \u2014 the question is what it shows you when it does. Our informant suspects tonight's platform still has its developer-facing error output switched on, in front of the public.",
		TargetLabel: "Open any lot page, then try breaking the URL",
		TargetURL:   "/lot/1",
		Hints: []string{
			"Lot pages are reached through a URL like  /lot/1 . That number is expected to be numeric \u2014 what happens if it isn't?",
			"Try visiting a lot URL with letters instead of a number, e.g.  /lot/report  \u2014 and read whatever error page comes back very carefully.",
			"A properly configured production app would show a generic 'Something went wrong' message. If instead you see internal details, variable names, or a token in the error output, that's the misconfiguration.",
		},
		Flag:      "BIDNOW{c8d571e2a94b036f7de8}",
		Technique: "Security Misconfiguration \u2014 verbose/debug-style error output was left enabled, so an unexpected input didn't just fail gracefully; it echoed internal details (including a sensitive token) straight into the public error page.",
		Fix:       "Disable verbose/debug error output in production; log full error details server-side only, and return a generic error message to the user.",
	},
}

func leadByID(id string) *Lead {
	for i := range leads {
		if leads[i].ID == id {
			return &leads[i]
		}
	}
	return nil
}
