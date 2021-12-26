package web

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type socialConnectBody struct {
	Provider   string `json:"provider"`
	SocialID   string `json:"uid"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	ProfileURL string `json:"profileURL"`
}

func (app *WebApp) VerifySocialAuthToken(client, socialApp, socialID string) (*User, error) {
	result := User{
		Client: client,
	}
	err := app.pdb.QueryRow(`SELECT u.id, u.username, u.first_name, u.last_name
	FROM social_account acc
	LEFT JOIN auth_user u ON acc.user_id = u.id
	WHERE acc.provider = $1 AND acc.uid = $2`, socialApp, socialID).Scan(&result.ID, &result.UserName, &result.FirstName, &result.LastName)
	if err != nil {
		// no user found
		return nil, errors.New("no social user found")
	}
	return &result, nil
}

// SocialConnect meant to be "check or create" social account & auth_user for other stuffs.
func (app *WebApp) SocialConnect(w http.ResponseWriter, req *http.Request) {
	var body socialConnectBody
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		resp := msgReponse{
			Result:  "failed",
			Message: "JSON decoded failed",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	if body.Provider == "" {
		resp := msgReponse{
			Result:  "failed",
			Message: "missing provider",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	if body.SocialID == "" {
		resp := msgReponse{
			Result:  "failed",
			Message: "missing uid",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	if body.Username == "" && body.Email == "" {
		resp := msgReponse{
			Result:  "failed",
			Message: "missing username/email",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	if body.Email == "" {
		body.Email = body.Username
	}

	_, err = app.VerifySocialAuthToken("", body.Provider, body.SocialID)
	if err != nil {
		// create one
		user, err := app.CreateUser(body.Username, "", "-", "-", body.ProfileURL)
		if err != nil {
			resp := msgReponse{
				Result:  "failed",
				Message: "create user error",
			}
			authJSONResponse(w, http.StatusBadRequest, resp)
			return
		}
		// create social_account to connect with auth_user
		err = app.CreateSocialAccount(user, body.Provider, body.SocialID)
		if err != nil {
			resp := msgReponse{
				Result:  "failed",
				Message: "create social account error",
			}
			authJSONResponse(w, http.StatusBadRequest, resp)
			return
		}
	}
	resp := msgReponse{
		Result:  "success",
		Message: "ok",
	}
	authJSONResponse(w, http.StatusOK, resp)
}

// CreateSocialAccount create social_account record to link with auth_user
func (app *WebApp) CreateSocialAccount(user *User, provider, socialID string) error {
	var sid int64
	// 1. create user -- capture id
	err := app.pdb.QueryRow(`
		INSERT INTO social_account("user_id", "provider", "uid", "date_joined")
		VALUES($1, $2, $3, NOW())

		RETURNING id
		`, user.ID, provider, socialID).Scan(&sid)
	if err != nil {
		log.Printf("[CreateSocAccount] %v", err)
		return err
	}
	return nil
}
