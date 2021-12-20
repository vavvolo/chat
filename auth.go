package main

import (
	"fmt"
	"net/http"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

const (
	sessionName      = "GO_CHAT_SESSION"
	sessionUserIDKey = "USER_ID"
)

type authHandler struct {
	nextHandler http.Handler
}

func MustAuth(h http.Handler) http.Handler {
	return &authHandler{nextHandler: h}
}

func (ah *authHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if _, err := getUserFromSession(req); err != nil {
		// user is not authenticated, redirect to /login
		res.Header().Set("Location", "/login")
		res.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	// all good, call the next handler
	ah.nextHandler.ServeHTTP(res, req)
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
	gothic.BeginAuthHandler(res, req)
}

func oauthCallbackHandler(res http.ResponseWriter, req *http.Request) {
	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	err = saveUserInSession(res, req, user)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// user authenticated successfully, redirect to /chat
	res.Header().Set("Location", "/chat")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func saveUserInSession(res http.ResponseWriter, req *http.Request, u goth.User) error {
	session, err := gothic.Store.Get(req, sessionName)
	if err != nil {
		return err
	}

	session.Values[sessionUserIDKey] = u.UserID
	err = session.Save(req, res)
	if err != nil {
		return err
	}

	return nil
}

func getUserFromSession(req *http.Request) (string, error) {
	session, err := gothic.Store.Get(req, sessionName)
	if err != nil {
		return "", err
	}

	v, ok := session.Values[sessionUserIDKey]
	if !ok {
		return "", fmt.Errorf("missing sessionUserKey")
	}

	userID, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("error converting sessionUserKey data to valid user identifier")
	}

	return userID, nil
}
