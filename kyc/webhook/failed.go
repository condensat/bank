package webhook

import (
	"net/http"
)

func failed(w http.ResponseWriter, req *http.Request) {
	debugHook("failed", w, req)
}
