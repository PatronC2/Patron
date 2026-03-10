package types

import (
	"time"

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
	Uuid      string    `json:"uuid"`
	Note      string    `json:"note"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tag struct {
	TagID int64  `json:"tagid"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Redirector struct {
	RedirectorID      string `json:"id" binding:"required"`
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description" binding:"required"`
	ForwardIP         string `json:"forwardip"`
	ForwardPort       string `json:"forwardport"`
	ListenIP          string `json:"listenip" binding:"required"`
	ListenPort        string `json:"listenport" binding:"required"`
	TransportProtocol string `json:"transportprotocol"`
	IPFamily          string `json:"ipfamily"`
	Status            string `json:"status" binding:"required"`
}

type RedirectorTemplateData struct {
	LinkingKey     string
	ApiIP          string
	ApiPort        string
	RedirectorPort string
	ExternalPort   string
	ForwardIP      string
	ForwardPort    string
	ListenIPv4     string
	ListenIPv6     string
}

type RedirectorStatusRequest struct {
	LinkingKey          string   `json:"linking_key" binding:"required"`
	RedirectorProtocols []string `json:"redirectorProtocols"`
	ExternalPort        string   `json:"external_port" binding:"required"`
	ListenIPv4          string   `json:"listen_ipv4" binding:"required"`
	ListenIPv6          string   `json:"listen_ipv6"`
}

type AgentMetrics struct {
	OnlineCount  string `json:"onlineCount"`
	OfflineCount string `json:"offlineCount"`
}

type TagKeyValues struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type Listener struct {
	ListenerID        int
	Name              string
	Description       string
	ListenIP          string
	ListenPort        int
	TransportProtocol string
}

type File struct {
	FileID    int       `json:"file_id" binding:"required"`
	AgentId   string    `json:"uuid"`
	Type      string    `json:"type"`
	Path      string    `json:"path"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
