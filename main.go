package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/openidConnect"
)

const (
	templateDir = "templates"
	assetsDir   = "assets"
)

const (
	envKeyOpenIDConnectKey          = "OPENID_CONNECT_KEY"
	envKeyOpenIDConnectSecret       = "OPENID_CONNECT_SECRET"
	envKeyOpenIDConnectDiscoveryURL = "OPENID_CONNECT_DISCOVERY_URL"
	envKeySessionSecretAuthKeyOne   = "SESSION_SECRET_AUTH_KEY_ONE"
	envKeySessionSecretEncKeyOne    = "SESSION_SECRET_ENC_KEY_ONE"
)

const (
	maxAge = 86400 // 1 day
	isProd = false // Set to true when serving over https
)

var store *sessions.CookieStore

type templateHandler struct {
	once     sync.Once
	filename string
	template *template.Template
}

func (th *templateHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// Do is intended for initialization that must be run exactly once.
	th.once.Do(func() {
		// Join joins any number of path elements into a single path, separating them with an OS specific Separator.
		templatePath := filepath.Join(templateDir, th.filename)
		// Must is a helper that wraps a call to a function returning (*Template, error) and panics if the error is non-nil.
		th.template = template.Must(template.ParseFiles(templatePath))
	})

	data := make(map[string]interface{})
	data["Host"] = req.Host

	if userData, err := getUserFromSession(req); err == nil {
		data["UserData"] = userData
	}

	th.template.Execute(res, data)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found.")
	}

	authKeyOne := []byte(os.Getenv(envKeySessionSecretAuthKeyOne))
	encryptionKeyOne := []byte(os.Getenv(envKeySessionSecretEncKeyOne))

	store = sessions.NewCookieStore(authKeyOne, encryptionKeyOne)
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.MaxAge = maxAge
	store.Options.Path = "/"
	store.Options.Secure = isProd

	gothic.Store = store
	gothic.GetProviderName = func(req *http.Request) (string, error) {
		return "openid-connect", nil
	}
}

func main() {
	// OpenID Connect is based on OpenID Connect Auto Discovery URL (https://openid.net/specs/openid-connect-discovery-1_0-17.html)
	// Because the OpenID Connect provider initialize it self in the New(), it can return an error which should be handled
	openidConnect, err := openidConnect.New(
		os.Getenv(envKeyOpenIDConnectKey),
		os.Getenv(envKeyOpenIDConnectSecret),
		"http://localhost:3000/auth/openid-connect/callback",
		os.Getenv(envKeyOpenIDConnectDiscoveryURL))
	if err != nil {
		log.Fatal(err)
		return
	}

	goth.UseProviders(openidConnect)

	r := newRoom()
	//r.tracer = trace.New(os.Stdout)

	// The second parameter of httpHandle is a Handler interface,
	// which is defined as an interface with a single method ServeHTTP(http.ResponseWriter, *http.Request).
	// Our templateHandler type conforms to the Handler interface
	http.Handle("/", MustAuth(&templateHandler{filename: "chat.html"}))

	http.Handle("/login", &templateHandler{filename: "login.html"})

	http.HandleFunc("/auth/openid-connect", loginHandler)
	http.HandleFunc("/auth/openid-connect/callback", oauthCallbackHandler)

	http.Handle("/room", r)

	// Handler for local js and css assets
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir(assetsDir))))

	// get the room main loop going
	// we are running the room in a separate goroutine
	// so that the chatting operations occur in backgroun
	// and the main goroutine can run the web server
	go r.run()

	// ListenAndServe starts an HTTP server with a given address and handler.
	// The handler is usually nil, which means to use DefaultServeMux.
	// DefaultServeMux is a ServeMux (HTTP request multiplexer).
	// It matches the URL of each incoming request against a list of registered patterns
	// and calls the handler for the pattern that most closely matches the URL.
	// Handle and HandleFunc add handlers to DefaultServeMux.
	log.Println("Starting web server on localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
