package api

import (
	"net/http"
	"os"
	"strings"
)

func (s *Server) serveLanding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := strings.ReplaceAll(landingHTMLTemplate, "{{SERVER_URL}}", s.cfg.ServerURL)
	w.Write([]byte(html))
}

func (s *Server) serveSkillMD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	md := strings.ReplaceAll(skillMDTemplate, "{{SERVER_URL}}", s.cfg.ServerURL)
	w.Write([]byte(md))
}

func (s *Server) serveRelease(w http.ResponseWriter, r *http.Request) {
	// Serve binary from bin/releases/
	name := r.URL.Path[len("/releases/"):]
	// Sanitize
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") {
		http.NotFound(w, r)
		return
	}
	path := "bin/releases/" + name
	f, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	stat, _ := f.Stat()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=pidrive")
	http.ServeContent(w, r, name, stat.ModTime(), f)
}

func (s *Server) serveInstallScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	script := strings.ReplaceAll(installScriptTemplate, "{{SERVER_URL}}", s.cfg.ServerURL)
	w.Write([]byte(script))
}

var skillMDTemplate = `---
name: pidrive
description: Private file storage for AI agents. Store, read, write, search, and share files backed by S3. Mount via SFTP and use standard unix commands (ls, cat, grep, cp, rm). No local storage — files live in S3.
---

# pidrive

Private file storage for AI agents. Files are stored in AWS S3 and accessed via SFTP mount. No files are stored locally — the mount point is a tunnel to the server.

## Install

curl -sSL {{SERVER_URL}}/install.sh | bash

This installs the pidrive CLI and sshfs (needed for mounting).

## Get started

pidrive register --email you@company.com --name "My Agent" --server {{SERVER_URL}}
pidrive verify --email you@company.com --code <check-email>
pidrive mount

After mount, your files are at /drive/. Use standard unix commands:

ls /drive/
echo "hello" > /drive/notes.txt
cat /drive/notes.txt
grep -r "error" /drive/logs/
head -20 /drive/data.csv
wc -l /drive/data.csv
cp local-file.txt /drive/
mkdir /drive/output/
rm /drive/old-file.txt

Every read/write goes through SFTP to the server, then to S3. Nothing is cached locally. If the VM dies, nothing is lost.

## Share

pidrive share report.txt --to other-agent@company.com
pidrive share data.csv --link
pidrive share data.csv --link --expires 7d
pidrive shared
pidrive revoke <share-id>
pidrive pull <share-url> [destination]

Link shares produce a public URL: {{SERVER_URL}}/s/<id>
Anyone with the URL can download the file. No auth needed.

## Search

pidrive search "quarterly revenue"

Full-text search across all your files. The server indexes text files in the background.

## Trash

pidrive trash
pidrive restore <path>

Deleted files are kept for 30 days.

## Account

pidrive whoami
pidrive status
pidrive usage
pidrive plans
pidrive upgrade --plan pro
pidrive login --email you@company.com
pidrive unmount

## Plans

free: 1 GB storage, 100 MB bandwidth/month
pro:  100 GB storage, 10 GB bandwidth/month — $5/mo
team: 1 TB storage, unlimited bandwidth — $20/mo

## Activity

pidrive activity

Shows mount, share, revoke, restore events.

## How it works

1. pidrive mount runs sshfs under the hood
2. sshfs connects to the pidrive SFTP server (port 2022) using your API key
3. The server maps your session to your directory on a JuiceFS filesystem
4. JuiceFS stores file data in AWS S3 and metadata in Redis
5. Each agent is isolated — you can only see your own files
6. Sharing copies files between agent directories on the server
`

var installScriptTemplate = `#!/bin/bash
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
esac

PIDRIVE_SERVER="{{SERVER_URL}}"

echo "Installing pidrive for $OS/$ARCH..."

# Install sshfs (needed for mounting)
if ! command -v sshfs &>/dev/null; then
  echo "Installing sshfs..."
  if [ "$OS" = "linux" ]; then
    if command -v apt &>/dev/null; then
      sudo apt update && sudo apt install -y sshfs
    elif command -v yum &>/dev/null; then
      sudo yum install -y fuse-sshfs
    fi
  elif [ "$OS" = "darwin" ]; then
    if ! test -d /Library/Filesystems/macfuse.fs; then
      echo "macFUSE required. Install from: https://osxfuse.github.io/"
      exit 1
    fi
    brew tap gromgit/fuse 2>/dev/null
    brew install gromgit/fuse/sshfs-mac 2>/dev/null || echo "Install sshfs: brew install gromgit/fuse/sshfs-mac"
  fi
fi

# Install pidrive CLI
echo "Installing pidrive CLI..."
curl -sSLo /usr/local/bin/pidrive \
  "${PIDRIVE_SERVER}/releases/pidrive-${OS}-${ARCH}"
chmod +x /usr/local/bin/pidrive

echo ""
echo "pidrive installed!"
echo ""
echo "Next steps:"
echo "  pidrive register --email you@company.com --name \"My Agent\" --server ${PIDRIVE_SERVER}"
echo "  pidrive mount"
echo "  ls /drive/"
`

var landingHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>pidrive</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#0a0a0a;color:#e0e0e0;font-family:"SF Mono","Fira Code",Menlo,monospace;font-size:15px;line-height:1.7;padding:60px 24px;max-width:680px;margin:0 auto}
h1{color:#fff;font-size:28px;font-weight:700;margin-bottom:6px}
.tl{color:#888;font-size:16px;margin-bottom:48px}
h2{font-size:14px;font-weight:600;text-transform:uppercase;letter-spacing:1.5px;margin-top:48px;margin-bottom:16px;color:#666}
.p{color:#aaa;font-size:16px;margin-bottom:40px;line-height:1.8}
.p strong{color:#fff}
pre{background:#141414;border:1px solid #222;border-radius:8px;padding:20px 24px;overflow-x:auto;margin-bottom:12px;font-size:14px;line-height:1.8}
.c{color:#7ec699}.f{color:#cc99cd}.s{color:#f8c555}.o{color:#6cb6ff}.d{color:#555}
.pl{display:flex;gap:16px;margin-top:16px;flex-wrap:wrap}
.pc{background:#141414;border:1px solid #222;border-radius:8px;padding:20px 24px;flex:1;min-width:180px}
.pn{color:#fff;font-weight:700;font-size:16px}
.pp{color:#7ec699;font-size:14px;margin:4px 0 8px}
.pd{color:#777;font-size:13px}
.lk{margin-top:48px;padding-top:24px;border-top:1px solid #1a1a1a;color:#555;font-size:13px}
.lk a{color:#6cb6ff;text-decoration:none}
.lk a:hover{text-decoration:underline}
.nt{margin-top:24px;color:#999;font-size:12px;line-height:1.6}
.nt code{color:#bbb;background:#1a1a1a;padding:2px 5px;border-radius:3px;font-size:12px}
</style>
</head>
<body>

<h1>pidrive</h1>
<p class="tl">file storage for AI agents</p>

<p class="p">
Your agents need files. S3 is an API.<br>
<strong>pidrive makes it a filesystem.</strong>
</p>

<h2>Install</h2>
<pre><span class="c">curl</span> <span class="f">-sSL</span> <span class="s">{{SERVER_URL}}/install.sh</span> | <span class="c">bash</span></pre>

<h2>Get started</h2>
<pre><span class="c">pidrive</span> <span class="f">register</span> <span class="d">--email</span> <span class="s">agent@company.com</span> <span class="d">--name</span> <span class="s">"My Agent"</span> <span class="d">--server</span> <span class="s">{{SERVER_URL}}</span>
<span class="c">pidrive</span> <span class="f">verify</span> <span class="d">--email</span> <span class="s">agent@company.com</span> <span class="d">--code</span> <span class="s">&lt;check-email&gt;</span>
<span class="c">pidrive</span> <span class="f">mount</span></pre>

<h2>Use unix on S3</h2>
<pre><span class="c">echo</span> <span class="s">"quarterly report"</span> &gt; /drive/report.txt
<span class="c">grep</span> <span class="f">-r</span> <span class="s">"error"</span> /drive/logs/
<span class="c">cat</span> /drive/data.csv | <span class="c">head</span> <span class="f">-20</span>
<span class="c">ls</span> <span class="f">-la</span> /drive/
<span class="c">cp</span> report.pdf /drive/</pre>

<h2>Share with a URL</h2>
<pre><span class="c">pidrive</span> <span class="f">share</span> data.csv <span class="d">--link</span>
<span class="o">&rarr; {{SERVER_URL}}/s/vxi4g6bu</span>

<span class="c">pidrive</span> <span class="f">share</span> report.txt <span class="d">--to</span> <span class="s">other-agent@company.com</span></pre>

<h2>Search</h2>
<pre><span class="c">pidrive</span> <span class="f">search</span> <span class="s">"quarterly revenue"</span>
<span class="o">  files/report.txt   Q4 2024 quarterly revenue up 340%</span>
<span class="o">  files/data.csv     quarterly revenue breakdown by region</span></pre>

<h2>Plans</h2>
<div class="pl">
<div class="pc"><div class="pn">Free</div><div class="pp">$0/mo</div><div class="pd">1 GB storage<br>100 MB bandwidth</div></div>
<div class="pc"><div class="pn">Pro</div><div class="pp">$5/mo</div><div class="pd">100 GB storage<br>10 GB bandwidth</div></div>
<div class="pc"><div class="pn">Team</div><div class="pp">$20/mo</div><div class="pd">1 TB storage<br>Unlimited bandwidth</div></div>
</div>

<div class="lk">
<a href="{{SERVER_URL}}/skill.md">skill.md</a> &middot;
<a href="{{SERVER_URL}}/install.sh">install.sh</a> &middot;
<a href="{{SERVER_URL}}/api/plans">API</a>
</div>

<p class="nt">macOS note: sshfs requires macFUSE. Run <code>brew install --cask macfuse</code>, approve the kernel extension in System Settings &rarr; Privacy &amp; Security &rarr; Enable System Extensions, reboot, then <code>brew install gromgit/fuse/sshfs-mac</code>. One-time setup. Ubuntu works out of the box.</p>

</body>
</html>`
