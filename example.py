import socket
import json
import urllib3
from concurrent.futures import ThreadPoolExecutor
from threading import local

HN_API = "https://hacker-news.firebaseio.com/v0/item/{}.json"
DB_HOST = "127.0.0.1"
DB_PORT = 26227
DB_NAME = "hackernews"

http = urllib3.PoolManager()
thread_local = local()


def get_client():
    if not hasattr(thread_local, "client"):
        thread_local.client = RoamDB(DB_NAME)
    return thread_local.client


class RoamDB:
    def __init__(self, database, host=DB_HOST, port=DB_PORT):
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

def setup():
    db = RoamDB(DB_NAME)
    db.query("""
        CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY,
            type TEXT,
            by TEXT,
            title TEXT,
            text TEXT,
            url TEXT,
            score INTEGER,
            time INTEGER,
            descendants INTEGER,
            raw TEXT
        )
    """)
    db.close()


def fetch_and_insert(item_id):
    try:
        resp = http.request("GET", HN_API.format(item_id), timeout=10)
        item = json.loads(resp.data.decode())
        if not item:
            return
        db = get_client()
        db.query(
            "INSERT OR REPLACE INTO items (id, type, by, title, text, url, score, time, descendants, raw) VALUES (?,?,?,?,?,?,?,?,?,?)",
            [
                item.get("id"),
                item.get("type"),
                item.get("by"),
                item.get("title"),
                item.get("text"),
                item.get("url"),
                item.get("score"),
                item.get("time"),
                item.get("descendants"),
                json.dumps(item),
            ]
        )
        print(f"[ok] {item_id} - {item.get('type')} by {item.get('by')}")
    except Exception as e:
        print(f"[err] {item_id}: {e}")

setup()
with ThreadPoolExecutor(max_workers=100) as pool:
    pool.map(fetch_and_insert, range(1, 1001))
