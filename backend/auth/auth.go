package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (j *JWTService) GenerateToken(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) ValidateToken(tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return "", "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}
	userID, _ := claims["user_id"].(string)
	role, _ := claims["role"].(string)
	return userID, role, nil
}

type TOTPService struct{}

func NewTOTPService() *TOTPService {
	return &TOTPService{}
}

func (t *TOTPService) GenerateSecret() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

func (t *TOTPService) GenerateCode(secret string) string {
	key, _ := base32.StdEncoding.DecodeString(secret)
	counter := time.Now().Unix() / 30
	msg := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		msg[i] = byte(counter)
		counter >>= 8
	}
	h := hmac.New(sha256.New, key)
	h.Write(msg)
	hash := h.Sum(nil)
	offset := hash[len(hash)-1] & 0x0f
	binary := (int(hash[offset])&0x7f)<<24 | (int(hash[offset+1])&0xff)<<16 | (int(hash[offset+2])&0xff)<<8 | (int(hash[offset+3]) & 0xff)
	return fmt.Sprintf("%06d", binary%1000000)
}

func (t *TOTPService) ValidateCode(secret, code string) bool {
	return t.GenerateCode(secret) == code
}

func (t *TOTPService) GenerateQRURL(email, secret string) string {
	return fmt.Sprintf("otpauth://totp/HumanGuard:%s?secret=%s&issuer=HumanGuard", email, secret)
}