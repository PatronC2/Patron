package types

import (
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaim struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

type User struct {
	ID           int    `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Role         string `db:"role"`
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

type UserCreationRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type UpdateUserRequest struct {
	NewPassword *string `json:"newPassword,omitempty"`
	NewRole     *string `json:"newRole,omitempty"`
}

type Note struct {
	NoteID int    `json:"noteid"`
	Note   string `json:"note"`
}

type Tag struct {
	TagID int    `json:"tagid"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Redirector struct {
	RedirectorID string `json:"id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description" binding:"required"`
	ForwardIP    string `json:"forwardip"`
	ForwardPort  string `json:"forwardport"`
	ListenIP     string `json:"listenip" binding:"required"`
	ListenPort   string `json:"listenport" binding:"required"`
	Status       string `json:"status" binding:"required"`
}

type RedirectorTemplateData struct {
	LinkingKey     string
	ApiIP          string
	ApiPort        string
	RedirectorPort string
	ExternalPort   string
	ForwardIP      string
	ForwardPort    string
}

type AgentMetrics struct {
	OnlineCount  string `json:"onlineCount"`
	OfflineCount string `json:"offlineCount"`
}

type TagKeyValues struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}
