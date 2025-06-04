package knife4g

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	//go:embed front
	front embed.FS
)

type Config struct {
	RelativePath  string // 访问前缀，如 "/doc"
	ServerName    string // 服务名称
	OpenAPI       *OpenAPI3
	SwagResources []*SwaggerResource
}

// Knife4jServer Knife4j服务器结构
type Knife4jServer struct {
	config   *Config
	staticFS fs.FS
}

// SwaggerResource 表示 Swagger 资源信息
type SwaggerResource struct {
	ConfigURL         string `json:"configUrl"`
	OAuth2RedirectURL string `json:"oauth2RedirectUrl"`
	URL               string `json:"url"`
	ValidatorURL      string `json:"validatorUrl"`
	Name              string `json:"name"`
	Location          string `json:"location"`
	SwaggerVersion    string `json:"swaggerVersion"`
	TagSort           string `json:"tagSort"`
	OperationSort     string `json:"operationSort"`
}

// Handler 返回 knife4g 文档服务 http.Handler
func Handler(config *Config) http.Handler {
	server, err := NewKnife4jServer(config)
	if err != nil {
		log.Fatalf("Failed to create Knife4j server: %v", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := r.URL.Path
		if config.RelativePath != "" && strings.HasPrefix(path, config.RelativePath) {
			path = strings.TrimPrefix(path, config.RelativePath)
		}

		// 设置 CORS 头
		server.setCORSHeaders(w)

		// 记录请求信息
		log.Printf("处理请求: %s", path)

		switch path {
		case "/v3/api-docs":
			w.Header().Set("Content-Type", "application/json")
			server.handleOpenAPIDocs(w, r)
		case "/v3/api-docs/swagger-config":
			w.Header().Set("Content-Type", "application/json")
			server.handleSwaggerConfig(w, r)
		case "/doc.html", "/":
			// 处理 doc.html 和根路径，设置 HTML 内容类型
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			server.handleStaticFile(w, r)
		default:
			// 处理静态文件请求
			if strings.HasPrefix(path, "/webjars") || strings.HasPrefix(path, "/doc") {
				server.handleStaticFile(w, r)
			} else {
				http.NotFound(w, r)
			}
		}
	})
}

// NewKnife4jServer 创建新的Knife4j服务器实例
func NewKnife4jServer(cfg *Config) (*Knife4jServer, error) {
	// 获取front子目录的FS
	subFS, err := fs.Sub(front, "front")
	if err != nil {
		return nil, fmt.Errorf("failed to get front subdirectory: %v", err)
	}

	if cfg.SwagResources == nil {
		// 设置默认的 SwaggerResource
		defaultResources := []*SwaggerResource{
			{
				URL:               "/v3/api-docs",
				ConfigURL:         "/v3/api-docs/swagger-config",
				OAuth2RedirectURL: "/swagger-ui/oauth2-redirect.html",
				ValidatorURL:      "",
				Name:              cfg.ServerName,
				Location:          "/v3/api-docs",
				SwaggerVersion:    "3.0.3",
				TagSort:           "order",
				OperationSort:     "order",
			},
		}
		cfg.SwagResources = defaultResources
	}

	server := &Knife4jServer{
		config:   cfg,
		staticFS: subFS,
	}
	return server, nil
}

// handleOpenAPIDocs 处理 OpenAPI 文档请求
func (s *Knife4jServer) handleOpenAPIDocs(w http.ResponseWriter, r *http.Request) {
	if s.config.OpenAPI == nil {
		http.Error(w, "OpenAPI document not loaded", http.StatusInternalServerError)
		return
	}

	openAPI3 := convertToOpenAPI3(s.config.OpenAPI, s.config)
	w.Header().Set("Content-Type", "application/json")
	s.setCORSHeaders(w)

	if err := json.NewEncoder(w).Encode(openAPI3); err != nil {
		log.Printf("Failed to encode OpenAPI document: %v", err)
		http.Error(w, "Failed to encode OpenAPI document", http.StatusInternalServerError)
	}
}

// handleSwaggerConfig 处理 Swagger 配置请求
func (s *Knife4jServer) handleSwaggerConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.setCORSHeaders(w)

	// 记录请求信息
	log.Printf("处理 Swagger 配置请求")

	// 确保返回正确的 JSON 格式
	config := map[string]interface{}{
		"urls": s.config.SwagResources,
	}

	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("Failed to encode swagger config: %v", err)
		http.Error(w, "Failed to encode swagger config", http.StatusInternalServerError)
	}
}

// handleStaticFile 处理静态文件请求
func (s *Knife4jServer) handleStaticFile(w http.ResponseWriter, r *http.Request) {
	// 获取请求路径
	path := strings.TrimPrefix(r.URL.Path, "/")

	// 处理根路径和默认文件
	if path == "" || path == "doc.html" {
		path = "doc.html"
	}

	log.Printf("尝试打开文件: %s", path)

	// 尝试打开文件
	file, err := s.staticFS.Open(path)
	if err != nil {
		log.Printf("Failed to open static file: %v, path: %s", err, path)
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	// 设置内容类型
	if path == "doc.html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else {
		s.setContentType(w, filepath.Ext(path))
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	// 复制文件内容到响应
	io.Copy(w, file)
}

// setCORSHeaders 设置CORS头
func (s *Knife4jServer) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// setContentType 设置内容类型
func (s *Knife4jServer) setContentType(w http.ResponseWriter, ext string) {
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".woff", ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	case ".ttf":
		w.Header().Set("Content-Type", "font/ttf")
	case ".eot":
		w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	}
}

// convertToOpenAPI3 将 OpenAPI 对象转换为标准的 OpenAPI 3.0 JSON 结构
func convertToOpenAPI3(openapi *OpenAPI3, config *Config) map[string]interface{} {
	result := make(map[string]interface{})

	// 基本信息
	result["openapi"] = "3.0.1" // 使用固定版本

	// 构建 info 对象
	info := map[string]interface{}{
		"title":   openapi.Info.Title,
		"version": openapi.Info.Version,
		"name":    config.ServerName, // 服务名称
	}

	// 解析 info 的注释
	infoParser := NewCommentParser().Parse(openapi.Info.Description)

	// 从解析器中获取标签值
	if infoParser.HasTag("description") {
		info["description"] = infoParser.GetString("description")
	}

	result["info"] = info

	// 处理 servers
	if len(openapi.Servers) > 0 {
		servers := make([]map[string]interface{}, len(openapi.Servers))
		for i, server := range openapi.Servers {
			serverMap := map[string]interface{}{
				"url":         server.URL,
				"description": server.Description,
			}
			if len(server.Variables) > 0 {
				variables := make(map[string]interface{})
				for name, variable := range server.Variables {
					variables[name] = map[string]interface{}{
						"default":     variable.Default,
						"description": variable.Description,
						"enum":        variable.Enum,
					}
				}
				serverMap["variables"] = variables
			}
			servers[i] = serverMap
		}
		result["servers"] = servers
	} else {
		// 如果没有配置服务器，添加默认服务器
		result["servers"] = []map[string]interface{}{
			{
				"url":         "http://localhost:8000",
				"description": "Generated server url",
			},
		}
	}

	// 处理 paths
	paths := make(map[string]interface{})
	for path, pathItem := range openapi.Paths {
		pathMap := make(map[string]interface{})

		// 处理各种 HTTP 方法
		if pathItem.Get != nil {
			pathMap["get"] = convertOperationToOpenAPI3(pathItem.Get)
		}
		if pathItem.Post != nil {
			pathMap["post"] = convertOperationToOpenAPI3(pathItem.Post)
		}
		if pathItem.Put != nil {
			pathMap["put"] = convertOperationToOpenAPI3(pathItem.Put)
		}
		if pathItem.Delete != nil {
			pathMap["delete"] = convertOperationToOpenAPI3(pathItem.Delete)
		}
		if pathItem.Patch != nil {
			pathMap["patch"] = convertOperationToOpenAPI3(pathItem.Patch)
		}

		paths[path] = pathMap
	}
	result["paths"] = paths

	// 处理 components
	components := make(map[string]interface{})
	components["schemas"] = convertSchemasToOpenAPI3(openapi.Components.Schemas)
	result["components"] = components

	return result
}

// convertOperationToOpenAPI3 将 Operation 转换为 OpenAPI 3.0 格式
func convertOperationToOpenAPI3(op *Operation) map[string]interface{} {
	result := make(map[string]interface{})

	// 基本信息
	result["tags"] = op.Tags
	result["summary"] = op.Summary
	result["operationId"] = op.OperationID

	// 使用注释解析器处理description 信息
	parser := NewCommentParser().Parse(op.Description)
	// 从解析器中获取标签值
	if parser.HasTag("description") {
		result["description"] = parser.GetString("description")
	}

	// 处理请求体
	if op.RequestBody != nil {
		requestBody := make(map[string]interface{})
		requestBody["required"] = op.RequestBody.Required
		requestBody["content"] = convertContentToOpenAPI3(op.RequestBody.Content)
		result["requestBody"] = requestBody
	}

	// 处理响应
	responses := make(map[string]interface{})
	for code, response := range op.Responses {
		responseMap := make(map[string]interface{})
		responseMap["description"] = response.Description
		if response.Content != nil {
			responseMap["content"] = convertContentToOpenAPI3(response.Content)
		}
		responses[code] = responseMap
	}
	result["responses"] = responses

	return result
}

// convertContentToOpenAPI3 将 Content 转换为 OpenAPI 3.0 格式
func convertContentToOpenAPI3(content map[string]MediaType) map[string]interface{} {
	result := make(map[string]interface{})
	for contentType, mediaType := range content {
		mediaTypeMap := make(map[string]interface{})
		if mediaType.Schema != nil {
			mediaTypeMap["schema"] = convertSchemaToOpenAPI3(mediaType.Schema)
		}
		if mediaType.Example != nil {
			mediaTypeMap["example"] = mediaType.Example
		}
		result[contentType] = mediaTypeMap
	}
	return result
}

// convertSchemasToOpenAPI3 将 Schemas 转换为 OpenAPI 3.0 格式
func convertSchemasToOpenAPI3(schemas map[string]Schema) map[string]interface{} {
	result := make(map[string]interface{})
	for name, schema := range schemas {
		result[name] = convertSchemaToOpenAPI3(&schema)
	}
	return result
}

// convertSchemaToOpenAPI3 将 Schema 转换为 OpenAPI 3.0 格式
func convertSchemaToOpenAPI3(schema *Schema) map[string]interface{} {
	if schema == nil {
		return nil
	}

	result := make(map[string]interface{})

	// 基本属性
	if schema.Type != "" {
		result["type"] = schema.Type
	}
	if schema.Format != "" {
		result["format"] = schema.Format
	}
	if schema.Title != "" {
		result["title"] = schema.Title
	}
	if schema.Description != "" {
		result["description"] = schema.Description
	}
	if schema.Default != nil {
		result["default"] = schema.Default
	}

	// 数值相关属性
	if schema.MultipleOf != nil {
		result["multipleOf"] = schema.MultipleOf
	}
	if schema.Maximum != nil {
		result["maximum"] = schema.Maximum
	}
	if schema.Minimum != nil {
		result["minimum"] = schema.Minimum
	}
	result["exclusiveMaximum"] = schema.ExclusiveMaximum
	result["exclusiveMinimum"] = schema.ExclusiveMinimum

	// 字符串相关属性
	if schema.MaxLength != nil {
		result["maxLength"] = schema.MaxLength
	}
	if schema.MinLength != nil {
		result["minLength"] = schema.MinLength
	}
	if schema.Pattern != "" {
		result["pattern"] = schema.Pattern
	}

	// 数组相关属性
	if schema.MaxItems != nil {
		result["maxItems"] = schema.MaxItems
	}
	if schema.MinItems != nil {
		result["minItems"] = schema.MinItems
	}
	result["uniqueItems"] = schema.UniqueItems

	// 对象相关属性
	if schema.MaxProperties != nil {
		result["maxProperties"] = schema.MaxProperties
	}
	if schema.MinProperties != nil {
		result["minProperties"] = schema.MinProperties
	}
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	// 枚举值
	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}

	// 属性定义
	if schema.Properties != nil {
		properties := make(map[string]interface{})
		for name, prop := range schema.Properties {
			properties[name] = convertSchemaToOpenAPI3(prop)
		}
		result["properties"] = properties
	}

	// 引用
	if schema.Ref != "" {
		result["$ref"] = schema.Ref
	}

	// 其他属性
	result["nullable"] = schema.Nullable
	result["readOnly"] = schema.ReadOnly
	result["writeOnly"] = schema.WriteOnly
	result["deprecated"] = schema.Deprecated

	return result
}
