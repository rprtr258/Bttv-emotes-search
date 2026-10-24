from json import load
from sys import argv


if len(argv) == 1:
    print("Usage: python query.py [query]")
    exit(1)

if len(argv) > 2:
    print("Too many arguments")
    exit(2)

query = argv[1]

with open("data.json", "r") as fd:
    db = load(fd)

for emote, len_ids in sorted([(k, len(v)) for k, v in db.items() if query in k.lower()], key=lambda x:-x[1]):
    print(f"{emote:20s}", len_ids)

