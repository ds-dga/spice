package web

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User stores user's info
type User struct {
	ID        uuid.UUID
	UserName  string
	FirstName string
	LastName  string
	Roles     []string
	Client    string
	IsStaff   bool
	IsActive  bool
}

// Extras to store anything not that important
type Extras struct {
	Source string `json:"source"`
}

// Media for easier access to media
type Media struct {
	ID         uuid.UUID
	Path       string
	ObjectType string
	ObjectUUID uuid.UUID
	ObjectID   int64
	UserID     uuid.UUID
	Extras     Extras
	Name       string
	Size       int64
}

// CreateUser
func (app *WebApp) CreateUser(email, password, firstName, lastName, profileURL string) (*User, error) {
	u := User{
		UserName:  email,
		FirstName: firstName,
		LastName:  lastName,
	}
	// 1. create user -- capture id
	err := app.pdb.QueryRow(`
		INSERT INTO auth_user("username", "email", "first_name", "last_name", "password", "profile_url")
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING id
		`, email, email, firstName, lastName, password, profileURL).Scan(&u.ID)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.New("email already registered")
		}
		return nil, err
	}
	return &u, nil
}

// UpdatePassword
func (app *WebApp) UpdatePassword(user *User, newPassword string) error {
	var returningID uuid.UUID
	err := app.pdb.QueryRow(`
		UPDATE auth_user SET password = $2
		WHERE id = $1
		RETURNING id
		`, user.ID, newPassword).Scan(&returningID)
	if err != nil {
		return err
	}
	return nil
}

func (app *WebApp) SetUserActiveByID(ID uuid.UUID) error {
	var result bool
	err := app.pdb.QueryRow(`
		UPDATE auth_user SET is_active = true
		WHERE id = $1
		RETURNING is_active`, ID).Scan(&result)
	if err != nil {
		return err
	}
	return nil
}

func (app *WebApp) SetUserLastLogin(ID uuid.UUID) error {
	var result bool
	err := app.pdb.QueryRow(`
		UPDATE auth_user SET last_login = NOW()
		WHERE id = $1
		RETURNING is_active`, ID).Scan(&result)
	if err != nil {
		return err
	}
	return nil
}

// FindUserByID returns user
func (app *WebApp) FindUserByID(ID uuid.UUID) (*User, error) {
	u := User{
		Roles: []string{"user"}, // no role at the moment -- "user" as default
	}
	err := app.pdb.QueryRow(`SELECT
		id, username, first_name, last_name, is_staff, is_active
	FROM auth_user
	WHERE id = $1`, ID).Scan(&u.ID, &u.UserName, &u.FirstName, &u.LastName, &u.IsStaff, &u.IsActive)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, errors.New("not registered yet")
		}
		return nil, err
	}
	return &u, nil
}

// FindUser
func (app *WebApp) FindUser(email, password string) (*User, error) {
	u := User{
		Roles: []string{"user"}, // no role at the moment -- "user" as default
	}
	var hashed string
	err := app.pdb.QueryRow(`SELECT
		id, username, first_name, last_name, password, is_staff, is_active
	FROM auth_user
	WHERE email = $1`, email).Scan(&u.ID, &u.UserName, &u.FirstName, &u.LastName, &hashed, &u.IsStaff, &u.IsActive)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, errors.New("not registered yet")
		}
		return nil, err
	}
	if !u.IsActive {
		return nil, errors.New("this account is not active")
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		// Password does not match!
		return nil, errors.New("invalid login credentials")
	}
	return &u, nil
}

// FindUserIfExists to check its existance for further processes such as forget password
func (app *WebApp) FindUserIfExists(email string) (*User, error) {
	u := User{
		Roles: []string{"user"}, // no role at the moment -- "user" as default
	}
	err := app.pdb.QueryRow(`SELECT
		id, username, first_name, last_name, is_staff, is_active
	FROM auth_user
	WHERE email = $1
	LIMIT 1`, email).Scan(&u.ID, &u.UserName, &u.FirstName, &u.LastName, &u.IsStaff, &u.IsActive)
	if err != nil {
		return nil, err
	}
	if u.UserName == "" {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

// VerifyJobDCToken to verify if the token is valid -- for login purposes
func (app *WebApp) VerifyAuthToken(client, jwtToken string) (*User, error) {
	user, tokenKind, err := app.ParseJWTToken(jwtToken)
	if err != nil {
		fmt.Printf("[VerifyToken] %v \n", err)
		return nil, err
	}
	// only tokenKind == "token" is valid here
	if tokenKind != "token" {
		return nil, errors.New("invalid token")
	}
	return user, nil
}

// VerifySpiceID to veried if spice's userID exists
func (app *WebApp) VerifySpiceID(client, userID string) (*User, error) {
	result := User{
		Client: client,
		Roles:  []string{"user"}, // no role at the moment -- "user" as default
	}
	err := app.pdb.QueryRow(`SELECT u.id, u.username, u.first_name, u.last_name
	FROM auth_user u
	WHERE u.id = $1`, userID).Scan(&result.ID, &result.UserName, &result.FirstName, &result.LastName)
	if err != nil {
		// no user found
		return nil, errors.New("no spice user found")
	}
	return &result, nil
}

// VerifySurveyUserID to veried if survey userID exists and has active session
func (app *WebApp) VerifySurveyUserID(client, userID string) (*User, error) {
	result := User{
		Client: client,
		Roles:  []string{"user"}, // no role at the moment -- "user" as default
	}
	var expires int64
	err := app.surveyDB.QueryRow(`SELECT u.id, u.name, u.email, ss.expires
	FROM users u
	INNER JOIN sessions ss ON u.id = ss."userId"
	WHERE u.id = $1
	ORDER BY ss.expires DESC
	LIMIT 1`, userID).Scan(&result.ID, &result.UserName, &result.FirstName, &expires)

	expirationTime := time.Unix(expires/1000, 0)
	if time.Since(expirationTime) > 0 {
		// if expires already pass, then return expired session error
		return nil, errors.New("expired session")
	}
	if err != nil {
		// no user found
		return nil, errors.New("no survey user found")
	}
	return &result, nil
}
