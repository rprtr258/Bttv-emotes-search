import glob
from json import load, dump

res = {}
files = glob.glob("_???.json")
print("NOT EMPTY RESULTS:", len(files), flush=True)
total = len(files)
current = 0
CHUNKS = 200
chunk_size = total // CHUNKS
for i in range(CHUNKS):
    files_chunk = files[i * chunk_size : (i + 1) * chunk_size]
    chunk = {}
    for f in files_chunk:
        with open(f, "r") as fd:
            json = load(fd)
            chunk = {**chunk, **json}
            current += 1
            print(f"{current/total*100:2.2f}%", flush=True, end="\r")
    res = {**res, **chunk}
print()
with open("data.json", "w") as fd:
    dump(res, fd)
