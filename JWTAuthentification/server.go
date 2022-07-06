package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type Guid struct {
	Guid string `json:"guid"`
}

func main() {
	logrus.Info("Connection to database...")
	initDB()
	r := mux.NewRouter()
	r.Handle("/api/create-token", GetTokensHandler).Methods("POST")
	r.Handle("/api/refresh-token", RefreshTokenHandler).Methods("PUT")
	addr := flag.String("addr", ":8000", "localhost")
	srv := &http.Server{
		Addr:    *addr,
		Handler: r,
	}
	logrus.Info("Server started on port", *addr)
	err := srv.ListenAndServe()
	if err != nil {
		logrus.Error(err)
	}
}

var GetTokensHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Запрос на получение токена")
	body, _ := io.ReadAll(r.Body)
	var guid Guid

	err := json.Unmarshal(body, &guid)
	if err != nil {
		logrus.Error(err, "xddddd")
		return
	}

	SendTokenResponse(guid.Guid, &w, InsertRefreshToken)
})

var RefreshTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.Header().Add("Host", "localhost")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.Error(err)
	}
	token, err := DecodingJsonToken(body)
	if err != nil {
		logrus.Error(err)
		return
	}

	if token.Refresh == "" || token.Access == "" {
		logrus.Error(err)
		return
	}

	if claims, err := ParseVerifiedAccessToken(token.Access); claims == nil || err != nil {
		logrus.Error(err)
		return
	} else {
		if err := RefreshTokenValidate(claims.Guid, token.Refresh); err == nil {
			SendTokenResponse(claims.Guid, &w, UpdateRefreshToken)
		} else {
			logrus.Error(err)
		}
	}
})

func SendTokenResponse(guid string, w *http.ResponseWriter, query func(string, string) error) {
	if guid == "" {
		logrus.Error("GUID отсутствует")
		return
	}

	access, err := GetNewAccessToken(guid)
	if err != nil {
		logrus.Error(err, "xdddd")
		return
	}

	refresh, err := CreateRefreshToken(guid, query)
	if err != nil {
		logrus.Error(err)
		return
	}

	response, err := TokenEncodeToJson(Tokens{Status: 1, Access: access, Refresh: refresh, Guid: guid})
	(*w).WriteHeader(http.StatusCreated)
	_, err = (*w).Write(response)
	logrus.Info("Токен успешно сгенерирован")
}
