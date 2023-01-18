package tools

import (
	bfs "brocade.be/base/fs"
	pregistry "brocade.be/pbladng/lib/registry"
)

func Log(data any) {
	logfile := pregistry.Registry["log-file"].(string)
	bfs.Store(logfile, data, "process")
}
