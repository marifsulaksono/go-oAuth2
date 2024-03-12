package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const PORT = ":8080"

func main() {
	// load .env files
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error load .env files : %v", err)
	}

	// create new oauth2 config
	var (
		OAuthGoogleConf = &oauth2.Config{
			ClientID:     os.Getenv("CLIENT_ID"),
			ClientSecret: os.Getenv("CLIENT_SECRET"),
			RedirectURL:  os.Getenv("REDIRECT_URL"),
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		}

		OAuthStateString = os.Getenv("STATE_STRING")
	)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome, please login first"))
	}).Methods(http.MethodGet)

	r.HandleFunc("/auth/google-auth", func(w http.ResponseWriter, r *http.Request) {
		URL, err := url.Parse(OAuthGoogleConf.Endpoint.AuthURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// set required parameters
		parameters := url.Values{}
		parameters.Add("client_id", OAuthGoogleConf.ClientID)
		parameters.Add("scope", strings.Join(OAuthGoogleConf.Scopes, " "))
		parameters.Add("redirect_uri", OAuthGoogleConf.RedirectURL)
		parameters.Add("response_type", "code")
		parameters.Add("state", OAuthStateString)
		URL.RawQuery = parameters.Encode()
		url := URL.String()
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}).Methods(http.MethodGet)

	r.HandleFunc("/auth/google-callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.FormValue("state")
		log.Println("New Callback from Google OAuth :\n" + state)
		if state != OAuthStateString {
			log.Printf("Invalid state. expected %s, got %s\n", OAuthStateString, state)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		code := r.FormValue("code")
		if code == "" {
			log.Println("Code not found")
			reason := r.FormValue("error_reason")
			if reason == "user_denied" {
				w.Write([]byte("User permission has denied"))
				return
			}

			w.Write([]byte("Code Not Found to provide AccessToken"))
		} else {
			token, err := OAuthGoogleConf.Exchange(oauth2.NoContext, code)
			if err != nil {
				log.Printf("OAuth Exchange failed : %v\n", err)
				return
			}

			log.Printf("[TOKEN_AUTH]Access Token : %s", token.AccessToken)
			log.Printf("[TOKEN_AUTH]Expiry Token : %s", token.Expiry.String())
			log.Printf("[TOKEN_AUTH]Refresh Token : %s", token.RefreshToken)

			response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
			if err != nil {
				log.Printf("Error Get Response : %v", err)
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
			defer response.Body.Close()

			bodyResponse, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Printf("Error Read Body Response : %v", err)
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}

			w.Write([]byte(string(bodyResponse)))
			return
		}
	})

	log.Printf("Server start at %s", PORT)
	err := http.ListenAndServe(PORT, r)
	if err != nil {
		log.Fatalf("Error start server : %v", err)
	}
}
