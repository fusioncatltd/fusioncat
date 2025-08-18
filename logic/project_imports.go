package logic

import (
	"fmt"
	"strings"

	asyncuri "github.com/fusioncatltd/lib-go-asyncresourceuri"
	"github.com/google/uuid"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// YAML structures for project import
type ProjectImportYAML struct {
	Version  int            `yaml:"version"`
	Servers  []ServerImport `yaml:"servers"`
	Schemas  []SchemaImport `yaml:"schemas"`
	Messages []MessageImport `yaml:"messages"`
	Apps     []AppImport     `yaml:"apps"`
}

type ServerImport struct {
	Name        string           `yaml:"name"`
	Type        string           `yaml:"type"`
	Description string           `yaml:"description"`
	Resources   []ResourceImport `yaml:"resources"`
	Binds       []BindImport     `yaml:"binds,omitempty"`
}

type ResourceImport struct {
	Name        string `yaml:"name"`
	Mode        string `yaml:"mode"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

type BindImport struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type SchemaImport struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Version     int    `yaml:"version"`
	Description string `yaml:"description"`
	Schema      string `yaml:"schema"`
}

type MessageImport struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Schema      SchemaReference `yaml:"schema"`
}

type SchemaReference struct {
	Name string `yaml:"name"`
}

type AppImport struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Sends       []SendImport    `yaml:"sends,omitempty"`
	Receives    []ReceiveImport `yaml:"receives,omitempty"`
}

type SendImport struct {
	Message  string `yaml:"message"`
	Resource string `yaml:"resource"`
}

type ReceiveImport struct {
	Message  string `yaml:"message"`
	Resource string `yaml:"resource"`
}

// ValidateProjectImportYAML validates the YAML structure and references
func ValidateProjectImportYAML(yamlStr string, projectID uuid.UUID) []string {
	var projectImport ProjectImportYAML
	var errors []string

	err := yaml.Unmarshal([]byte(yamlStr), &projectImport)
	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to unmarshal YAML: %v", err))
		return errors
	}

	if projectImport.Version != 1 {
		errors = append(errors, fmt.Sprintf("invalid version: %d, expected 1", projectImport.Version))
	}

	// Create maps to track resources and servers for validation
	serverResources := make(map[string]map[string]bool) // server -> resource name -> exists
	serverTypes := make(map[string]string)              // server name -> type
	resourceModes := make(map[string]string)            // server.resource -> mode
	resourceTypes := make(map[string]string)            // server.resource -> type

	// Validate servers and their resources
	for _, server := range projectImport.Servers {
		if server.Name == "" {
			errors = append(errors, "server name is required")
			continue
		}
		if server.Type == "" {
			errors = append(errors, fmt.Sprintf("server type is required for server: %s", server.Name))
			continue
		}

		// Validate server protocol
		validProtocols := []string{"kafka", "amqp", "mqtt", "nats", "redis", "webhook", "http", "db"}
		isValidProtocol := false
		for _, protocol := range validProtocols {
			if server.Type == protocol {
				isValidProtocol = true
				break
			}
		}
		if !isValidProtocol {
			errors = append(errors, fmt.Sprintf("invalid server type '%s' for server: %s", server.Type, server.Name))
		}

		// Check server name uniqueness
		serversManager := ServersObjectsManager{}
		if !serversManager.CanNameBeUsed(server.Name, projectID) {
			errors = append(errors, fmt.Sprintf("server name '%s' already exists in the project", server.Name))
		}

		serverTypes[server.Name] = server.Type
		serverResources[server.Name] = make(map[string]bool)

		// Validate resources
		for _, resource := range server.Resources {
			if resource.Name == "" {
				errors = append(errors, fmt.Sprintf("resource name is required for server: %s", server.Name))
				continue
			}

			// Validate resource using asyncresourceuri library
			// The library expects async+ prefix for protocols
			protocol := server.Type
			if protocol != "" && !strings.HasPrefix(protocol, "async+") {
				protocol = "async+" + protocol
			}
			resourceURI := fmt.Sprintf("%s://%s@%s/%s/%s", 
				protocol, server.Name, resource.Mode, resource.Type, resource.Name)
			_, err := asyncuri.ParseAsyncResourceReference(resourceURI)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid resource '%s' in server '%s': %v", resource.Name, server.Name, err))
				continue
			}

			if resource.Mode == "" {
				errors = append(errors, fmt.Sprintf("resource mode is required for resource: %s in server: %s", resource.Name, server.Name))
				continue
			}
			if resource.Type == "" {
				errors = append(errors, fmt.Sprintf("resource type is required for resource: %s in server: %s", resource.Name, server.Name))
				continue
			}

			// Validate mode
			validModes := []string{"read", "write", "readwrite"}
			isValidMode := false
			for _, mode := range validModes {
				if resource.Mode == mode {
					isValidMode = true
					break
				}
			}
			if !isValidMode {
				errors = append(errors, fmt.Sprintf("invalid mode '%s' for resource: %s in server: %s", resource.Mode, resource.Name, server.Name))
			}

			// Validate type
			validTypes := []string{"topic", "exchange", "queue", "table", "endpoint"}
			isValidType := false
			for _, t := range validTypes {
				if resource.Type == t {
					isValidType = true
					break
				}
			}
			if !isValidType {
				errors = append(errors, fmt.Sprintf("invalid type '%s' for resource: %s in server: %s", resource.Type, resource.Name, server.Name))
			}

			serverResources[server.Name][resource.Name] = true
			resourceKey := fmt.Sprintf("%s.%s", server.Name, resource.Name)
			resourceModes[resourceKey] = resource.Mode
			resourceTypes[resourceKey] = resource.Type
		}

		// Validate binds
		for _, bind := range server.Binds {
			if bind.Source == "" || bind.Target == "" {
				errors = append(errors, fmt.Sprintf("bind source and target are required for server: %s", server.Name))
				continue
			}
			if !serverResources[server.Name][bind.Source] {
				errors = append(errors, fmt.Sprintf("bind source '%s' not found in server: %s", bind.Source, server.Name))
			}
			if !serverResources[server.Name][bind.Target] {
				errors = append(errors, fmt.Sprintf("bind target '%s' not found in server: %s", bind.Target, server.Name))
			}
		}
	}

	// Validate schemas
	schemaNames := make(map[string]bool)
	for _, schema := range projectImport.Schemas {
		if schema.Name == "" {
			errors = append(errors, "schema name is required")
			continue
		}
		if schema.Type == "" {
			errors = append(errors, fmt.Sprintf("schema type is required for schema: %s", schema.Name))
			continue
		}
		if schema.Schema == "" {
			errors = append(errors, fmt.Sprintf("schema content is required for schema: %s", schema.Name))
			continue
		}

		// Validate schema type
		if schema.Type != "jsonschema" {
			errors = append(errors, fmt.Sprintf("invalid schema type '%s' for schema: %s (only 'jsonschema' is supported)", schema.Type, schema.Name))
		}

		// Validate JSON schema
		if schema.Type == "jsonschema" {
			_, err := jsonschema.CompileString("", schema.Schema)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid JSON schema for schema '%s': %v", schema.Name, err))
			}
		}

		schemaNames[schema.Name] = true

		// Check schema name uniqueness
		schemasManager := SchemaObjectsManager{}
		if !schemasManager.CanNameBeUsed(schema.Name, projectID) {
			errors = append(errors, fmt.Sprintf("schema name '%s' already exists in the project", schema.Name))
		}
	}

	// Validate messages
	messageNames := make(map[string]bool)
	for _, message := range projectImport.Messages {
		if message.Name == "" {
			errors = append(errors, "message name is required")
			continue
		}
		if message.Schema.Name == "" {
			errors = append(errors, fmt.Sprintf("schema reference is required for message: %s", message.Name))
			continue
		}
		if !schemaNames[message.Schema.Name] {
			errors = append(errors, fmt.Sprintf("schema '%s' referenced by message '%s' not found", message.Schema.Name, message.Name))
		}

		messageNames[message.Name] = true

		// Check message name uniqueness
		messagesManager := MessagesObjectsManager{}
		if !messagesManager.CanNameBeUsed(message.Name, projectID) {
			errors = append(errors, fmt.Sprintf("message name '%s' already exists in the project", message.Name))
		}
	}

	// Validate apps
	for _, app := range projectImport.Apps {
		if app.Name == "" {
			errors = append(errors, "app name is required")
			continue
		}

		// Check app name uniqueness
		appsManager := AppsObjectsManager{}
		if !appsManager.CanNameBeUsed(app.Name, projectID) {
			errors = append(errors, fmt.Sprintf("app name '%s' already exists in the project", app.Name))
		}

		// Validate sends
		for _, send := range app.Sends {
			if send.Message == "" {
				errors = append(errors, fmt.Sprintf("message is required for send in app: %s", app.Name))
				continue
			}
			if send.Resource == "" {
				errors = append(errors, fmt.Sprintf("resource is required for send in app: %s", app.Name))
				continue
			}
			if !messageNames[send.Message] {
				errors = append(errors, fmt.Sprintf("message '%s' referenced in app '%s' send not found", send.Message, app.Name))
			}

			// Validate resource reference using asyncresourceuri
			_, err := asyncuri.ParseAsyncResourceReference(send.Resource)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid resource reference '%s' in app '%s' send: %v", send.Resource, app.Name, err))
			}
		}

		// Validate receives
		for _, receive := range app.Receives {
			if receive.Message == "" {
				errors = append(errors, fmt.Sprintf("message is required for receive in app: %s", app.Name))
				continue
			}
			if receive.Resource == "" {
				errors = append(errors, fmt.Sprintf("resource is required for receive in app: %s", app.Name))
				continue
			}
			if !messageNames[receive.Message] {
				errors = append(errors, fmt.Sprintf("message '%s' referenced in app '%s' receive not found", receive.Message, app.Name))
			}

			// Validate resource reference using asyncresourceuri
			_, err := asyncuri.ParseAsyncResourceReference(receive.Resource)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid resource reference '%s' in app '%s' receive: %v", receive.Resource, app.Name, err))
			}
		}
	}

	return errors
}

// ImportProjectFromYAML imports the project architecture from YAML
func ImportProjectFromYAML(yamlStr string, projectID uuid.UUID, userID uuid.UUID) error {
	var projectImport ProjectImportYAML

	err := yaml.Unmarshal([]byte(yamlStr), &projectImport)
	if err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	// Import servers and resources
	serverIDMap := make(map[string]uuid.UUID)
	resourceIDMap := make(map[string]uuid.UUID)

	serversManager := ServersObjectsManager{}
	resourcesManager := ResourcesObjectsManager{}

	for _, server := range projectImport.Servers {
		// Create server
		serverObj, err := serversManager.CreateANewServer(
			server.Name,
			server.Description,
			server.Type,
			projectID,
			userID,
		)
		if err != nil {
			return fmt.Errorf("failed to create server %s: %v", server.Name, err)
		}
		serverIDMap[server.Name] = serverObj.GetID()

		// Create resources
		for _, resource := range server.Resources {
			resourceObj, err := resourcesManager.CreateANewResource(
				serverObj.GetID(),
				projectID,
				resource.Name,
				resource.Mode,
				resource.Type,
				resource.Description,
				userID,
			)
			if err != nil {
				return fmt.Errorf("failed to create resource %s in server %s: %v", resource.Name, server.Name, err)
			}
			resourceKey := fmt.Sprintf("%s.%s", server.Name, resource.Name)
			resourceIDMap[resourceKey] = resourceObj.GetID()
		}

		// Create bindings
		bindingsManager := ResourceBindingsObjectsManager{}
		for _, bind := range server.Binds {
			sourceKey := fmt.Sprintf("%s.%s", server.Name, bind.Source)
			targetKey := fmt.Sprintf("%s.%s", server.Name, bind.Target)
			
			sourceID, sourceExists := resourceIDMap[sourceKey]
			targetID, targetExists := resourceIDMap[targetKey]
			
			if !sourceExists || !targetExists {
				continue // Skip if resources not found
			}

			_, err := bindingsManager.CreateABinding(sourceID, targetID)
			if err != nil {
				return fmt.Errorf("failed to create binding between %s and %s: %v", bind.Source, bind.Target, err)
			}
		}
	}

	// Import schemas
	schemaIDMap := make(map[string]uuid.UUID)
	schemasManager := SchemaObjectsManager{}

	for _, schema := range projectImport.Schemas {
		schemaObj, err := schemasManager.CreateANewSchema(
			schema.Name,
			schema.Description,
			schema.Schema,
			schema.Type,
			"user",
			userID,
			userID,
			projectID,
		)
		if err != nil {
			return fmt.Errorf("failed to create schema %s: %v", schema.Name, err)
		}
		schemaIDMap[schema.Name] = schemaObj.GetID()
	}

	// Import messages
	messageIDMap := make(map[string]uuid.UUID)
	messagesManager := MessagesObjectsManager{}

	for _, message := range projectImport.Messages {
		schemaID, schemaExists := schemaIDMap[message.Schema.Name]
		if !schemaExists {
			return fmt.Errorf("schema %s not found for message %s", message.Schema.Name, message.Name)
		}

		// Get latest schema version
		schemaObj, _ := schemasManager.GetByID(schemaID)
		latestVersion := schemaObj.GetLatestVersion()

		messageObj, err := messagesManager.CreateANewMessage(
			message.Description,
			userID,
			projectID,
			message.Name,
			schemaID,
			latestVersion,
		)
		if err != nil {
			return fmt.Errorf("failed to create message %s: %v", message.Name, err)
		}
		messageIDMap[message.Name] = messageObj.GetID()
	}

	// Import apps
	appsManager := AppsObjectsManager{}
	appResourceMessagesManager := AppsResourcesMessagesObjectsManager{}
	
	for _, app := range projectImport.Apps {
		appObj, err := appsManager.CreateANewApp(
			app.Name,
			app.Description,
			projectID,
			userID,
		)
		if err != nil {
			return fmt.Errorf("failed to create app %s: %v", app.Name, err)
		}

		// Process app sends
		for _, send := range app.Sends {
			messageID, exists := messageIDMap[send.Message]
			if !exists {
				return fmt.Errorf("message '%s' not found for app '%s' send", send.Message, app.Name)
			}

			// Parse the resource URI to get resource ID
			parsedResource, err := asyncuri.ParseAsyncResourceReference(send.Resource)
			if err != nil {
				return fmt.Errorf("failed to parse resource URI for app '%s' send: %v", app.Name, err)
			}

			// Find the resource ID from the map
			resourceKey := fmt.Sprintf("%s.%s", parsedResource.Server, parsedResource.Name)
			resourceID, exists := resourceIDMap[resourceKey]
			if !exists {
				return fmt.Errorf("resource '%s' not found for app '%s' send (looking for key: %s)", send.Resource, app.Name, resourceKey)
			}

			// Create the app-resource-message connection for sends
			_, err = appResourceMessagesManager.CreateConnection(
				appObj.GetID(),
				resourceID,
				messageID,
				"sends",
				userID,
			)
			if err != nil {
				return fmt.Errorf("failed to create send connection for app '%s': %v", app.Name, err)
			}
		}

		// Process app receives
		for _, receive := range app.Receives {
			messageID, exists := messageIDMap[receive.Message]
			if !exists {
				return fmt.Errorf("message '%s' not found for app '%s' receive", receive.Message, app.Name)
			}

			// Parse the resource URI to get resource ID
			parsedResource, err := asyncuri.ParseAsyncResourceReference(receive.Resource)
			if err != nil {
				return fmt.Errorf("failed to parse resource URI for app '%s' receive: %v", app.Name, err)
			}

			// Find the resource ID from the map
			resourceKey := fmt.Sprintf("%s.%s", parsedResource.Server, parsedResource.Name)
			resourceID, exists := resourceIDMap[resourceKey]
			if !exists {
				return fmt.Errorf("resource '%s' not found for app '%s' receive (looking for key: %s)", receive.Resource, app.Name, resourceKey)
			}

			// Create the app-resource-message connection for receives
			_, err = appResourceMessagesManager.CreateConnection(
				appObj.GetID(),
				resourceID,
				messageID,
				"receives",
				userID,
			)
			if err != nil {
				return fmt.Errorf("failed to create receive connection for app '%s': %v", app.Name, err)
			}
		}
	}

	return nil
}