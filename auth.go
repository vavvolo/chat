package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"strings"

	"github.com/markbates/goth/gothic"
)

const (
	sessionName    = "GO_CHAT_SESSION"
	sessionUserKey = "USER"
)

func init() {
	// Register the type with gob, otherwise gorilla/sessions
	// won't be able to save the data in the cookie
	gob.Register(map[string]interface{}{})
}

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

	userData := map[string]interface{}{
		"UserID":   user.UserID,
		"FullName": strings.Join([]string{user.FirstName, user.LastName}, " "),
	}

	err = saveUserInSession(res, req, userData)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// user authenticated successfully, redirect to /chat
	res.Header().Set("Location", "/chat")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func logoutHandler(res http.ResponseWriter, req *http.Request) {
	err := gothic.Logout(res, req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	err = clearUserInSession(res, req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", "/login")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func saveUserInSession(res http.ResponseWriter, req *http.Request, userData map[string]interface{}) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(req, sessionName)

	session.Values[sessionUserKey] = userData
	err := session.Save(req, res)
	if err != nil {
		return err
	}

	return nil
}

func getUserFromSession(req *http.Request) (map[string]interface{}, error) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(req, sessionName)

	v, ok := session.Values[sessionUserKey]
	if !ok {
		return nil, fmt.Errorf("missing sessionUserKey")
	}

	userData, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error converting sessionUserKey data to valid user identifier")
	}

	return userData, nil
}

func clearUserInSession(res http.ResponseWriter, req *http.Request) error {
	session, _ := store.Get(req, sessionName)

	// Setting MaxAge to -1 indicates that the cookie
	// should be deleted immediately by the browser.
	// Not all browsers are forced to delete the cookie,
	// which is why we also set an empty map in session.Values,
	// removing the previously stored data.
	session.Options.MaxAge = -1
	session.Values = make(map[interface{}]interface{})
	err := session.Save(req, res)
	if err != nil {
		return err
	}

	return nil
}
