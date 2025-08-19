# Code Generation Templates

This directory contains templates used for generating code from AsyncAPI schemas and applications.

## Structure

```
templates/
└── go/                          # Go language templates
    ├── template.tmpl           # Main template that combines all parts
    ├── interfaces.tmpl         # Interface definitions
    ├── schema_implementation.tmpl    # Schema struct implementations
    ├── message_implementation.tmpl   # Message implementations
    ├── resource_implementation.tmpl  # Resource implementations
    ├── server_implementation.tmpl    # Server implementations
    └── app_implementation.tmpl       # Application implementations
```

## Adding New Languages

To add support for a new language:

1. Create a new directory under `templates/` with the language name (e.g., `python`, `typescript`)
2. Add the required template files for that language
3. Update the code generation logic to support the new language

## Template Variables

The templates use Go's `text/template` syntax and receive the following data structures:

- **Schemas**: List of schema definitions with ID, Name, Version, Description, and generated Code
- **Messages**: List of messages with ID, Name, Description, SchemaID, and SchemaVersion
- **Resources**: List of resources with ID, Name, Description, Type, Mode, and ServerName
- **Servers**: List of servers with ID, Name, Description, Protocol, and associated Resources
- **Apps**: List of applications with ID, Name, Description, and their Sends/Receives mappings

## Environment Variable

The path to this templates directory must be set in the `PATH_TO_STUBS_TEMPLATES_FOLDER` environment variable.

Default value in `.env.template`: `./templates`