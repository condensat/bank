package webhook

import (
	"net/http"
)

func ready(w http.ResponseWriter, req *http.Request) {
	debugHook("ready", w, req)
}
