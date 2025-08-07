package main

import (
	"github.com/fusioncatltd/fusioncat/api/protected_endpoints"
	"github.com/fusioncatltd/fusioncat/api/public_endpoints"
	"github.com/fusioncatltd/fusioncat/common"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"runtime"
)

func main() {
	// Preparing logging
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Load environment variables from .env file
	_, filename, _, ok := runtime.Caller(0)
	pathToEnvFile := path.Dir(filename) + "/.env"
	if !ok {
		panic("Can't get path to current source file. This is needed to load .env file.")
	}
	err := godotenv.Load(pathToEnvFile)

	if err != nil {
		log.Warning(".env file not found at %, using environment variables.", pathToEnvFile)
	} else {
		log.Infof(".env file loaded from %s", pathToEnvFile)
	}

	// Start the application
	r := gin.Default()

	// Set up CORS
	config := cors.DefaultConfig()
	config.AllowCredentials = true
	config.AddExposeHeaders("Authorization", "Set-Cookie", "Content-Type")
	config.AddAllowHeaders("Authorization")
	config.AllowOriginWithContextFunc = func(c *gin.Context, origin string) bool {
		return true
	}
	r.Use(cors.New(config))

	// Set up API routes
	V1PublicRoutesGroup := r.Group("/v1/public")
	public_endpoints.UsersPublicRouterV1(V1PublicRoutesGroup)

	V1ProtectedRoutesGroup := r.Group("/v1/protected")
	V1ProtectedRoutesGroup.Use(common.JwtOrApiKeyAuthMiddleware())
	protected_endpoints.AuthenticationProtectedRoutesV1(V1ProtectedRoutesGroup)

	// Launching server
	serverAddressPort := os.Getenv("SERVER_ADDRESS_AND_PORT")
	serverMode := os.Getenv("SERVER_MODE")

	gin.SetMode(serverMode)
	launchErr := r.Run(serverAddressPort)

	if launchErr != nil {
		panic("Failed to start the fusioncat server: " + launchErr.Error())
	}
}
