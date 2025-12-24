package socket.io

import "go.k6.io/k6/js/modules"

const importPath = "k6/x/socket.io"

func init() {
	modules.Register(importPath, new(rootModule))
}
