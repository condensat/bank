package handlers

import (
	"sync"
)

var ShellExecLock sync.Mutex
