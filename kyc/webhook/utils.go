package webhook

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"git.condensat.tech/bank/logger"
	"github.com/sirupsen/logrus"
)

func BadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "400 bad request", http.StatusBadRequest)
}

func SecretsFromContext(ctx context.Context) Secrets {
	if secrets, ok := ctx.Value(KeySynapsSecrets).(*Secrets); ok {
		return *secrets
	}
	return Secrets{}
}

func toLogrusFields(dic map[string][]string) logrus.Fields {
	fields := logrus.Fields{}
	for key, value := range dic {
		fields[key] = value
	}
	return fields
}

func bodyToMap(reader io.ReadCloser) (map[string]interface{}, []byte) {
	result := make(map[string]interface{})

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return result, nil
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		result["Body"] = string(body)
	}

	return result, body
}

func logRequestFields(log *logrus.Entry, req *http.Request) ([]byte, *logrus.Entry) {
	log = log.WithFields(toLogrusFields(req.Header))
	log = log.WithFields(toLogrusFields(req.URL.Query()))

	bodyMap, body := bodyToMap(req.Body)
	return body, log.WithFields(bodyMap)
}

func debugHook(hookName string, w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := logger.Logger(ctx).WithField("Method", "kyc.webhook.debugHook")
	secrets := SecretsFromContext(ctx)

	log = log.WithField("HookName", hookName)
	_, log = logRequestFields(log, req)

	secret := req.FormValue("secret")
	if secret != secrets.Get(hookName) {
		log.WithField("ExpectedSecret", secrets.Get(hookName)).
			Error("Hook secret missmatch")
		http.Error(w, "Wrong secret", http.StatusBadRequest)
		return
	}

	// Prepare response
	response := make(map[string]interface{})
	response["result"] = "success"

	js, err := json.Marshal(response)
	if err != nil {
		log.WithError(err).
			Error("Failed to Marshal response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		log.WithError(err).
			Error("Failed to Write response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debugf("WebHook called")
}
