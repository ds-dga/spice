package web

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type JwtClaims struct {
	jwt.StandardClaims
	Kind    string    `json:"kd"`
	UserID  uuid.UUID `json:"uid"`
	IsStaff bool      `json:"staff"`
	Email   string    `json:"email"`
}

// msgReponse contains necessary stuffs for the next stuffs
type msgReponse struct {
	Message string `json:"message"`
	Result  string `json:"result"`
}

func authJSONResponse(w http.ResponseWriter, httpCode int, resp msgReponse) {
	w.WriteHeader(httpCode)
	w.Header().Set("Content-Type", "application/json")
	jMessage, _ := json.Marshal(resp)
	w.Write(jMessage)
}

// HeaderAuthenticate verifies if auth is ok
func (app *WebApp) HeaderAuthenticate(req *http.Request) (*User, error) {
	/* Required headers:
	* winter-falcon-client: client-as-string
	* winter-falcon-social-app: facebook, google, email
	* winter-falcon-uid: userid-string
	 */
	headers := req.Header.Clone()
	client := headers.Get("winter-falcon-client")
	socialApp := headers.Get("winter-falcon-social-app")
	uid := headers.Get("winter-falcon-uid")
	jwtToken := headers.Get("winter-falcon-jwt")

	if uid == "anonymous" {
		return &User{
			Roles: []string{"anonymous"},
		}, nil
	}

	if socialApp == "email" {
		if len(jwtToken) != 0 {
			return app.VerifyAuthToken(client, jwtToken)
		}
		return nil, errors.New("bad request")
	}

	if socialApp == "spice" {
		if len(jwtToken) != 0 {
			return app.VerifySpiceID(client, uid)
		}
		return nil, errors.New("bad request")
	}

	if len(socialApp) == 0 && len(uid) == 0 {
		return nil, errors.New("bad request")
	}

	return app.VerifySocialAuthToken(client, socialApp, uid)
}

type LoginBody struct {
	Email    string
	Password string
}

func (app *WebApp) Login(w http.ResponseWriter, req *http.Request) {
	var body LoginBody
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
	user, err := app.FindUser(body.Email, body.Password)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: err.Error(),
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	app.JWTResponse(user, w)
}

// JWTResponse return jwt token for GQL request
func (app *WebApp) JWTResponse(user *User, w http.ResponseWriter) {
	token, err := app.GenerateJWTToken(user, "token")
	if err != nil {
		resp := msgReponse{
			Result:  "failed: JWT generator",
			Message: err.Error(),
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}

	err = app.SetUserLastLogin(user.ID)
	if err != nil {
		resp := msgReponse{
			Result:  "failed: User updater",
			Message: err.Error(),
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}

	mt := map[string]string{}
	mt["result"] = "success"
	mt["jwt"] = token
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	msg, _ := json.Marshal(mt)
	w.Write(msg)
}

type SignUpBody struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func (app *WebApp) SignUp(w http.ResponseWriter, req *http.Request) {
	dataIn, err := io.ReadAll(req.Body)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: "JSON decoded failed",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	var body SignUpBody
	err = json.Unmarshal(dataIn, &body)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: "JSON decoded failed",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	fmt.Println("signup body: ", body)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: err.Error(),
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
	}
	user, err := app.CreateUser(body.Email, string(hashedPassword), body.FirstName, body.LastName)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: err.Error(),
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}

	resp := msgReponse{
		Result:  "success",
		Message: user.ID.String(),
	}
	authJSONResponse(w, http.StatusOK, resp)

	// send mail
	unixTime := time.Now().Unix()
	message := fmt.Sprintf("Please confirm your email by click at the following link"+
		"\n\n"+
		"https://ds.10z.dev/email-confirmation?nbt=%d&em=%s&key=%s"+
		"\n\n"+
		"Thank you", unixTime, user.UserName, user.SeedGenerator(unixTime))
	SendMail(user.UserName, "Registration confirmation", message)
}

func (u User) SeedGenerator(tm int64) string {
	seed := fmt.Sprintf("%d-%s-%s", tm, u.ID, u.UserName)
	hash := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%x", hash)
}

type ForgetPasswordBody struct {
	Email string
}

func (app *WebApp) ForgetPassword(w http.ResponseWriter, req *http.Request) {
	var body ForgetPasswordBody
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: "JSON decoded failed",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	user, err := app.FindUserIfExists(body.Email)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: fmt.Sprintf("%s does not exists in the system.", body.Email),
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	magicToken, err := app.GenerateJWTToken(user, "magic-link")
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: err.Error(),
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	// Sending mail
	unixTime := time.Now().Unix()
	token, err := app.GenerateJWTToken(user, "magic")
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: err.Error(),
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	message := fmt.Sprintf("Please click at the magic link below to signin to the system automatically."+
		"\n\n"+
		"https://ds.10z.dev/magic-link?nbt=%d&key=%s"+
		"\n\n"+
		"Thank you", unixTime, token)
	SendMail(user.UserName, "Forget password?", message)

	mt := map[string]string{}
	mt["result"] = "success"
	mt["magic"] = magicToken
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	msg, _ := json.Marshal(mt)
	w.Write(msg)
}

type MagicBody struct {
	Nbt   string
	Email string
	Key   string
}

// Confirmation sets active to the account
func (app *WebApp) Confirmation(w http.ResponseWriter, req *http.Request) {
	var body MagicBody
	// magicBody comes in GET query
	q := req.URL.Query()
	body.Nbt = q.Get("nbt")
	body.Email = q.Get("em")
	body.Key = q.Get("key")

	NotBefore, err := strconv.Atoi(body.Nbt)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: "invalid time",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	unixTime := time.Now().Unix()
	if int(unixTime) < NotBefore {
		resp := msgReponse{
			Result:  "failed",
			Message: "invalid time",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	user, err := app.FindUserIfExists(body.Email)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: "bad request",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	if user.IsActive {
		resp := msgReponse{
			Result:  "failed",
			Message: "already active",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	seed := user.SeedGenerator(int64(NotBefore))
	if seed != body.Key {
		resp := msgReponse{
			Result:  "failed",
			Message: "bad request",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	err = app.SetUserActiveByID(user.ID)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: err.Error(),
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}

	resp := msgReponse{
		Result:  "success",
		Message: "active now",
	}
	authJSONResponse(w, http.StatusOK, resp)
}

// MagicLink handles special very short-life OTP like
func (app *WebApp) MagicLink(w http.ResponseWriter, req *http.Request) {
	var body MagicBody
	// magicBody comes in GET query
	q := req.URL.Query()
	body.Nbt = q.Get("nbt")
	body.Email = q.Get("em")
	body.Key = q.Get("key")

	NotBefore, err := strconv.Atoi(body.Nbt)
	if err != nil {
		resp := msgReponse{
			Result:  "failed",
			Message: "nah, you've got no magic",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	unixTime := time.Now().Unix()
	if int(unixTime) < NotBefore {
		resp := msgReponse{
			Result:  "failed",
			Message: "nah, you've got no magic",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	user, kind, err := app.ParseJWTToken(body.Key)
	if err != nil || kind != "magic" {
		resp := msgReponse{
			Result:  "failed",
			Message: "nah, you've got no magic",
		}
		authJSONResponse(w, http.StatusBadRequest, resp)
		return
	}
	app.JWTResponse(user, w)
}

// GenerateJWTToken to gen JWT token
func (app *WebApp) GenerateJWTToken(user *User, kind string) (string, error) {
	claims := JwtClaims{
		Kind:    kind,
		UserID:  user.ID,
		IsStaff: user.IsStaff,
		Email:   user.UserName,
	}
	claims.NotBefore = time.Date(2021, 8, 1, 12, 0, 0, 0, time.UTC).Unix()
	claims.IssuedAt = time.Now().Unix()
	claims.Issuer = "@sipp11"
	claims.ExpiresAt = time.Now().Add(time.Hour * 24 * 30).Unix() // 30 days
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(app.secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseJWTToken parse the token and return token kind ("token" or "magic")
func (app *WebApp) ParseJWTToken(jwtToken string) (*User, string, error) {

	token, err := jwt.ParseWithClaims(jwtToken, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return app.secret, nil
	})

	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		// fmt.Println(claims["gmk"], claims["keyboard"], claims["nbf"])
		user, errU := app.FindUserByID(claims.UserID)
		if errU != nil {
			// fmt.Printf("[ParseJWTToken] 01 %v \n", errU.Error())
			return nil, "", errors.New("user not found")
		}
		return user, claims.Kind, nil
	}
	// fmt.Printf("[ParseJWTToken] 11 %v \n", err.Error())
	return nil, "", err
}
