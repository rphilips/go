# py3

from anet.core import base

cp = base.catch("go", args=["build", "-o", "toolcatgo3", "."])
stderr = cp.stderr.strip()
stdout = cp.stdout.strip()
if stderr:
    print(stderr.decode("UTF-8"))
if stdout:
    print(stdout.decode("UTF-8"))

    