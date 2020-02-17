package webhook

import (
	"net/http"
)

func revoked(w http.ResponseWriter, req *http.Request) {
	debugHook("revoked", w, req)
}
