package main

import (
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/api/protected_endpoints"
	"github.com/fusioncatltd/fusioncat/api/public_endpoints"
	"github.com/fusioncatltd/fusioncat/common"
	_ "github.com/fusioncatltd/fusioncat/docs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	ff "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
)

// @title FusionCat API
// @version 1.0
// @description API Server for FusionCat application

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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
	public_endpoints.AuthenticationPublicRoutesV1(V1PublicRoutesGroup)

	V1ProtectedRoutesGroup := r.Group("/v1/protected")
	V1ProtectedRoutesGroup.Use(common.JwtOrApiKeyAuthMiddleware())
	protected_endpoints.AuthenticationProtectedRoutesV1(V1ProtectedRoutesGroup)
	protected_endpoints.MeProtectedRoutesV1(V1ProtectedRoutesGroup)
	protected_endpoints.ProjectsProtectedRoutesV1(V1ProtectedRoutesGroup)
	protected_endpoints.SchemasProtectedRoutesV1(V1ProtectedRoutesGroup)
	protected_endpoints.MessagesProtectedRoutesV1(V1ProtectedRoutesGroup)

	// Set up Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(ff.Handler))

	// Assigning custom validators
	// Modifying the behavior of the default validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {

		// Preserve the original names of the JSON fields in order to re-user them
		// later in validation error responses
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			// Get the string before the first comma in the `json` tag
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		// Add new custom validators here
		validators := map[string]validator.Func{
			"valid_json_schema":                      input_contracts.ValidJSONSchemaValidator,
			"alphanum_with_underscore":               input_contracts.ValidateAlphanumWithUnderscore,
			"alphanum_with_underscore_and_dots":      input_contracts.ValidateAlphanumWithUnderscoreAndDots,
			"valid_existing_schema_id_and_version":   input_contracts.ValidExistingSchemaIDAndVersionValidator,
		}

		for name, fn := range validators {
			if err := v.RegisterValidation(name, fn); err != nil {
				panic(err)
			}
		}
	}

	// Launching server
	serverAddressPort := os.Getenv("SERVER_ADDRESS_AND_PORT")
	serverMode := os.Getenv("SERVER_MODE")

	gin.SetMode(serverMode)
	launchErr := r.Run(serverAddressPort)

	if launchErr != nil {
		panic("Failed to start the fusioncat server: " + launchErr.Error())
	}
}
