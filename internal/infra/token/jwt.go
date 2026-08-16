package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrMethod       = errors.New("unexpected signing method")
	ErrInvalidToken = errors.New("invalid access token")
	ErrNotFound     = errors.New("missing user id")
)

type Config struct {
	Secret     string        `env:"SECRET"`
	Issuer     string        `env:"ISSUER"`
	RefreshTTL time.Duration `env:"RT_TTL"`
	AccessTTL  time.Duration `env:"AT_TTL"`
}

type JWTService struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewJWTService(conf Config) *JWTService {
	return &JWTService{
		secret: []byte(conf.Secret),
		issuer: conf.Issuer,
		ttl:    conf.AccessTTL,
	}
}

type accessTokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *JWTService) Generate(userID uuid.UUID) (string, error) {
	now := time.Now()

	claims := accessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(s.secret)
}

func (s *JWTService) Validate(tokenString string) (uuid.UUID, error) {
	var claims accessTokenClaims

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrMethod
			}

			return s.secret, nil
		},
		jwt.WithIssuer(s.issuer),
	)

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	if claims.UserID == uuid.Nil {
		return uuid.Nil, ErrNotFound
	}

	return claims.UserID, nil
}
