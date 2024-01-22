package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

/*type Token struct {
Raw       string                 // The raw token.  Populated when you Parse a token
Method    SigningMethod          // The signing method used or to be used
Header    map[string]interface{} // The first segment of the token
Claims    Claims                 // The second segment of the token
Signature string                 // The third segment of the token.  Populated when you Parse a token
Valid     bool                   // Is the token valid?  Populated when you Parse/Verify a token
*/

const minSecretKeySize = 32

//var ErrInvalidToken = errors.New("Invalid token hai Sir Ji!!")

type JwtMaker struct {
	secretKey string
}

func NewJwtMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("Invalid key size :must be at least %d characters", minSecretKeySize)
	}
	return &JwtMaker{secretKey}, nil //this will return error until JwtMaker implement Maker interface functions
}

func (maker *JwtMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

func (maker *JwtMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC) //checking the header is of type
		//var jwt.SigningMethodHS256 *jwt.SigningMethodHMAC
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)

	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	payload, ok := jwtToken.Claims.(*Payload)

	if !ok {
		return nil, ErrInvalidToken
	}
	return payload, nil
}
