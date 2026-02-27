# 🧳 RoamDB

> *Your database, but make it wanderlust.*

RoamDB is a lightweight SQLite server that lets your data **pack its bags and go**. Each database is just a plain `.db` file — throw it in a backpack, email it to your future self, or copy it to a USB stick and dramatically hand it to someone in a parking garage. Your data goes where you go.

It speaks raw TCP (because HTTP has too much *baggage* 😏), handles concurrent writes without corrupting everything, and asks zero questions about who you are.

---

```
  client 1 ──┐
  client 2 ──┼──► TCP :26227 ──► [myapp worker] ──► myapp.db
  client 3 ──┘                 ↘ [analytics worker] ──► analytics.db
```

No ORM. No migrations framework. No cloud vendor lock-in. Just vibes and SQLite.

---

## ✈️ How It Works

- Every database is a **SQLite file** living in `./data/`
- Each DB gets its own **goroutine worker** — writes serialize automatically, no data corruption, no locks exploding in your face
- Two different databases? Written to **in parallel**, no problem
- Transport is **raw TCP with newline-delimited JSON** — fast, simple, and your grandmother could implement a client
- **Zero authentication** — security is your problem, bestie

---

## 🚀 Getting Started

### Run the server

```bash
cd server
go mod tidy
go run main.go
# Listening on :26227
```

Databases are created on demand in `./data/`. No setup. No `CREATE DATABASE`. Just ask for one and it appears, like room service.

### Protocol

Send a JSON object, get a JSON object back. That's it. That's the whole protocol.

**Request:**
```json
{ "database": "myapp", "query": "SELECT * FROM users WHERE id = ?", "params": [42] }
```

**Response:**
```json
{ "columns": ["id", "name"], "rows": [{ "id": 42, "name": "Alice" }] }
```

**Error response:**
```json
{ "error": "no such table: users" }
```

The connection is persistent — send as many queries as you want before closing it. Your TCP connection is a long-haul flight, not a taxi ride.

---

## 🐍 Python Client

```python
import socket
import json

class RoamDB:
    def __init__(self, database, host='127.0.0.1', port=26227):
        self.database = database
        self.sock = socket.create_connection((host, port))
        self.f = self.sock.makefile("rwb")

    def query(self, sql, params=None):
        req = {"database": self. database, "query": sql, "params": params or []}
        self.f.write((json.dumps(req) + "\n").encode())
        self.f.flush()
        return json.loads(self.f.readline())

    def close(self):
        self.sock.close()

db = RoamDB("myapp")
db.query("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)")
db.query("INSERT INTO users (name) VALUES (?)", ["Alice"])
print(db.query("SELECT * FROM users"))
db.close()
```

No dependencies. Not even `requests`. Pure stdlib. The way the universe intended.

---

## 🗞️ Example: Scraping HackerNews into a portable DB

See [`example.py`](./example.py) for a fully working example that fetches HackerNews items 1–1000 with **100 concurrent workers** and inserts them into a local SQLite file.

```bash
python example.py
```

When it's done, you have a single `data/hackernews.db` file. Copy it anywhere. Open it with any SQLite browser. Email it to a friend. The data is yours, fully portable, no cloud account required.

---

## 🗺️ Portability: The Whole Point

Most databases make your data a **hostage**. Dumps, exports, migrations, vendor formats — it's a whole thing.

RoamDB's entire philosophy: **your data is a file**. 

| Scenario | What you do |
|---|---|
| Move to a new server | `cp server/data/ new-server:server/data/` |
| Back up everything | `tar -czf backup.tar.gz server/data/` |
| Share a dataset | Send the `.db` file |
| Inspect the data | Open with DB Browser for SQLite |
| Escape RoamDB entirely | It's just SQLite. Open it with anything. |

---

## 🏗️ Project Structure

```
.
├── server/
│   ├── main.go             # The whole server. One file. ~100 lines.
│   └── data/               # Your databases live here (created on first run)
│       ├── myapp.db
│       └── hackernews.db
│── example.py              # HackerNews scraper demo
```

---

## ⚠️ Caveats & Honest Confessions

- **No authentication.** Bind to localhost or put it behind a firewall. Or don't, and live dangerously.
- **No TLS.** Same as above.
- **SQLite isn't Postgres.** It's SQLite. You know what you signed up for.
- **Goroutines for databases are never cleaned up.** If you create 10,000 databases, you'll have 10,000 goroutines. RoamDB trusts you to not do that.

---

## 📦 Dependencies

```
github.com/mattn/go-sqlite3
```

That's it. One dependency. It's a C binding so you need `gcc`. On most systems this just works. On Windows it's a fun adventure.

---

## 🌍 Philosophy

> *"Data portability isn't a feature. It's a right."*

RoamDB exists because your data should travel as freely as you do. SQLite files are the most portable database format on earth — they'll still open in 30 years on hardware that doesn't exist yet. RoamDB just puts a concurrent TCP server in front of them so your apps can use them without tripping over each other.

Pack light. Move fast. Own your data.

---

*Made with ☕, vibes, and a deep suspicion of managed databases.*
