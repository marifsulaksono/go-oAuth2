package controller

import (
	"g-oAuth2/domain"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
)

func LoginGoogle(w http.ResponseWriter, r *http.Request) {
	URL, err := url.Parse(domain.OAuthGoogleConf.Endpoint.AuthURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set required parameters
	parameters := url.Values{}
	parameters.Add("client_id", domain.OAuthGoogleConf.ClientID)
	parameters.Add("scope", strings.Join(domain.OAuthGoogleConf.Scopes, " "))
	parameters.Add("redirect_uri", domain.OAuthGoogleConf.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", domain.OAuthStateString)
	URL.RawQuery = parameters.Encode()
	url := URL.String()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackGoogle(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	log.Println("New Callback from Google OAuth :\n" + state)
	if state != domain.OAuthStateString {
		log.Printf("Invalid state. expected %s, got %s\n", domain.OAuthStateString, state)
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
		token, err := domain.OAuthGoogleConf.Exchange(oauth2.NoContext, code)
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
}
