from itertools import product
from string import ascii_lowercase, digits
from collections import defaultdict
from json import dump
import asyncio
import os

import aiohttp


LIMIT = 100


def emote_gifs_urls_by_id(id):
    return f"https://cdn.betterttv.net/emote/{id}/1x", \
           f"https://cdn.betterttv.net/emote/{id}/2x", \
           f"https://cdn.betterttv.net/emote/{id}/3x"

def emote_query_url(query, offset):
    return f"https://api.betterttv.net/3/emotes/shared/search?query={query}&offset={offset}&limit={LIMIT}"

async def find_emotes(lock, cnt, total, session, query):
    filename = f"_{query}.json"
    if os.path.isfile(filename):
        async with lock:
            cnt[0] += 1
            print(f"{cnt[0]/total*100:.2f}%", flush=True, end="\r")
        return
    result = defaultdict(list)
    offset = 0
    while True:
        response = await session.get(emote_query_url(query, offset))
        if response.headers["content-type"] == "application/json; charset=utf-8":
            json = await response.json()
            for x in json:
                result[x["code"]].append(x["id"] )
            if len(json) < 100:
                break
            offset += LIMIT
        else:
            await asyncio.sleep(10)
    async with lock:
        cnt[0] += 1
        print(f"{cnt[0]/total*100:.2f}%", flush=True, end="\r")
        filename = f"_{query}.json"
        with open(filename, "w") as fd:
            dump(result, fd)

async def main():
    alphabet = ascii_lowercase + digits + "'"
    base_queries = ["".join(x) for x in product(alphabet, repeat=3)]
    async with aiohttp.ClientSession() as session:
        lock = asyncio.Lock()
        cnt = [0]
        total = len(base_queries)
        await asyncio.gather(*[find_emotes(lock, cnt, total, session, query) for query in base_queries])
        print()

loop = asyncio.get_event_loop()
loop.run_until_complete(main())
