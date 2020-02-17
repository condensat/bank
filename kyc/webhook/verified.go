package webhook

import (
	"net/http"
)

func verified(w http.ResponseWriter, req *http.Request) {
	debugHook("verified", w, req)
}
