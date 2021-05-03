# py3

# export GOOS=windows
# export GOARCH=amd64
from anet.core import base

cp = base.catch("go", args=["build", "-o", ".", "-ldflags", "-H=windowsgui"])
stderr = cp.stderr.strip()
stdout = cp.stdout.strip()
if stderr:
    print(stderr.decode("UTF-8"))
if stdout:
    print(stdout.decode("UTF-8"))

    
