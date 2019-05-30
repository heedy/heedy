import json

print("Running PYTHON test exec")
v = json.loads(input())

# print("finished reading input:", v)

with open(v["data_dir"] + "/test.txt", "w") as f:
    f.write(json.dumps(v))
