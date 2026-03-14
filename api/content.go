package api

// contentPage wraps body content in the shared page shell
func contentPage(title, description, canonical, body string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + title + `</title>
<meta name="description" content="` + description + `">
<meta name="keywords" content="file storage AI agents, S3 filesystem, persistent storage LLM, agent file system, cloud storage AI, pidrive">
<link rel="canonical" href="` + canonical + `">
<meta property="og:title" content="` + title + `">
<meta property="og:description" content="` + description + `">
<meta property="og:url" content="` + canonical + `">
<meta property="og:type" content="article">
<meta property="og:site_name" content="pidrive">
<meta name="twitter:card" content="summary">
<meta name="twitter:title" content="` + title + `">
<meta name="twitter:description" content="` + description + `">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#0a0a0a;color:#e0e0e0;font-family:"SF Mono","Fira Code",Menlo,monospace;font-size:15px;line-height:1.7;padding:60px 24px;max-width:720px;margin:0 auto}
a{color:#6cb6ff;text-decoration:none}
a:hover{text-decoration:underline}
.hd{margin-bottom:48px}
.hd a{color:#888;font-size:13px}
h1{color:#fff;font-size:24px;font-weight:700;margin:16px 0 8px}
h2{color:#ccc;font-size:18px;font-weight:600;margin:40px 0 12px}
h3{color:#aaa;font-size:15px;font-weight:600;margin:28px 0 8px}
p{color:#aaa;font-size:15px;margin-bottom:16px;line-height:1.8}
p strong{color:#fff}
ul,ol{color:#aaa;margin:0 0 16px 24px;line-height:2}
pre{background:#141414;border:1px solid #222;border-radius:8px;padding:16px 20px;overflow-x:auto;margin-bottom:16px;font-size:14px;line-height:1.8}
code{color:#7ec699;font-size:14px}
.cta{margin:40px 0;padding:24px;background:#141414;border:1px solid #222;border-radius:8px}
.cta pre{margin:0;border:0;padding:0;background:none}
.ft{margin-top:60px;padding-top:24px;border-top:1px solid #1a1a1a;color:#555;font-size:13px}
</style>
</head>
<body>
<div class="hd"><a href="{{SERVER_URL}}">&larr; pidrive</a></div>
` + body + `
<div class="ft"><a href="{{SERVER_URL}}">pidrive</a> &middot; <a href="{{SERVER_URL}}/docs">docs</a> &middot; <a href="{{SERVER_URL}}/skill.md">skill.md</a></div>
</body>
</html>`
}

// --- Docs page ---

var docsHTMLTemplate = contentPage(
	"pidrive docs — file storage for AI agents",
	"Complete documentation for pidrive. Install, mount, share files, search, and manage storage for your AI agents.",
	"{{SERVER_URL}}/docs",
	`<h1>pidrive documentation</h1>
<p>pidrive gives your AI agents a filesystem backed by S3. Install the CLI, mount your drive, and use standard unix commands. No SDK, no API calls — just files.</p>

<h2>Install</h2>
<pre>curl -sSL {{SERVER_URL}}/install.sh | bash</pre>
<p>This installs the <code>pidrive</code> CLI binary. On Linux it also installs <code>davfs2</code> for WebDAV mount support. macOS has WebDAV built in.</p>

<h2>Register</h2>
<pre>pidrive register --email agent@company.com --name "My Agent" --server {{SERVER_URL}}</pre>
<p>Check your email for the verification code:</p>
<pre>pidrive verify --email agent@company.com --code 123456</pre>
<p>Your API key is saved to <code>~/.pidrive/credentials</code>.</p>

<h2>Mount your drive</h2>
<pre>pidrive mount</pre>
<p>This mounts your private storage at <code>/drive/</code>. Every file you create there is stored in S3. Nothing is saved locally.</p>

<h2>Read and write files</h2>
<p>Use any unix command. They all work:</p>
<pre>ls /drive/
echo "hello world" &gt; /drive/notes.txt
cat /drive/notes.txt
cp report.pdf /drive/
mkdir /drive/output/
grep -r "error" /drive/logs/
head -20 /drive/data.csv
wc -l /drive/data.csv
rm /drive/old-file.txt</pre>

<h2>Share files</h2>

<h3>Share with another agent</h3>
<pre>pidrive share report.pdf --to other-agent@company.com</pre>
<p>The other agent gets a copy in their drive.</p>

<h3>Share with a link</h3>
<pre>pidrive share data.csv --link</pre>
<p>Returns a public URL like <code>{{SERVER_URL}}/s/abc123</code>. Anyone with the link can download the file.</p>

<h3>Share with expiry</h3>
<pre>pidrive share data.csv --link --expires 7d</pre>

<h3>List and revoke shares</h3>
<pre>pidrive shared
pidrive revoke &lt;share-id&gt;</pre>

<h3>Download a shared file</h3>
<pre>pidrive pull {{SERVER_URL}}/s/abc123 ./local-copy.csv</pre>

<h2>Search</h2>
<pre>pidrive search "quarterly revenue"</pre>
<p>Full-text search across all your files. Text files are indexed automatically in the background. Supported formats: .txt, .md, .csv, .json, .yaml, .py, .js, .go, .log, and more.</p>

<h2>Trash</h2>
<pre>pidrive trash
pidrive restore report.txt</pre>
<p>Deleted files are kept for 30 days before permanent removal.</p>

<h2>Check status</h2>
<pre>pidrive status</pre>
<p>Shows mount status, server connection, account info, and storage usage.</p>

<pre>pidrive whoami</pre>
<p>Shows your email, plan, and storage quota.</p>

<pre>pidrive usage</pre>
<p>Shows storage used and bandwidth consumed.</p>

<h2>Activity log</h2>
<pre>pidrive activity</pre>
<p>Shows recent events: mounts, shares, revokes, restores.</p>

<h2>Plans</h2>
<pre>pidrive plans</pre>
<table style="color:#aaa;font-size:14px;margin:16px 0;border-collapse:collapse;width:100%">
<tr style="border-bottom:1px solid #222"><td style="padding:8px 16px 8px 0"><strong style="color:#fff">Free</strong></td><td>$0/mo</td><td>1 GB storage</td><td>100 MB bandwidth</td></tr>
<tr style="border-bottom:1px solid #222"><td style="padding:8px 16px 8px 0"><strong style="color:#fff">Pro</strong></td><td>$5/mo</td><td>100 GB storage</td><td>10 GB bandwidth</td></tr>
<tr><td style="padding:8px 16px 8px 0"><strong style="color:#fff">Team</strong></td><td>$20/mo</td><td>1 TB storage</td><td>Unlimited bandwidth</td></tr>
</table>
<pre>pidrive upgrade --plan pro</pre>

<h2>Unmount</h2>
<pre>pidrive unmount</pre>

<h2>All commands</h2>
<pre>pidrive register    Register a new agent
pidrive login       Login to existing account
pidrive verify      Verify email code
pidrive whoami      Show account info
pidrive mount       Mount your drive
pidrive unmount     Unmount your drive
pidrive status      Show connection status
pidrive share       Share a file
pidrive shared      List shares
pidrive revoke      Revoke a share
pidrive pull        Download shared file
pidrive search      Full-text search
pidrive activity    Recent events
pidrive trash       List deleted files
pidrive restore     Restore from trash
pidrive usage       Storage stats
pidrive plans       Available plans
pidrive upgrade     Change plan</pre>

<h2>FAQ</h2>

<h3>Are files stored on my machine?</h3>
<p>No. The mount point (<code>/drive/</code>) is a tunnel to the server. All data lives in S3. If your VM dies, nothing is lost.</p>

<h3>Can I use any unix tool?</h3>
<p>Yes. <code>ls</code>, <code>cat</code>, <code>grep</code>, <code>cp</code>, <code>mv</code>, <code>rm</code>, <code>head</code>, <code>tail</code>, <code>wc</code>, <code>find</code>, pipes, redirects — everything works. The mount behaves like a normal directory.</p>

<h3>How is this different from S3?</h3>
<p>S3 is an API. You need SDKs, presigned URLs, multipart uploads. pidrive is a filesystem. You just <code>echo "data" &gt; /drive/file.txt</code>.</p>

<h3>Can agents see each other's files?</h3>
<p>No. Each agent is isolated. You only see your own files. Sharing is explicit — you choose what to share and with whom.</p>

<h3>What happens if I delete a file?</h3>
<p>It goes to trash. You have 30 days to restore it. After that, it is permanently deleted.</p>

<h3>Is there a file size limit?</h3>
<p>Free plan: 1 GB total storage. Pro: 100 GB. Team: 1 TB. Individual files can be up to your remaining quota.</p>

<h3>Can I use this from a CI/CD pipeline?</h3>
<p>Yes. Install the CLI, set credentials, mount, read/write files. Works in GitHub Actions, GitLab CI, Docker containers — anywhere Linux runs.</p>
`)

// --- Blog pages ---

var blogPages = map[string]string{
	"why-agents-need-files": contentPage(
		"Why AI agents need file storage — pidrive",
		"AI agents generate reports, logs, and data but have nowhere to put them. Here's why agents need persistent file storage and how to give it to them.",
		"{{SERVER_URL}}/blog/why-agents-need-files",
		`<h1>Why AI agents need file storage</h1>

<p>AI agents are getting good at doing work. They write code, analyze data, generate reports, scrape websites, process documents. But when the work is done, where does the output go?</p>

<p><strong>Most agents have no persistent storage.</strong> They run in ephemeral containers. They write to /tmp. The container dies, the files are gone.</p>

<h2>The problem</h2>

<p>Consider an agent that:</p>
<ul>
<li>Generates a weekly sales report every Monday</li>
<li>Scrapes competitor pricing daily</li>
<li>Processes uploaded invoices and extracts line items</li>
<li>Runs data analysis and produces CSV output</li>
</ul>

<p>Each of these produces files. Where do they go?</p>

<p><strong>Option 1: Local disk.</strong> Gone when the container restarts. Not accessible from other agents or services.</p>

<p><strong>Option 2: S3 API.</strong> Works, but now your agent needs AWS SDKs, credentials management, presigned URLs for sharing, multipart upload logic for large files. That is a lot of code for "save a file".</p>

<p><strong>Option 3: Database blob.</strong> Technically works. Practically terrible. No directory structure, no search, no unix tools.</p>

<h2>What agents actually want</h2>

<p>Agents want what every program wants: a filesystem.</p>

<pre>echo "Q4 revenue: $2.4M" &gt; /drive/reports/q4-2024.txt
grep -r "error" /drive/logs/
ls /drive/output/</pre>

<p>Simple. No SDKs. No API calls. Just files and directories. Standard unix commands that have worked for 50 years.</p>

<h2>The solution</h2>

<p>pidrive gives agents a mounted filesystem backed by S3. Install the CLI, run <code>pidrive mount</code>, and <code>/drive/</code> is your agent's private storage.</p>

<ul>
<li><strong>Persistent</strong> — files survive container restarts, VM termination, even datacenter moves. Data lives in S3.</li>
<li><strong>Private</strong> — each agent sees only its own files. No cross-contamination.</li>
<li><strong>Shareable</strong> — share files with other agents by email, or create public URLs.</li>
<li><strong>Searchable</strong> — full-text search across all your files.</li>
<li><strong>No SDK required</strong> — standard unix commands. If your agent can run <code>cat</code>, it can use pidrive.</li>
</ul>

<div class="cta">
<p><strong>Get started in 30 seconds:</strong></p>
<pre>curl -sSL {{SERVER_URL}}/install.sh | bash
pidrive register --email agent@company.com --name "My Agent" --server {{SERVER_URL}}
pidrive mount
echo "it works" &gt; /drive/test.txt</pre>
</div>
`),

	"s3-is-not-a-filesystem": contentPage(
		"S3 is not a filesystem — why AI agents need something better",
		"S3 is great for storage but terrible as a filesystem for AI agents. No ls, no grep, no pipes. Here's a better way.",
		"{{SERVER_URL}}/blog/s3-is-not-a-filesystem",
		`<h1>S3 is not a filesystem</h1>

<p>S3 is incredible infrastructure. It is cheap, durable, and scales to exabytes. But it is not a filesystem, and pretending it is one causes real pain for AI agents.</p>

<h2>What agents want to do</h2>

<pre>ls /data/
cat /data/report.txt
grep -r "error" /data/logs/
echo "result" &gt; /data/output.txt
cp /data/report.txt /shared/</pre>

<h2>What S3 makes you do</h2>

<pre># List files
aws s3api list-objects-v2 --bucket my-bucket --prefix data/

# Read a file
aws s3 cp s3://my-bucket/data/report.txt -

# Search? No grep. Download everything first.
aws s3 sync s3://my-bucket/data/ /tmp/data/
grep -r "error" /tmp/data/logs/

# Write a file
aws s3 cp - s3://my-bucket/data/output.txt

# Share? Generate a presigned URL
aws s3 presign s3://my-bucket/data/report.txt --expires-in 3600</pre>

<p>Every operation requires the AWS SDK, credentials, bucket names, and prefix management. Your agent spends more time on S3 plumbing than on actual work.</p>

<h2>The real problems</h2>

<ul>
<li><strong>No directory listing.</strong> S3 has "prefixes", not directories. <code>ls</code> does not work.</li>
<li><strong>No append.</strong> You cannot append to an S3 object. You must re-upload the entire file.</li>
<li><strong>No search.</strong> You cannot grep across S3 objects without downloading them all.</li>
<li><strong>No pipes.</strong> <code>cat file | grep error | wc -l</code> — impossible directly on S3.</li>
<li><strong>No permissions per agent.</strong> IAM policies are bucket-level or prefix-level, not per-agent.</li>
<li><strong>Credentials management.</strong> Every agent needs AWS credentials. Rotating them is painful.</li>
</ul>

<h2>What if S3 was a filesystem?</h2>

<p>That is what pidrive does. Your agent mounts <code>/drive/</code> and gets a real POSIX filesystem. Under the hood, data goes to S3. But your agent never knows or cares.</p>

<pre># Same files, stored in S3, accessed as a filesystem
ls /drive/
cat /drive/report.txt
grep -r "error" /drive/logs/
echo "result" &gt; /drive/output.txt</pre>

<p>No SDK. No credentials in your agent code. No presigned URLs. Just files.</p>

<div class="cta">
<p><strong>Try it:</strong></p>
<pre>curl -sSL {{SERVER_URL}}/install.sh | bash
pidrive register --email you@company.com --name "My Agent" --server {{SERVER_URL}}
pidrive mount
ls /drive/</pre>
</div>
`),

	"sharing-files-between-agents": contentPage(
		"How to share files between AI agents — pidrive",
		"AI agents need to pass files to each other: reports, data, configs. pidrive makes agent-to-agent file sharing simple with one command.",
		"{{SERVER_URL}}/blog/sharing-files-between-agents",
		`<h1>Sharing files between AI agents</h1>

<p>Modern AI systems are not one agent. They are teams of agents. A research agent gathers data. An analysis agent processes it. A reporting agent writes the summary. A delivery agent sends it out.</p>

<p><strong>They all need to pass files to each other.</strong></p>

<h2>How it usually works (badly)</h2>

<ul>
<li><strong>Shared S3 bucket.</strong> All agents dump files in one bucket. No isolation. Agent A can accidentally delete Agent B's data. No audit trail.</li>
<li><strong>Message passing.</strong> Stuff file contents into JSON messages. Works until someone tries to pass a 50 MB CSV.</li>
<li><strong>Shared database.</strong> Store file bytes in Postgres. Slow, expensive, no unix tools.</li>
<li><strong>Shared volume mount.</strong> All agents mount the same directory. No privacy, no permissions, no sharing controls.</li>
</ul>

<h2>How pidrive handles it</h2>

<p>Each agent has private storage. Sharing is explicit.</p>

<h3>Share with a specific agent</h3>
<pre>pidrive share report.pdf --to analyst@company.com</pre>
<p>The analyst agent gets a copy in their drive. Your original is untouched.</p>

<h3>Share with anyone via URL</h3>
<pre>pidrive share data.csv --link
&rarr; {{SERVER_URL}}/s/x7k2m9</pre>
<p>Anyone with the URL can download it. No auth needed. Great for webhooks, dashboards, or external consumers.</p>

<h3>Share with expiry</h3>
<pre>pidrive share credentials.txt --link --expires 1h</pre>
<p>Link dies after one hour. Sensitive data does not live forever.</p>

<h3>Revoke access</h3>
<pre>pidrive shared
pidrive revoke &lt;share-id&gt;</pre>
<p>Changed your mind? Revoke it. The shared copy is deleted.</p>

<h2>The workflow</h2>

<pre># Research agent
echo "competitor data..." &gt; /drive/research/competitors.csv
pidrive share research/competitors.csv --to analyst@company.com

# Analyst agent
cat /drive/competitors.csv
# ... process data ...
echo "analysis results..." &gt; /drive/analysis.txt
pidrive share analysis.txt --to reporter@company.com

# Reporter agent
cat /drive/analysis.txt
# ... write report ...
pidrive share report.pdf --link
&rarr; {{SERVER_URL}}/s/abc123</pre>

<p>Each agent has its own private space. Files flow between them explicitly. There is an audit trail of every share.</p>

<div class="cta">
<p><strong>Get started:</strong></p>
<pre>curl -sSL {{SERVER_URL}}/install.sh | bash
pidrive register --email agent@company.com --name "My Agent" --server {{SERVER_URL}}
pidrive mount
pidrive share myfile.txt --to other-agent@company.com</pre>
</div>
`),
}

// --- VS pages ---

var vsPages = map[string]string{
	"google-drive": contentPage(
		"pidrive vs Google Drive for AI agents",
		"Google Drive is built for humans. pidrive is built for AI agents. Compare file storage solutions for autonomous agents and LLM applications.",
		"{{SERVER_URL}}/vs/google-drive",
		`<h1>pidrive vs Google Drive for AI agents</h1>

<p>Google Drive is great for humans. You open a browser, drag files, click Share. But AI agents do not have browsers. They have terminals.</p>

<h2>The comparison</h2>

<table style="color:#aaa;font-size:14px;margin:16px 0;border-collapse:collapse;width:100%">
<tr style="border-bottom:1px solid #333;color:#888"><td style="padding:12px 16px 12px 0"></td><td style="padding:12px 8px"><strong style="color:#fff">pidrive</strong></td><td style="padding:12px 8px"><strong style="color:#fff">Google Drive</strong></td></tr>
<tr style="border-bottom:1px solid #1a1a1a"><td style="padding:8px 16px 8px 0">Access method</td><td style="padding:8px">Unix commands (ls, cat, grep)</td><td style="padding:8px">OAuth + REST API</td></tr>
<tr style="border-bottom:1px solid #1a1a1a"><td style="padding:8px 16px 8px 0">Auth</td><td style="padding:8px">API key</td><td style="padding:8px">OAuth2 flow (needs browser)</td></tr>
<tr style="border-bottom:1px solid #1a1a1a"><td style="padding:8px 16px 8px 0">File operations</td><td style="padding:8px">echo, cat, cp, rm, grep</td><td style="padding:8px">HTTP POST/GET with JSON</td></tr>
<tr style="border-bottom:1px solid #1a1a1a"><td style="padding:8px 16px 8px 0">Search</td><td style="padding:8px">pidrive search "query"</td><td style="padding:8px">API call with query params</td></tr>
<tr style="border-bottom:1px solid #1a1a1a"><td style="padding:8px 16px 8px 0">Share</td><td style="padding:8px">pidrive share file --link</td><td style="padding:8px">API call + permission model</td></tr>
<tr style="border-bottom:1px solid #1a1a1a"><td style="padding:8px 16px 8px 0">Agent isolation</td><td style="padding:8px">Built in, per agent</td><td style="padding:8px">Per Google account</td></tr>
<tr style="border-bottom:1px solid #1a1a1a"><td style="padding:8px 16px 8px 0">Pipes</td><td style="padding:8px">cat file | grep x | wc -l</td><td style="padding:8px">Not possible</td></tr>
<tr style="border-bottom:1px solid #1a1a1a"><td style="padding:8px 16px 8px 0">Setup time</td><td style="padding:8px">30 seconds</td><td style="padding:8px">OAuth app setup, scopes, tokens</td></tr>
<tr><td style="padding:8px 16px 8px 0">Built for</td><td style="padding:8px">AI agents</td><td style="padding:8px">Humans</td></tr>
</table>

<h2>The real issue: OAuth</h2>

<p>Google Drive requires OAuth2 authentication. That means:</p>

<ol>
<li>Register a Google Cloud project</li>
<li>Enable the Drive API</li>
<li>Create OAuth credentials</li>
<li>Implement the OAuth flow (redirect URL, token exchange, refresh tokens)</li>
<li>Store and rotate tokens</li>
<li>Handle scope consent screens</li>
</ol>

<p>All that before your agent can save a single file. With pidrive:</p>

<pre>pidrive register --email agent@company.com --name "My Agent" --server {{SERVER_URL}}
pidrive mount
echo "done" &gt; /drive/result.txt</pre>

<p>Three commands. No OAuth. No Google Cloud console. No browser.</p>

<h2>When to use Google Drive</h2>

<p>Google Drive is the right choice when:</p>
<ul>
<li>Humans need to view and edit files in a browser</li>
<li>You need Google Docs/Sheets/Slides integration</li>
<li>You are already deep in the Google Workspace ecosystem</li>
</ul>

<h2>When to use pidrive</h2>

<p>pidrive is the right choice when:</p>
<ul>
<li>AI agents need to read and write files programmatically</li>
<li>You want standard unix commands, not API calls</li>
<li>You need per-agent isolation without managing Google accounts</li>
<li>You want simple sharing (one command, get a URL)</li>
<li>You are running agents in containers, VMs, or CI/CD</li>
</ul>

<div class="cta">
<p><strong>Try pidrive:</strong></p>
<pre>curl -sSL {{SERVER_URL}}/install.sh | bash</pre>
</div>
`),
}
