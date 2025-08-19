package large_chunks_of_logic

import (
	"bytes"
	"fmt"
	"github.com/fusioncatltd/fusioncat/common"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/google/uuid"
)

// GenerateAppCode generates the complete Go code for an application using templates
func GenerateAppCode(app *logic.AppObject, usage *logic.AppUsageMatrixResponse) (string, error) {
	// Get templates folder from environment
	templatesFolder := os.Getenv("PATH_TO_STUBS_TEMPLATES_FOLDER")
	if templatesFolder == "" {
		return "", fmt.Errorf("PATH_TO_STUBS_TEMPLATES_FOLDER environment variable not set")
	}

	// Debug log
	fmt.Printf("Using templates folder: %s\n", templatesFolder)

	// Define template data structures
	type SchemaTemplateData struct {
		ID          string
		Name        string
		Version     int
		Description string
		StructName  string
		Code        string
	}

	type MessageTemplateData struct {
		ID               string
		Name             string
		Description      string
		SchemaID         string
		SchemaVersion    int
		StructName       string
		SchemaStructName string
	}

	type ResourceTemplateData struct {
		ID           string
		Name         string
		Description  string
		Type         string
		Mode         string
		ServerName   string
		ResourcePath string
		StructName   string
	}

	type ServerTemplateData struct {
		ID          string
		Name        string
		Description string
		Protocol    string
		StructName  string
		Resources   []ResourceTemplateData
	}

	type AppTemplateData struct {
		ID          string
		Name        string
		Description string
		StructName  string
		Sends       []struct {
			ResourceStructName string
			Messages           []MessageTemplateData
		}
		Receives []struct {
			ResourceStructName string
			Messages           []MessageTemplateData
		}
	}

	// Wrapper structures for templates
	type SchemaTemplateDataWrapper struct {
		Schemas []SchemaTemplateData
	}

	type MessageTemplateDataWrapper struct {
		Messages []MessageTemplateData
	}

	type ResourceTemplateDataWrapper struct {
		Resources []ResourceTemplateData
	}

	type ServerTemplateDataWrapper struct {
		Servers []ServerTemplateData
	}

	type AppTemplateDataWrapper struct {
		Apps []AppTemplateData
	}

	// Initialize managers
	schemasManager := logic.SchemaObjectsManager{}
	messagesManager := logic.MessagesObjectsManager{}
	resourcesManager := logic.ResourcesObjectsManager{}
	serversManager := logic.ServersObjectsManager{}

	// Collect unique entities
	schemas := make(map[string]*logic.SchemaObject)
	messages := make(map[string]*logic.MessageObject)
	resources := make(map[string]*logic.ResourceObject)
	servers := make(map[string]*logic.ServerObject)

	// Process all sends and receives to collect entities
	allUsages := append(usage.Sends, usage.Receives...)
	for _, u := range allUsages {
		// Get message
		if u.Message != nil && u.Message.ID != "" {
			if _, exists := messages[u.Message.ID]; !exists {
				msg, err := messagesManager.GetByID(uuid.MustParse(u.Message.ID))
				if err == nil {
					messages[u.Message.ID] = msg

					// Get schema for this message
					schema, err := schemasManager.GetByID(msg.GetSchemaID())
					if err == nil {
						schemas[schema.GetID().String()] = schema
					}
				}
			}
		}

		// Get resource
		if u.Resource != nil && u.Resource.ID != "" {
			if _, exists := resources[u.Resource.ID]; !exists {
				res, err := resourcesManager.GetByID(uuid.MustParse(u.Resource.ID))
				if err == nil {
					resources[u.Resource.ID] = res
				}
			}
		}

		// Get server
		if u.Server != nil && u.Server.ID != "" {
			if _, exists := servers[u.Server.ID]; !exists {
				srv, err := serversManager.GetByID(uuid.MustParse(u.Server.ID))
				if err == nil {
					servers[u.Server.ID] = srv
				}
			}
		}
	}

	// Initialize template data collections
	schemaImplData := SchemaTemplateDataWrapper{
		Schemas: []SchemaTemplateData{},
	}
	messageImplData := MessageTemplateDataWrapper{
		Messages: []MessageTemplateData{},
	}
	resourceImplData := ResourceTemplateDataWrapper{
		Resources: []ResourceTemplateData{},
	}
	serverImplData := ServerTemplateDataWrapper{
		Servers: []ServerTemplateData{},
	}
	appImplData := AppTemplateDataWrapper{
		Apps: []AppTemplateData{},
	}

	// Track struct names for references
	schemaStructNames := make(map[string]string)
	messageStructNames := make(map[string]string)
	resourceStructNames := make(map[string]string)
	serverStructNames := make(map[string]string)

	// Generate schemas first
	for schemaID, schema := range schemas {
		// Get schema version for code generation
		schemaVersion, err := schemasManager.GetSpecificVersionOfSchema(schema.GetID(), schema.GetCurrentVersion())
		if err != nil {
			continue
		}

		// Generate schema code
		schemaCode, structName, err := schemaVersion.GenerateCode("go", schema.GetType(), schema.GetName())
		if err != nil {
			continue
		}

		// Use the struct name with a suffix to ensure uniqueness
		formattedStructName := structName + "FusioncatGeneratedSchema"
		schemaStructNames[schemaID] = formattedStructName

		// Replace the struct name in the generated code
		schemaCodeWithCorrectName := strings.Replace(schemaCode, "type "+structName+" struct", "type "+formattedStructName+" struct", 1)

		// Add schema implementation data
		schemaImplData.Schemas = append(schemaImplData.Schemas, SchemaTemplateData{
			ID:          schemaID,
			Name:        schema.GetName(),
			Version:     schema.GetCurrentVersion(),
			Description: strings.ReplaceAll(schema.Serialize().Description, "\"", "\\\""),
			StructName:  formattedStructName,
			Code:        schemaCodeWithCorrectName,
		})
	}

	// Generate messages
	processedMessages := make(map[string]bool)
	for messageID, message := range messages {
		if processedMessages[messageID] {
			continue
		}
		processedMessages[messageID] = true

		schemaStructName, exists := schemaStructNames[message.GetSchemaID().String()]
		if !exists {
			continue
		}

		// Convert message name to proper format
		msgStructName := common.ToCamelCase(message.Serialize().Name) + "FusioncatGeneratedMessage"
		messageStructNames[messageID] = msgStructName

		messageImplData.Messages = append(messageImplData.Messages, MessageTemplateData{
			ID:               messageID,
			Name:             message.Serialize().Name,
			Description:      strings.ReplaceAll(message.Serialize().Description, "\"", "\\\""),
			SchemaID:         message.GetSchemaID().String(),
			SchemaVersion:    message.GetSchemaVersion(),
			StructName:       msgStructName,
			SchemaStructName: schemaStructName,
		})
	}

	// Generate resources and servers
	processedResources := make(map[string]bool)
	processedServers := make(map[string]bool)

	for resourceID, resource := range resources {
		if processedResources[resourceID] {
			continue
		}
		processedResources[resourceID] = true

		// Sanitize resource name by replacing dots with underscores
		sanitizedResourceName := strings.ReplaceAll(resource.Serialize().Name, ".", "_")
		resourceStructName := common.ToCamelCase(sanitizedResourceName) + "FusioncatGeneratedResource"
		resourceStructNames[resourceID] = resourceStructName

		// Parse resource path from resource name or use a default
		resourcePath := "topic/" + strings.ToLower(resource.Serialize().Name)
		if resource.Serialize().ResourceType == "endpoint" {
			resourcePath = "/" + strings.ToLower(resource.Serialize().Name)
		}

		resourceImplData.Resources = append(resourceImplData.Resources, ResourceTemplateData{
			ID:           resourceID,
			Name:         resource.Serialize().Name,
			Description:  strings.ReplaceAll(resource.Serialize().Description, "\"", "\\\""),
			Type:         resource.Serialize().ResourceType,
			Mode:         resource.Serialize().Mode,
			ServerName:   "", // Will be filled when processing servers
			ResourcePath: resourcePath,
			StructName:   resourceStructName,
		})
	}

	// Generate servers
	for serverID, server := range servers {
		if processedServers[serverID] {
			continue
		}
		processedServers[serverID] = true

		serverStructName := common.ToCamelCase(server.Serialize().Name) + "FusioncatGeneratedServer"
		serverStructNames[serverID] = serverStructName

		// Get all resources for this server
		var serverResources []ResourceTemplateData
		for _, resource := range resources {
			if resource.GetServerID().String() == serverID {
				resStructName := resourceStructNames[resource.GetID().String()]
				serverResources = append(serverResources, ResourceTemplateData{
					ID:          resource.GetID().String(),
					Name:        resource.Serialize().Name,
					Description: strings.ReplaceAll(resource.Serialize().Description, "\"", "\\\""),
					Type:        resource.Serialize().ResourceType,
					Mode:        resource.Serialize().Mode,
					StructName:  resStructName,
				})
			}
		}

		serverImplData.Servers = append(serverImplData.Servers, ServerTemplateData{
			ID:          serverID,
			Name:        server.Serialize().Name,
			Description: strings.ReplaceAll(server.Serialize().Description, "\"", "\\\""),
			Protocol:    server.Serialize().Protocol,
			StructName:  serverStructName,
			Resources:   serverResources,
		})
	}

	// Generate app
	appStructName := common.ToCamelCase(app.Serialize().Name) + "FusioncatGeneratedApp"

	// Process sends
	sendsByResource := make(map[string][]MessageTemplateData)
	for _, send := range usage.Sends {
		if send.Resource != nil && send.Message != nil {
			resID := send.Resource.ID
			msgStructName := messageStructNames[send.Message.ID]
			if msgStructName != "" {
				sendsByResource[resID] = append(sendsByResource[resID], MessageTemplateData{
					StructName: msgStructName,
				})
			}
		}
	}

	var sends []struct {
		ResourceStructName string
		Messages           []MessageTemplateData
	}
	for resID, msgs := range sendsByResource {
		if resStructName, exists := resourceStructNames[resID]; exists {
			sends = append(sends, struct {
				ResourceStructName string
				Messages           []MessageTemplateData
			}{
				ResourceStructName: resStructName,
				Messages:           msgs,
			})
		}
	}

	// Process receives
	receivesByResource := make(map[string][]MessageTemplateData)
	for _, recv := range usage.Receives {
		if recv.Resource != nil && recv.Message != nil {
			resID := recv.Resource.ID
			msgStructName := messageStructNames[recv.Message.ID]
			if msgStructName != "" {
				receivesByResource[resID] = append(receivesByResource[resID], MessageTemplateData{
					StructName: msgStructName,
				})
			}
		}
	}

	var receives []struct {
		ResourceStructName string
		Messages           []MessageTemplateData
	}
	for resID, msgs := range receivesByResource {
		if resStructName, exists := resourceStructNames[resID]; exists {
			receives = append(receives, struct {
				ResourceStructName string
				Messages           []MessageTemplateData
			}{
				ResourceStructName: resStructName,
				Messages:           msgs,
			})
		}
	}

	appImplData.Apps = append(appImplData.Apps, AppTemplateData{
		ID:          app.GetID().String(),
		Name:        app.Serialize().Name,
		Description: strings.ReplaceAll(app.Serialize().Description, "\"", "\\\""),
		StructName:  appStructName,
		Sends:       sends,
		Receives:    receives,
	})

	// Load and execute templates
	mainTmpl, err := template.ParseFiles(filepath.Join(templatesFolder, "go", "template.tmpl"))
	if err != nil {
		return "", fmt.Errorf("failed to load main template: %v", err)
	}

	schemaImplTmpl, err := template.ParseFiles(filepath.Join(templatesFolder, "go", "schema_implementation.tmpl"))
	if err != nil {
		return "", fmt.Errorf("failed to load schema implementation template: %v", err)
	}

	messageImplTmpl, err := template.ParseFiles(filepath.Join(templatesFolder, "go", "message_implementation.tmpl"))
	if err != nil {
		return "", fmt.Errorf("failed to load message implementation template: %v", err)
	}

	resourceImplTmpl, err := template.ParseFiles(filepath.Join(templatesFolder, "go", "resource_implementation.tmpl"))
	if err != nil {
		return "", fmt.Errorf("failed to load resource implementation template: %v", err)
	}

	serverImplTmpl, err := template.ParseFiles(filepath.Join(templatesFolder, "go", "server_implementation.tmpl"))
	if err != nil {
		return "", fmt.Errorf("failed to load server implementation template: %v", err)
	}

	appImplTmpl, err := template.ParseFiles(filepath.Join(templatesFolder, "go", "app_implementation.tmpl"))
	if err != nil {
		return "", fmt.Errorf("failed to load app implementation template: %v", err)
	}

	// Execute templates
	var schemaImplBuffer bytes.Buffer
	if err := schemaImplTmpl.Execute(&schemaImplBuffer, schemaImplData); err != nil {
		return "", fmt.Errorf("failed to generate schema implementations: %v", err)
	}

	var messageImplBuffer bytes.Buffer
	if err := messageImplTmpl.Execute(&messageImplBuffer, messageImplData); err != nil {
		return "", fmt.Errorf("failed to generate message implementations: %v", err)
	}

	var resourceImplBuffer bytes.Buffer
	if err := resourceImplTmpl.Execute(&resourceImplBuffer, resourceImplData); err != nil {
		return "", fmt.Errorf("failed to generate resource implementations: %v", err)
	}

	var serverImplBuffer bytes.Buffer
	if err := serverImplTmpl.Execute(&serverImplBuffer, serverImplData); err != nil {
		return "", fmt.Errorf("failed to generate server implementations: %v", err)
	}

	var appImplBuffer bytes.Buffer
	if err := appImplTmpl.Execute(&appImplBuffer, appImplData); err != nil {
		return "", fmt.Errorf("failed to generate app implementations: %v", err)
	}

	// Create the final template data
	templateData := struct {
		SchemaImplementations   string
		MessageImplementations  string
		ResourceImplementations string
		ServerImplementations   string
		AppImplementations      string
	}{
		SchemaImplementations:   schemaImplBuffer.String(),
		MessageImplementations:  messageImplBuffer.String(),
		ResourceImplementations: resourceImplBuffer.String(),
		ServerImplementations:   serverImplBuffer.String(),
		AppImplementations:      appImplBuffer.String(),
	}

	// Execute main template
	var mainBuffer bytes.Buffer
	if err := mainTmpl.Execute(&mainBuffer, templateData); err != nil {
		return "", fmt.Errorf("failed to generate main code: %v", err)
	}

	return mainBuffer.String(), nil
}
