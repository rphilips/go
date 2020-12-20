import json
from anet.core import base


def showitems(D, prefix=""):
    for a in "type", "description", "default", "uniqueItems", "enum":
        if a in D:
            value = D[a]
            if a in ["default", "enum"]:
                value = json.dumps(value)
            if a == "type":
                value = "*"+value+"*"
            print(prefix+"-", "`" + a+"`:", value)
    for aspect in ["items"]:
        if aspect in D:
            print(prefix+"-", "`" + aspect+"`:")
            DD = D[aspect]
            for asp in "type", "description", "enum":
                if asp not in DD:
                    continue
                if asp in DD:
                    value = DD[asp]
                if aspect in ["default", "enum"]:
                    value = json.dumps(value)
                if aspect == "type":
                    value = "*"+value+"*"
                print(prefix+"  -", asp+":", value)
    for aspect in ["properties", "value"]:
        if aspect not in D:
            continue
        E = D[aspect]
        showitems(E, prefix+"  ")


def markdown(fname):
    S = base.fetch(fname, nature="json")
    P = sorted(S["properties"])
    for p in P:
        print("###", p)
        D = S["properties"][p]
        showitems(D)
        print()

    print()
    print()
    for p in P:
        print(p)


if __name__ == "__main__":
    markdown("qtechng.schema.json")
