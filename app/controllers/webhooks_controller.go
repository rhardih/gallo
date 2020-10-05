package controllers

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"gallo/lib"
	"net/http"
	"time"
)

var trelloKey string

func init() {
	trelloKey = lib.MustGetEnv("TRELLO_KEY")
}

type Webhook struct {
	Action struct {
		ID              string `json:"id"`
		IDMemberCreator string `json:"idMemberCreator"`
		Data            struct {
			Board struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"board"`
			Card struct {
				IDShort int    `json:"idShort"`
				Name    string `json:"name"`
				ID      string `json:"id"`
			} `json:"card"`
			Voted bool `json:"voted"`
		} `json:"data"`
		Type          string    `json:"type"`
		Date          time.Time `json:"date"`
		MemberCreator struct {
			ID         string `json:"id"`
			AvatarHash string `json:"avatarHash"`
			FullName   string `json:"fullName"`
			Initials   string `json:"initials"`
			Username   string `json:"username"`
		} `json:"memberCreator"`
	} `json:"action"`
	Model struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Desc           string `json:"desc"`
		Closed         bool   `json:"closed"`
		IDOrganization string `json:"idOrganization"`
		Pinned         bool   `json:"pinned"`
		URL            string `json:"url"`
		Prefs          struct {
			PermissionLevel string `json:"permissionLevel"`
			Voting          string `json:"voting"`
			Comments        string `json:"comments"`
			Invitations     string `json:"invitations"`
			SelfJoin        bool   `json:"selfJoin"`
			CardCovers      bool   `json:"cardCovers"`
			CanBePublic     bool   `json:"canBePublic"`
			CanBeOrg        bool   `json:"canBeOrg"`
			CanBePrivate    bool   `json:"canBePrivate"`
			CanInvite       bool   `json:"canInvite"`
		} `json:"prefs"`
		LabelNames struct {
			Yellow string `json:"yellow"`
			Red    string `json:"red"`
			Purple string `json:"purple"`
			Orange string `json:"orange"`
			Green  string `json:"green"`
			Blue   string `json:"blue"`
		} `json:"labelNames"`
	} `json:"model"`
}

type WebhooksController struct {
}

func (WebhooksController) Head(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Post is the receiving handler for webhook requests from Trello
//
// Some inspiration drawn from
// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
func (WebhooksController) Post(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		msg := "Content-Type header is not application/json"
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return
	}

	actualSignature := r.Header.Get("X-Trello-Webhook")

	if actualSignature == "" {
		msg := "Request not signed"
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return
	} else {
		mac := hmac.New(sha1.New, []byte(trelloKey))

		// mac.Write(fullRequestBody)
		// mac.Write(callbackUrlAsProvidedDuringCreation)

		expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		if expectedSignature == actualSignature {
			// Do the cache busting here
		} else {
			// do nothing
		}
	}

	w.WriteHeader(http.StatusOK)

	// TODO
	// 1. Check signature

	//Webhook Signatures
	// Trello also signs webhook requests so you can optionally verify that they originated from Trello. Each webhook trigger contains the HTTP header X-Trello-Webhook. The header is a base64 digest of an HMAC-SHA1 hash. The hashed content should be the binary representation of the concatenation of the full request body and the callbackURL exactly as it was provided during webhook creation. The key used to sign this text is your application’s secret. Your application secret can be found at the bottom of https://trello.com/app-key and is also used as the OAuth1.0 secret.

	// 2. Check content type

	// 3. Filter on action type
}
