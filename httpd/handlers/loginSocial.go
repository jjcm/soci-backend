package handlers

import (
	"net/http"
	"os"

	"gopkg.in/danilopolani/gocialite.v1"
)

var gocial = gocialite.NewDispatcher()

// LoginSocial try and authenticate a user over oauth
func LoginSocial(w http.ResponseWriter, r *http.Request) {
	// other providers documented here
	// https://github.com/danilopolani/gocialite/wiki/Multi-provider-example

	providerSecrets := map[string]map[string]string{
		"github": {
			"clientID":     os.Getenv("OAUTH_ID"),
			"clientSecret": os.Getenv("OAUTH_SECRET"),
			"redirectURL":  "http://localhost:9000/login-social/callback",
		},
	}
	providerScopes := map[string][]string{
		"github": []string{""},
	}

	providerData := providerSecrets["github"]
	actualScopes := providerScopes["github"]
	authURL, err := gocial.New().
		Driver("github").
		Scopes(actualScopes).
		Redirect(
			providerData["clientID"],
			providerData["clientSecret"],
			providerData["redirectURL"],
		)

	if err != nil {
		Log.Error(err.Error())
		return
	}

	http.Redirect(w, r, authURL, http.StatusSeeOther)
}

// LoginSocialCallback callback for oauth app
func LoginSocialCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.FormValue("code")
	state := r.FormValue("state")

	// Handle callback and check for errors
	user, _, err := gocial.Handle(state, code)
	if err != nil {
		Log.Error(err.Error())
		return
	}

	token, err := tokenCreator(user.Email)
	if err != nil {
		SendResponse(w, MakeError("There was an error signing your JWT token: "+err.Error()), 500)
		return
	}

	response := map[string]string{
		"token": token,
	}
	SendResponse(w, response, 200)
}
