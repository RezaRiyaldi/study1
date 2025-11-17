package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type EndpointInfo struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Parameters  []ParameterInfo
	Responses   []ResponseInfo
	HandlerName string
	ModuleName  string
	HasBody     bool
}

type ParameterInfo struct {
	Name        string
	In          string
	Type        string
	Required    bool
	Description string
	Default     string
}

type ResponseInfo struct {
	Code        string
	Description string
	Type        string
}

type ModelInfo struct {
	Name      string
	Fields    []FieldInfo
	IsRequest bool
	Module    string
}

type FieldInfo struct {
	Name     string
	Type     string
	JSONTag  string
	GormTag  string
	Required bool
	Example  string
}

type ModuleInfo struct {
	Name      string
	Endpoints []EndpointInfo
	Models    []ModelInfo
	Requests  []ModelInfo
}

func main() {
	log.Println("🔄 Generating API documentation...")

	// Scan modules directory for handlers
	modulesPath := "./internal/modules"
	docsPath := "./docs"

	// Ensure docs directory exists
	if err := os.MkdirAll(docsPath, 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	// Collect all modules
	modules := make(map[string]*ModuleInfo)

	// Walk through modules directory
	err := filepath.Walk(modulesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip root modules directory
			if path == modulesPath {
				return nil
			}

			moduleName := filepath.Base(path)
			if _, exists := modules[moduleName]; !exists {
				modules[moduleName] = &ModuleInfo{
					Name: moduleName,
				}
			}
			return nil
		}

		// Process files in module directories
		dir := filepath.Dir(path)
		moduleName := filepath.Base(dir)

		if moduleName == "modules" {
			return nil // Skip root
		}

		module := modules[moduleName]
		if module == nil {
			module = &ModuleInfo{Name: moduleName}
			modules[moduleName] = module
		}

		switch {
		case strings.HasSuffix(info.Name(), "handler.go"):
			endpoints, models := parseHandlerFile(path)
			module.Endpoints = append(module.Endpoints, endpoints...)
			module.Models = append(module.Models, models...)

		case strings.HasSuffix(info.Name(), "model.go"):
			models := parseModelFile(path)
			module.Models = append(module.Models, models...)

		case strings.HasSuffix(info.Name(), "dto.go"):
			requests := parseModelFile(path)
			module.Requests = append(module.Requests, requests...)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error scanning modules: %v", err)
	}

	// Generate documentation for each module
	for moduleName, module := range modules {
		if len(module.Endpoints) > 0 {
			log.Printf("📦 Processing module: %s (%d endpoints)", moduleName, len(module.Endpoints))

			// Generate OpenAPI documentation
			generateOpenAPIDocs(module, filepath.Join(docsPath, fmt.Sprintf("%s_openapi.yaml", moduleName)))

			// Generate Markdown documentation
			generateMarkdownDocs(module.Endpoints, module.Models, filepath.Join(docsPath, fmt.Sprintf("%s_api.md", moduleName)))

			// Generate Postman collection
			generatePostmanCollection(module.Endpoints, module.Models, filepath.Join(docsPath, fmt.Sprintf("%s_postman.json", moduleName)))
		}
	}

	// Generate main documentation file that references all modules
	generateMainDocumentation(modules, docsPath)

	log.Println("✅ API documentation generated successfully!")
}

// ... (extractJSONTag, extractGormTag, extractTagValue functions tetap sama)

// (previous generateExample with moduleName removed; keep later 2-arg version below)

func generateOpenAPIDocs(module *ModuleInfo, outputPath string) {
	// Group endpoints by path
	endpointsByPath := make(map[string][]EndpointInfo)
	for _, endpoint := range module.Endpoints {
		endpointsByPath[endpoint.Path] = append(endpointsByPath[endpoint.Path], endpoint)
	}

	tmpl := `openapi: 3.0.3
info:
  title: {{ .Module.Name }} API
  description: Auto-generated API documentation for {{ .Module.Name }}
  version: 1.0.0

servers:
  - url: http://localhost:8080
    description: Development server

paths:
{{- range $path, $endpoints := .EndpointsByPath }}
  {{ $path }}:
    {{- range $endpoints }}
    {{ .Method | toLower }}:
      summary: {{ .Summary }}
      description: {{ .Description }}
      tags:
        - {{ .ModuleName }}
      {{- if .HasBody }}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{{ .Method }}{{ .ModuleName }}Request'
            example: {{ generateExample .ModuleName .Method }}
      {{- end }}
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Response'
              example: {{ generateResponseExample .ModuleName }}
        '201':
          description: Created
        '400':
          description: Bad Request
        '404':
          description: Not Found
        '500':
          description: Internal Server Error
    {{- end }}
{{- end }}

components:
  schemas:
    Response:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
        error:
          type: string
        meta:
          type: object
          properties:
            page:
              type: integer
            page_size:
              type: integer
            total:
              type: integer
            pages:
              type: integer

    {{- range .Module.Requests }}
    {{ .Name }}:
      type: object
      properties:
      {{- range .Fields }}
        {{ if .JSONTag }}{{ .JSONTag }}{{ else }}{{ .Name | toLower }}{{ end }}:
          type: {{ .Type | getOpenAPIType }}
          {{- if .Required }}
          required: true
          {{- end }}
          example: {{ .Example }}
      {{- end }}
    {{- end }}

    {{- range .Module.Models }}
    {{ .Name }}:
      type: object
      properties:
      {{- range .Fields }}
        {{ if .JSONTag }}{{ .JSONTag }}{{ else }}{{ .Name | toLower }}{{ end }}:
          type: {{ .Type | getOpenAPIType }}
          example: {{ .Example }}
      {{- end }}
    {{- end }}
`

	funcMap := template.FuncMap{
		"getOpenAPIType":          getOpenAPIType,
		"toLower":                 strings.ToLower,
		"generateExample":         generateRequestBodyExample,
		"generateResponseExample": generateResponseBodyExample,
	}

	template := template.Must(template.New("openapi").Funcs(funcMap).Parse(tmpl))

	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create OpenAPI file for %s: %v", module.Name, err)
	}
	defer file.Close()

	data := struct {
		Module          *ModuleInfo
		EndpointsByPath map[string][]EndpointInfo
	}{
		Module:          module,
		EndpointsByPath: endpointsByPath,
	}

	if err := template.Execute(file, data); err != nil {
		log.Fatalf("Failed to generate OpenAPI docs for %s: %v", module.Name, err)
	}
}

func generateRequestBodyExample(moduleName, method string) string {
	switch method {
	case "POST":
		return fmt.Sprintf(`{
  "name": "New %s",
  "email": "%s@example.com",
  "age": 25
}`, strings.Title(moduleName), moduleName)
	case "PUT":
		return fmt.Sprintf(`{
  "name": "Updated %s",
  "email": "%s.updated@example.com", 
  "age": 26
}`, strings.Title(moduleName), moduleName)
	default:
		return "{}"
	}
}

func generateResponseBodyExample(moduleName string) string {
	return fmt.Sprintf(`{
  "success": true,
  "data": {
    "id": 1,
    "name": "Example %s",
    "email": "%s@example.com",
    "age": 25,
    "created_at": "2023-10-20T10:00:00Z",
    "updated_at": "2023-10-20T10:00:00Z"
  }
}`, strings.Title(moduleName), moduleName)
}

func parseHandlerFile(filePath string) ([]EndpointInfo, []ModelInfo) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Warning: Could not parse %s: %v", filePath, err)
		return nil, nil
	}

	var endpoints []EndpointInfo
	var models []ModelInfo

	// Extract module name from file path first
	moduleName := extractModuleName(filePath)

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if isHandlerMethod(x) {
				endpoint := extractEndpointInfo(x, filePath, moduleName)
				if endpoint.Method != "" {
					endpoints = append(endpoints, endpoint)
				}
			}
		case *ast.GenDecl:
			// Extract models from type definitions
			if x.Tok == token.TYPE {
				for _, spec := range x.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							model := extractModelInfo(typeSpec.Name.Name, structType)
							models = append(models, model)
						}
					}
				}
			}
		}
		return true
	})

	return endpoints, models
}

func parseModelFile(filePath string) []ModelInfo {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Warning: Could not parse %s: %v", filePath, err)
		return nil
	}

	var models []ModelInfo

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				for _, spec := range x.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							model := extractModelInfo(typeSpec.Name.Name, structType)
							models = append(models, model)
						}
					}
				}
			}
		}
		return true
	})

	return models
}

func isHandlerMethod(fn *ast.FuncDecl) bool {
	// Check if it's a handler method (has receiver with *Handler)
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		if starExpr, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				return strings.HasSuffix(ident.Name, "Handler")
			}
		}
	}
	return false
}

func extractEndpointInfo(fn *ast.FuncDecl, filePath string, moduleName string) EndpointInfo {
	endpoint := EndpointInfo{
		HandlerName: fn.Name.Name,
		ModuleName:  moduleName,
		Tags:        []string{moduleName}, // Default tag
	}

	// Extract from comments
	if fn.Doc != nil {
		for _, comment := range fn.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

			if strings.Contains(text, "@Summary") {
				endpoint.Summary = strings.TrimSpace(strings.TrimPrefix(text, "@Summary"))
			} else if strings.Contains(text, "@Description") {
				endpoint.Description = strings.TrimSpace(strings.TrimPrefix(text, "@Description"))
			} else if strings.Contains(text, "@Tags") {
				tags := strings.TrimSpace(strings.TrimPrefix(text, "@Tags"))
				endpoint.Tags = strings.Split(tags, ",")
				for i, tag := range endpoint.Tags {
					endpoint.Tags[i] = strings.TrimSpace(tag)
				}
			} else if strings.Contains(text, "@Router") {
				parts := strings.Fields(strings.TrimPrefix(text, "@Router"))
				if len(parts) >= 2 {
					endpoint.Path = parts[0]
					endpoint.Method = strings.Trim(parts[1], "[]")
				}
			}
		}
	}

	// If no annotations found, generate from function name and parameters
	if endpoint.Method == "" {
		endpoint = generateEndpointFromFunction(fn, endpoint)
	}

	return endpoint
}

func generateEndpointFromFunction(fn *ast.FuncDecl, endpoint EndpointInfo) EndpointInfo {
	funcName := fn.Name.Name

	// Determine HTTP method from function name
	switch {
	case strings.HasPrefix(funcName, "Get"):
		endpoint.Method = "GET"
		if strings.HasPrefix(funcName, "GetAll") || strings.HasPrefix(funcName, "List") {
			endpoint.Summary = "Get all " + endpoint.ModuleName
		} else {
			endpoint.Summary = "Get " + endpoint.ModuleName + " by ID"
		}
	case strings.HasPrefix(funcName, "Create") || strings.HasPrefix(funcName, "Add"):
		endpoint.Method = "POST"
		endpoint.Summary = "Create " + endpoint.ModuleName
	case strings.HasPrefix(funcName, "Update"):
		endpoint.Method = "PUT"
		endpoint.Summary = "Update " + endpoint.ModuleName
	case strings.HasPrefix(funcName, "Delete"):
		endpoint.Method = "DELETE"
		endpoint.Summary = "Delete " + endpoint.ModuleName
	default:
		endpoint.Method = "GET" // default
		endpoint.Summary = funcName
	}

	// Generate path from function name and module
	basePath := "/api/v1/" + endpoint.ModuleName

	switch {
	case strings.Contains(funcName, "ByID") || (!strings.HasPrefix(funcName, "GetAll") && !strings.HasPrefix(funcName, "List") && strings.HasPrefix(funcName, "Get")):
		endpoint.Path = basePath + "/{id}"
	case strings.HasPrefix(funcName, "GetAll") || strings.HasPrefix(funcName, "List"):
		endpoint.Path = basePath
	default:
		endpoint.Path = basePath
	}

	// Set default description if empty
	if endpoint.Description == "" {
		endpoint.Description = endpoint.Summary
	}

	return endpoint
}

func extractModuleName(filePath string) string {
	dir := filepath.Dir(filePath)
	module := filepath.Base(dir)

	// Clean up module name
	module = strings.ToLower(module)
	if strings.HasSuffix(module, "s") && module != "users" {
		module = strings.TrimSuffix(module, "s")
	}

	return module
}

func extractModelInfo(name string, structType *ast.StructType) ModelInfo {
	model := ModelInfo{Name: name}

	if structType.Fields == nil {
		return model
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		fieldInfo := FieldInfo{
			Name: field.Names[0].Name,
		}

		// Extract type
		if ident, ok := field.Type.(*ast.Ident); ok {
			fieldInfo.Type = ident.Name
		} else if starExpr, ok := field.Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				fieldInfo.Type = "*" + ident.Name
			}
		} else if selector, ok := field.Type.(*ast.SelectorExpr); ok {
			if ident, ok := selector.X.(*ast.Ident); ok {
				fieldInfo.Type = ident.Name + "." + selector.Sel.Name
			}
		}

		// Extract tags
		if field.Tag != nil {
			tag := field.Tag.Value
			fieldInfo.JSONTag = extractJSONTag(tag)
			fieldInfo.GormTag = extractGormTag(tag)

			// Determine if required
			fieldInfo.Required = strings.Contains(tag, `binding:"required"`) ||
				!strings.Contains(fieldInfo.JSONTag, "omitempty")

			// Generate example
			fieldInfo.Example = generateExample(fieldInfo.Type, fieldInfo.Name)
		}

		model.Fields = append(model.Fields, fieldInfo)
	}

	return model
}

func extractJSONTag(tag string) string {
	return extractTagValue(tag, "json")
}

func extractGormTag(tag string) string {
	return extractTagValue(tag, "gorm")
}

func extractTagValue(tag, key string) string {
	tag = strings.Trim(tag, "`")
	parts := strings.Split(tag, " ")

	for _, part := range parts {
		if strings.HasPrefix(part, key+":") {
			value := strings.TrimPrefix(part, key+":")
			return strings.Trim(value, `"`)
		}
	}
	return ""
}

func generateExample(fieldType, fieldName string) string {
	fieldName = strings.ToLower(fieldName)

	switch {
	case strings.Contains(fieldName, "email"):
		return "user@example.com"
	case strings.Contains(fieldName, "name"):
		return "John Doe"
	case strings.Contains(fieldName, "id"):
		return "1"
	case strings.Contains(fieldType, "string"):
		return "example"
	case strings.Contains(fieldType, "int"), strings.Contains(fieldType, "float"):
		return "123"
	case strings.Contains(fieldType, "bool"):
		return "true"
	case strings.Contains(fieldType, "time"):
		return "2023-10-20T10:00:00Z"
	default:
		return "example"
	}
}

func generateMarkdownDocs(endpoints []EndpointInfo, models []ModelInfo, outputPath string) {
	tmpl := `# API Documentation

## Endpoints

{{- range .Endpoints }}
### {{ .Method }} {{ .Path }}

**Summary:** {{ .Summary }}

**Description:** {{ .Description }}

**Tags:** {{ range .Tags }}{{ . }} {{ end }}

**Handler:** {{ .HandlerName }}

---

{{- end }}

## Data Models

{{- range .Models }}
### {{ .Name }}

| Field | Type | JSON Tag | Required | Example |
|-------|------|----------|----------|---------|
{{- range .Fields }}
| {{ .Name }} | {{ .Type }} | {{ .JSONTag }} | {{ .Required }} | {{ .Example }} |
{{- end }}

{{- end }}
`

	template := template.Must(template.New("markdown").Parse(tmpl))

	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create Markdown file: %v", err)
	}
	defer file.Close()

	data := struct {
		Endpoints []EndpointInfo
		Models    []ModelInfo
	}{
		Endpoints: endpoints,
		Models:    models,
	}

	if err := template.Execute(file, data); err != nil {
		log.Fatalf("Failed to generate Markdown docs: %v", err)
	}
}

func generatePostmanCollection(endpoints []EndpointInfo, models []ModelInfo, outputPath string) {
	// Simplified Postman collection generator
	tmpl := `{
  "info": {
    "name": "Study1 API",
    "description": "Auto-generated Postman Collection",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {{- range $index, $endpoint := .Endpoints }}
    {
      "name": "{{ $endpoint.Summary }}",
      "request": {
        "method": "{{ $endpoint.Method }}",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "url": {
          "raw": "http://localhost:8080{{ $endpoint.Path }}",
          "protocol": "http",
          "host": ["localhost"],
          "port": "8080",
          "path": [{{ range $i, $part := splitPath $endpoint.Path }}{{ if $i }},{{ end }}"{{ $part }}"{{ end }}]
        }
      },
      "response": []
    }{{ if not (last $index $.Endpoints) }},{{ end }}
    {{- end }}
  ]
}`

	funcMap := template.FuncMap{
		"splitPath": func(path string) []string {
			return strings.Split(strings.Trim(path, "/"), "/")
		},
		"last": func(index int, endpoints []EndpointInfo) bool {
			return index == len(endpoints)-1
		},
	}

	template := template.Must(template.New("postman").Funcs(funcMap).Parse(tmpl))

	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create Postman file: %v", err)
	}
	defer file.Close()

	data := struct {
		Endpoints []EndpointInfo
		Models    []ModelInfo
	}{
		Endpoints: endpoints,
		Models:    models,
	}

	if err := template.Execute(file, data); err != nil {
		log.Fatalf("Failed to generate Postman collection: %v", err)
	}
}

func getOpenAPIType(goType string) string {
	switch {
	case strings.Contains(goType, "string"):
		return "string"
	case strings.Contains(goType, "int"), strings.Contains(goType, "uint"):
		return "integer"
	case strings.Contains(goType, "float"):
		return "number"
	case strings.Contains(goType, "bool"):
		return "boolean"
	case strings.Contains(goType, "time"):
		return "string"
	default:
		return "string"
	}
}

// generateMainDocumentation creates a simple index file listing generated module docs.
func generateMainDocumentation(modules map[string]*ModuleInfo, docsPath string) {
	filePath := filepath.Join(docsPath, "index.md")
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create main documentation file: %v", err)
	}
	defer f.Close()

	fmt.Fprintln(f, "# API Documentation Index")
	fmt.Fprintln(f, "This file lists generated documentation for each module:")

	for name := range modules {
		openapi := fmt.Sprintf("%s_openapi.yaml", name)
		md := fmt.Sprintf("%s_api.md", name)
		postman := fmt.Sprintf("%s_postman.json", name)

		fmt.Fprintf(f, "- **%s**: [OpenAPI](%s), [Markdown](%s), [Postman](%s)\n", name, openapi, md, postman)
	}
}
