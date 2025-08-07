package common

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func GenerateGwtToken(userID uuid.UUID) (string, error) {
	token_lifespan, err := strconv.Atoi(os.Getenv("JWT_TOKEN_LIFESPAN_IN_HOURS"))

	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(token_lifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ExtractAuthTokenOrKeyFromHeader(c *gin.Context) string {
	var extractedToken string

	token := c.Query("token")
	if token != "" {
		extractedToken = token
		return extractedToken
	}
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		extractedToken = strings.Split(bearerToken, " ")[1]
		return extractedToken
	}
	return ""
}

func IsCallMadeViaUsersAPIKey(c *gin.Context, token string) bool {
	c.Set("IsExternalAPICall", false)
	return false
}

func ExtractJWTTokenFromCookie(c *gin.Context) string {
	cookieName := os.Getenv("COOKIE_NAME")
	cookie, err := c.Request.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func IsJWTTokenValid(c *gin.Context) error {
	tokenString := ExtractAuthTokenOrKeyFromHeader(c)
	if tokenString == "" {
		tokenString = ExtractJWTTokenFromCookie(c)
	}
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return err
	}
	return nil
}

func ExtractUserIdFromJWTToken(c *gin.Context) (uuid.UUID, error) {
	tokenString := ExtractAuthTokenOrKeyFromHeader(c)
	if tokenString == "" {
		tokenString = ExtractJWTTokenFromCookie(c)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, err
	}
	userId, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
}

// JwtOrApiKeyAuthMiddleware middleware checks if the request is authenticated either via JWT token or API key.
func JwtOrApiKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := IsJWTTokenValid(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		userId, err := ExtractUserIdFromJWTToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, nil)
			c.Abort()
		}
		c.Set("UserID", userId)

		c.Next()
	}
}
