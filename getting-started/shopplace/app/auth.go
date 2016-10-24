// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"encoding/gob"
	"errors"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/Chidi150/golang-samples/getting-started/shopplace"
	"github.com/satori/go.uuid"
	"google.golang.org/api/plus/v1"
)

const (
	defaultSessionID = "default"
	// The following keys are used for the default session. For example:
	//  session, _ := shopplace.SessionStore.New(r, defaultSessionID)
	//  session.Values[oauthTokenSessionKey]
	googleProfileSessionKey = "google_profile"
	oauthTokenSessionKey    = "oauth_token"

	// This key is used in the OAuth flow session to store the URL to redirect the
	// user to after the OAuth flow is complete.
	oauthFlowRedirectKey = "redirect"
)

func init() {
	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})
	gob.Register(&plus.Person{})
}

// loginHandler initiates an OAuth flow to authenticate the user.
func loginHandler(w http.ResponseWriter, r *http.Request) *appError {
	sessionID := uuid.NewV4().String()

	session, err := shopplace.SessionStore.New(r, sessionID)
	if err != nil {
		return appErrorf(err, "could not create oauth session: %v", err)
	}
	session.Options.MaxAge = 10 * 60 // 10 minutes: this time is for this session to persist
	//while the authentication process lasts

	//Required for Oauth2 callback url
	redirectURL, err := validateRedirectURL(r.FormValue("redirect"))
	if err != nil {
		return appErrorf(err, "invalid redirect URL: %v", err)
	}
	session.Values[oauthFlowRedirectKey] = redirectURL

	if err := session.Save(r, w); err != nil {
		return appErrorf(err, "could not save session: %v", err)
	}

	// Use the session ID for the "state" parameter.
	// This protects against CSRF (cross-site request forgery).
	// See https://godoc.org/golang.org/x/oauth2#Config.AuthCodeURL for more detail.

	url := shopplace.OAuthConfig.AuthCodeURL(sessionID, oauth2.ApprovalForce,
		oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
	return nil
}

// validateRedirectURL checks that the URL provided is valid.
// If the URL is missing, redirect the user to the application's root.
// The URL must not be absolute (i.e., the URL must refer to a path within this
// application).
func validateRedirectURL(path string) (string, error) {
	if path == "" {
		return "/", nil
	}

	// Ensure redirect URL is valid and not pointing to a different server.
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "/", err
	}
	if parsedURL.IsAbs() {
		return "/", errors.New("URL must be absolute")
	}
	return path, nil
}

// oauthCallbackHandler completes the OAuth flow, retreives the user's profile
// information and stores it in a session.
func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) *appError {
	oauthFlowSession, err := shopplace.SessionStore.Get(r, r.FormValue("state"))
	if err != nil {
		return appErrorf(err, "invalid state parameter. try logging in again.")
	}

	redirectURL, ok := oauthFlowSession.Values[oauthFlowRedirectKey].(string)
	// Validate this callback request came from the app.
	if !ok {
		return appErrorf(err, "invalid state parameter. try logging in again.")
	}

	code := r.FormValue("code")
	tok, err := shopplace.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return appErrorf(err, "could not get auth token: %v", err)
	}

	session, err := shopplace.SessionStore.New(r, defaultSessionID)
	if err != nil {
		return appErrorf(err, "could not get default session: %v", err)
	}

	ctx := context.Background()
	profile, err := fetchProfile(ctx, tok)
	if err != nil {
		return appErrorf(err, "could not fetch Google profile: %v", err)
	}

	session.Values[oauthTokenSessionKey] = tok
	// Strip the profile to only the fields we need. Otherwise the struct is too big.
	session.Values[googleProfileSessionKey] = stripProfile(profile)
	if err := session.Save(r, w); err != nil {
		return appErrorf(err, "could not save session: %v", err)
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
	return nil
}

// fetchProfile retrieves the Google+ profile of the user associated with the
// provided OAuth token.
func fetchProfile(ctx context.Context, tok *oauth2.Token) (*plus.Person, error) {
	client := oauth2.NewClient(ctx, shopplace.OAuthConfig.TokenSource(ctx, tok))
	plusService, err := plus.New(client)
	if err != nil {
		return nil, err
	}
	return plusService.People.Get("me").Do()
}

// logoutHandler clears the default session.
func logoutHandler(w http.ResponseWriter, r *http.Request) *appError {
	session, err := shopplace.SessionStore.New(r, defaultSessionID)
	if err != nil {
		return appErrorf(err, "could not get default session: %v", err)
	}
	session.Options.MaxAge = -1 // Clear session.
	if err := session.Save(r, w); err != nil {
		return appErrorf(err, "could not save session: %v", err)
	}
	redirectURL := r.FormValue("redirect")

	if redirectURL == "" || redirectURL == "/shops/mine" {
		redirectURL = "/"
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
	return nil
}

// profileFromSession retreives the Google+ profile from the default session.
// Returns nil if the profile cannot be retreived (e.g. user is logged out).
func profileFromSession(r *http.Request) *plus.Person {
	session, err := shopplace.SessionStore.Get(r, defaultSessionID)
	if err != nil {
		return nil
	}
	tok, ok := session.Values[oauthTokenSessionKey].(*oauth2.Token)
	if !ok || !tok.Valid() {
		return nil
	}
	profile, ok := session.Values[googleProfileSessionKey].(*plus.Person)
	if !ok {
		return nil
	}
	return profile
}

// stripProfile returns a subset of a plus.Person.
func stripProfile(p *plus.Person) *plus.Person {
	return &plus.Person{
		Id:          p.Id,
		DisplayName: p.DisplayName,
		Image:       p.Image,
		Etag:        p.Etag,
		Name:        p.Name,
		Url:         p.Url,
	}
}
