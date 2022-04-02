# py3

# export GOOS=windows
# export GOARCH=amd64
from anet.core import base

for (goos, arch) in [("windows", "amd64"), ("linux", "amd64"),("darwin", "amd64"),]:
    if goos == "windows":
        exe = "sshtest.exe"
    elif goos=="darwin":
        exe = "sshtest-darwin"
    else:
        exe = "sshtest"

    
    cp = base.catch("env", args=["GOOS="+goos, "GOARCH="+arch, "go", "build", "-o", exe])
    stderr = cp.stderr.strip()
    stdout = cp.stdout.strip()
    if stderr:
        print(stderr.decode("UTF-8"))
    if stdout:
        print(stdout.decode("UTF-8"))
