package auth

import (
	"kentech-project/pkg/logger"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey []byte
	logger    *logger.Logger
}

func NewJWTService(secretKey string, log *logger.Logger) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		logger:    log,
	}
}

func (j *JWTService) GenerateToken(userID uuid.UUID, username string) (string, error) {
	j.logger.Debugf("GenerateToken called: user_id=%s, username=%s", userID.String(), username)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(j.secretKey)
	if err != nil {
		j.logger.Error("Failed to sign JWT token: " + err.Error())
		return "", err
	}
	j.logger.Infof("JWT token generated for user_id=%s", userID.String())
	return signedToken, nil
}

func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	j.logger.Debug("ValidateToken called")
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})

	if err != nil {
		j.logger.Warn("JWT token validation failed: " + err.Error())
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		j.logger.Info("JWT token validated successfully")
		return claims, nil
	}

	j.logger.Error("JWT token signature invalid")
	return nil, jwt.ErrSignatureInvalid
}
