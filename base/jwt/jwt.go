package jwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	signKey = []byte("lijun")
)

var (
	TokenClaimsInvalid error = errors.New("Token claims is invalid")
)

type UserClaims struct {
	Name string `json:"userName"`
	IP   string `json:"ip"`
	jwt.StandardClaims
}

func CreateToken(name string, ip string, duration time.Duration) (string, error) {
	return CreateTokenWithClaims(&UserClaims{
		Name: name,
		IP:   ip,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(duration).Unix(),
		},
	})
}

func CreateTokenWithClaims(claims *UserClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signKey)
}

func VerifyToken(token string) (*UserClaims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := jwtToken.Claims.(*UserClaims); ok && jwtToken.Valid {
		return claims, nil
	} else {
		return nil, TokenClaimsInvalid
	}
}

func ResetToken(token string, duration time.Duration) (string, error) {
	claims, err := VerifyToken(token)
	if err != nil {
		return "", err
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(duration).Unix()
	return CreateTokenWithClaims(claims)
}
