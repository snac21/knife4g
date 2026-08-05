package knife4g

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	// MIME 媒体类型常量
	MIMEApplicationJSON   = "application/json"
	MIMEMultipartFormData = "multipart/form-data"

	// OpenAPI Parameter 位置与数据类型常量
	ParamInFormData   = "formData"
	ParamTypeFile     = "file"
	ParamTypeString   = "string"
	ParamFormatBinary = "binary"

	// Knife4g 注释扩展 Key 常量
	TagConsumes    = "consumes"
	TagFile        = "file"
	TagDescription = "description"
	TagSummary     = "summary"
	TagOperationID = "operationId"
	TagTags        = "tags"
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
		slog.Debug("处理请求", "path", path)

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
		slog.Debug("Failed to encode OpenAPI document", "err", err)
		http.Error(w, "Failed to encode OpenAPI document", http.StatusInternalServerError)
	}
}

// handleSwaggerConfig 处理 Swagger 配置请求
func (s *Knife4jServer) handleSwaggerConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.setCORSHeaders(w)

	// 记录请求信息
	slog.Debug("处理 Swagger 配置请求")

	// 确保返回正确的 JSON 格式
	config := map[string]any{
		"urls": s.config.SwagResources,
	}

	if err := json.NewEncoder(w).Encode(config); err != nil {
		slog.Debug("Failed to encode swagger config", "err", err)
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

	slog.Debug("尝试打开doc.html文件", "path", path)

	// 尝试打开文件
	file, err := s.staticFS.Open(path)
	if err != nil {
		slog.Debug("Failed to open static file", "path", path, "err", err)
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
func convertToOpenAPI3(openapi *OpenAPI3, config *Config) map[string]any {
	result := make(map[string]any)

	// 基本信息
	if openapi.OpenAPI != "" {
		result["openapi"] = openapi.OpenAPI
	} else {
		result["openapi"] = "3.0.1"
	}

	// 构建 info 对象
	info := map[string]any{
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
		servers := make([]map[string]any, len(openapi.Servers))
		for i, server := range openapi.Servers {
			serverMap := map[string]any{
				"url":         server.URL,
				"description": server.Description,
			}
			if len(server.Variables) > 0 {
				variables := make(map[string]any)
				for name, variable := range server.Variables {
					variables[name] = map[string]any{
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
	}

	// 处理全局 tags 列表
	if len(openapi.Tags) > 0 {
		tags := make([]map[string]any, 0, len(openapi.Tags))
		for _, tag := range openapi.Tags {
			tagName := tag.Name
			tagDesc := tag.Description

			// 解析 Service 节点注释文本
			//（位于 tag.Description 中）
			// 提取自定义的 @tags: 标签名与 @description: 描述信息
			if tag.Description != "" {
				parser := NewCommentParser().Parse(tag.Description)
				if parser.HasTag("tags") {
					if customTag := parser.GetString("tags"); customTag != "" {
						// 使用 Proto 中定义的 Service 级 @tags 替换默认的服务 Tag 名称
						tagName = customTag
					}
				}
				if parser.HasTag("description") {
					tagDesc = parser.GetString("description") // 提取过滤掉扩展标记后的纯文本描述
				}
			} else {
				// 若单proto服务编译时 tags description 为空，
				// 解析 info 里的 tags 定义
				if infoParser.HasTag("tags") {
					if customTag := infoParser.GetString("tags"); customTag != "" {
						tagName = customTag
					}
				}
			}

			tagMap := map[string]any{
				"name": tagName,
			}
			if tagDesc != "" {
				tagMap["description"] = tagDesc
			}
			if tag.ExternalDocs != nil {
				tagMap["externalDocs"] = map[string]any{
					"description": tag.ExternalDocs.Description,
					"url":         tag.ExternalDocs.URL,
				}
			}
			tags = append(tags, tagMap)
		}
		result["tags"] = tags
	}

	// 处理 paths
	paths := make(map[string]any)
	for path, pathItem := range openapi.Paths {
		pathMap := make(map[string]any)

		// 处理各种 HTTP 方法
		if pathItem.Get != nil {
			pathMap["get"] = convertOperationToOpenAPI3(pathItem.Get, openapi.Components.Schemas)
		}
		if pathItem.Post != nil {
			pathMap["post"] = convertOperationToOpenAPI3(pathItem.Post, openapi.Components.Schemas)
		}
		if pathItem.Put != nil {
			pathMap["put"] = convertOperationToOpenAPI3(pathItem.Put, openapi.Components.Schemas)
		}
		if pathItem.Delete != nil {
			pathMap["delete"] = convertOperationToOpenAPI3(pathItem.Delete, openapi.Components.Schemas)
		}
		if pathItem.Patch != nil {
			pathMap["patch"] = convertOperationToOpenAPI3(pathItem.Patch, openapi.Components.Schemas)
		}

		paths[path] = pathMap
	}
	result["paths"] = paths

	// 处理 components
	components := make(map[string]any)
	components["schemas"] = convertSchemasToOpenAPI3(openapi.Components.Schemas)
	result["components"] = components

	return result
}

// getTargetSchema 极简解开 Operation 请求体中 $ref 所引用的 Component Schema 节点
func getTargetSchema(op *Operation, componentsSchemas map[string]Schema) *Schema {
	if op.RequestBody == nil || op.RequestBody.Content == nil {
		return nil
	}
	for _, media := range op.RequestBody.Content {
		if media.Schema != nil {
			if media.Schema.Ref != "" {
				schemaName := filepath.Base(media.Schema.Ref)
				if compSchema, exists := componentsSchemas[schemaName]; exists {
					return &compSchema
				}
			} else {
				return media.Schema
			}
		}
	}
	return nil
}

// convertOperationToOpenAPI3 将 Operation 转换为 OpenAPI 3.0 格式并解析注释扩展指令
func convertOperationToOpenAPI3(op *Operation, componentsSchemas map[string]Schema) map[string]any {
	result := make(map[string]any)

	// 基本信息
	result["tags"] = op.Tags
	result["summary"] = op.Summary
	result["operationId"] = op.OperationID

	// 使用注释解析器处理 RPC 操作的 description 文本
	parser := NewCommentParser().Parse(op.Description)
	if parser.HasTag(TagDescription) {
		result["description"] = parser.GetString(TagDescription)
	}
	if (result["summary"] == nil || result["summary"] == "") && parser.HasTag(TagSummary) {
		result["summary"] = parser.GetString(TagSummary)
	}
	if (result["operationId"] == nil || result["operationId"] == "") && parser.HasTag(TagOperationID) {
		result["operationId"] = parser.GetString(TagOperationID)
	}
	if parser.HasTag(TagTags) && len(parser.GetArray(TagTags)) > 0 {
		result["tags"] = parser.GetArray(TagTags)
	}

	// 解开 $ref 追溯查找当前 Operation 请求体所引用的 Component Schema 节点
	targetSchema := getTargetSchema(op, componentsSchemas)

	// 根据 RPC 方法注释上的 @consumes 精准认定文件上传接口
	consumesVal := parser.GetString(TagConsumes)
	isFileOperation := consumesVal == MIMEMultipartFormData || parser.HasTag(TagFile)

	if isFileOperation {
		result["consumes"] = []string{MIMEMultipartFormData}
		result["produces"] = []string{MIMEApplicationJSON}
	} else if consumesVal != "" {
		result["consumes"] = []string{consumesVal}
	}

	// 处理请求体
	if op.RequestBody != nil {
		requestBody := make(map[string]any)
		requestBody["required"] = op.RequestBody.Required
		contentMap := convertContentToOpenAPI3(op.RequestBody.Content)
		requestBody["content"] = contentMap
		// 对于文件上传接口，不向前端暴露 requestBody 节点，消除 Knife4j Vue 前端产生 in: "body" 并强制切换为 raw 的死锁
		if !isFileOperation {
			result["requestBody"] = requestBody
		}
	}

	// 针对 Knife4j 前端 UI 调试面板展平注入 formData 属性，确保前端能 100% 渲染出文件上传选择控件与输入框
	if isFileOperation && targetSchema != nil && len(targetSchema.Properties) > 0 {
		reqSet := make(map[string]bool)
		for _, reqField := range targetSchema.Required {
			reqSet[reqField] = true
		}

		formDataParams := make([]map[string]any, 0, len(targetSchema.Properties))
		for propName, propSchema := range targetSchema.Properties {
			pType := propSchema.Type
			if pType == "" {
				pType = ParamTypeString
			}
			pDesc := propSchema.Description
			pParser := NewCommentParser().Parse(pDesc)
			if pParser.HasTag(TagDescription) {
				pDesc = pParser.GetString(TagDescription)
			}
			pItem := map[string]any{
				"name":        propName,
				"in":          ParamInFormData,
				"type":        pType,
				"description": pDesc,
			}

			isFileField := pParser.HasTag(TagFile)

			if reqSet[propName] || isFileField {
				pItem["required"] = true
			}
			if isFileField {
				pItem["type"] = ParamTypeFile
			}
			formDataParams = append(formDataParams, pItem)
		}
		if len(formDataParams) > 0 {
			result["parameters"] = formDataParams
		}
	} else if len(op.Parameters) > 0 {
		params := make([]map[string]any, len(op.Parameters))
		for i, param := range op.Parameters {
			paramMap := map[string]any{
				"name":        param.Name,
				"in":          param.In,
				"description": param.Description,
				"required":    param.Required,
			}
			if param.Schema != nil {
				paramMap["schema"] = convertSchemaToOpenAPI3(param.Schema)
			}
			params[i] = paramMap
		}
		result["parameters"] = params
	}

	// 处理响应
	responses := make(map[string]any)
	for code, response := range op.Responses {
		responseMap := make(map[string]any)
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
func convertContentToOpenAPI3(content map[string]MediaType) map[string]any {
	result := make(map[string]any)
	for contentType, mediaType := range content {
		mediaTypeMap := make(map[string]any)
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
func convertSchemasToOpenAPI3(schemas map[string]Schema) map[string]any {
	result := make(map[string]any)
	for name, schema := range schemas {
		result[name] = convertSchemaToOpenAPI3(&schema)
	}
	return result
}

// convertSchemaToOpenAPI3 将 Schema 转换为 OpenAPI 3.0 格式
func convertSchemaToOpenAPI3(schema *Schema) map[string]any {
	if schema == nil {
		return nil
	}

	result := make(map[string]any)

	// 设置基本属性
	if schema.Type != "" {
		result["type"] = schema.Type
	}
	if schema.Format != "" {
		result["format"] = schema.Format
	}
	if schema.Title != "" {
		result["title"] = schema.Title
	}
	if schema.Default != nil {
		result["default"] = schema.Default
	}

	// 使用注释解析器处理描述
	parser := NewCommentParser().Parse(schema.Description)

	// OpenAPI 3.0 规范：若属性带有 @file 标注，强制转换为 type: "string", format: "binary"
	if parser.HasTag(TagFile) {
		result["type"] = ParamTypeString
		result["format"] = ParamFormatBinary
	}
	// 从解析器中获取标签值
	if parser.HasTag("description") {
		result["description"] = parser.GetString("description")
	}
	if parser.HasTag("example") {
		result["example"] = parser.GetString("example")
	}
	if parser.HasTag("format") {
		result["format"] = parser.GetString("format")
	}
	if parser.HasTag("enum") {
		result["enum"] = parser.GetArray("enum")
	}
	if parser.HasTag("minLength") {
		result["minLength"] = int64(parser.GetNumber("minLength"))
	}
	if parser.HasTag("maxLength") {
		result["maxLength"] = int64(parser.GetNumber("maxLength"))
	}
	if parser.HasTag("minimum") {
		result["minimum"] = parser.GetNumber("minimum")
	}
	if parser.HasTag("maximum") {
		result["maximum"] = parser.GetNumber("maximum")
	}
	if parser.HasTag("pattern") {
		result["pattern"] = strings.Trim(parser.GetString("pattern"), "\"")
	}

	// 处理其他属性
	if schema.MaxLength != nil {
		result["maxLength"] = schema.MaxLength
	}
	if schema.MinLength != nil {
		result["minLength"] = schema.MinLength
	}
	if schema.Pattern != "" {
		result["pattern"] = schema.Pattern
	}
	if schema.MaxItems != nil {
		result["maxItems"] = schema.MaxItems
	}
	if schema.MinItems != nil {
		result["minItems"] = schema.MinItems
	}
	result["uniqueItems"] = schema.UniqueItems
	if schema.MaxProperties != nil {
		result["maxProperties"] = schema.MaxProperties
	}
	if schema.MinProperties != nil {
		result["minProperties"] = schema.MinProperties
	}
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}
	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}
	if schema.Items != nil {
		result["items"] = convertSchemaToOpenAPI3(schema.Items)
	}
	if schema.AdditionalItems != nil {
		result["additionalItems"] = convertSchemaToOpenAPI3(schema.AdditionalItems)
	}
	if len(schema.AllOf) > 0 {
		allOf := make([]map[string]any, 0, len(schema.AllOf))
		for _, item := range schema.AllOf {
			allOf = append(allOf, convertSchemaToOpenAPI3(item))
		}
		result["allOf"] = allOf
	}
	if len(schema.OneOf) > 0 {
		oneOf := make([]map[string]any, 0, len(schema.OneOf))
		for _, item := range schema.OneOf {
			oneOf = append(oneOf, convertSchemaToOpenAPI3(item))
		}
		result["oneOf"] = oneOf
	}
	if len(schema.AnyOf) > 0 {
		anyOf := make([]map[string]any, 0, len(schema.AnyOf))
		for _, item := range schema.AnyOf {
			anyOf = append(anyOf, convertSchemaToOpenAPI3(item))
		}
		result["anyOf"] = anyOf
	}
	if schema.Not != nil {
		result["not"] = convertSchemaToOpenAPI3(schema.Not)
	}

	// 处理属性定义
	if schema.Properties != nil {
		properties := make(map[string]any)
		for name, prop := range schema.Properties {
			properties[name] = convertSchemaToOpenAPI3(prop)
		}
		result["properties"] = properties
	}
	if schema.AdditionalProperties != nil {
		if schema.AdditionalProperties.IsBool {
			result["additionalProperties"] = schema.AdditionalProperties.Allows
		} else if schema.AdditionalProperties.Schema != nil {
			result["additionalProperties"] = convertSchemaToOpenAPI3(schema.AdditionalProperties.Schema)
		}
	}

	// 处理引用
	if schema.Ref != "" {
		result["$ref"] = schema.Ref
	}

	// 设置其他属性
	result["nullable"] = schema.Nullable
	result["readOnly"] = schema.ReadOnly
	result["writeOnly"] = schema.WriteOnly
	result["deprecated"] = schema.Deprecated

	return result
}
