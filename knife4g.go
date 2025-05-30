package knife4g

import (
	"embed"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	//go:embed front
	front embed.FS
)

type Config struct {
	RelativePath   string // 访问前缀，如 "/doc"
	OpenApiContent []byte // openapi.yaml的数据内容
	StaticPath     string // 静态资源路径（可选，默认 embed 的 front）
	KService       Knife4gService
}

type Knife4gService struct {
	Name           string `json:"name"`
	Url            string `json:"url"`
	SwaggerVersion string `json:"swaggerVersion"`
	Location       string `json:"location"`
}

// OpenAPI文档结构
type OpenAPI struct {
	OpenAPI    string                 `yaml:"openapi" json:"swagger"`
	Info       map[string]interface{} `yaml:"info" json:"info"`
	Paths      map[string]interface{} `yaml:"paths" json:"paths"`
	Components map[string]interface{} `yaml:"components" json:"definitions"`
	Tags       []struct {
		Name        string `yaml:"name" json:"name"`
		Description string `yaml:"description" json:"description"`
	} `yaml:"tags" json:"tags"`
}

// Handler 返回 knife4g 文档服务 http.Handler
func Handler(config *Config) http.Handler {
	docYamlPath := config.RelativePath + "/docYaml"
	servicesPath := config.RelativePath + "/front/service"
	docPath := config.RelativePath + "/index"
	appjsPath := config.RelativePath + "/front/webjars/js/app.42aa019b.js"

	config.KService.Url = "/docYaml"
	config.KService.Location = "/docYaml"
	config.KService.Name = "API Documentation"
	config.KService.SwaggerVersion = "2.0"

	appjsTemplate, err := template.New("app.42aa019b.js").
		Delims("{[(", ")]}").
		ParseFS(front, "front/webjars/js/app.42aa019b.js")
	if err != nil {
		log.Println(err)
	}
	docTemplate, err := template.New("doc.html").
		Delims("{[(", ")]}").
		ParseFS(front, "front/doc.html")
	if err != nil {
		log.Println(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Printf("Received request: %s", r.URL.Path)

		// 特殊处理 service 请求
		if strings.HasSuffix(r.URL.Path, "/front/service") {
			log.Printf("Handling service request: %s", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Knife4gService{config.KService})
			return
		}

		switch r.URL.Path {
		case appjsPath:
			log.Printf("Handling appjsPath: %s", r.URL.Path)
			err := appjsTemplate.Execute(w, config)
			if err != nil {
				log.Printf("Failed to execute appjs template: %v", err)
			}
		case servicesPath:
			log.Printf("Handling servicesPath: %s", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Knife4gService{config.KService})
		case docPath:
			log.Printf("Rendering doc template with RelativePath: %s", config.RelativePath)
			err := docTemplate.Execute(w, config)
			if err != nil {
				log.Printf("Failed to execute doc template: %v", err)
			}
		case docYamlPath:
			log.Printf("Handling docYamlPath: %s", r.URL.Path)
			// 解析OpenAPI文档
			var openapi OpenAPI
			if err := yaml.Unmarshal(config.OpenApiContent, &openapi); err != nil {
				log.Printf("Failed to parse OpenAPI document: %v\nContent: %s", err, string(config.OpenApiContent))
				http.Error(w, "Failed to parse OpenAPI document", http.StatusInternalServerError)
				return
			}

			// 转换为Swagger格式
			swagger := convertToSwagger(&openapi)

			// 返回JSON格式的Swagger文档
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(swagger); err != nil {
				log.Printf("Failed to encode Swagger document: %v", err)
				http.Error(w, "Failed to encode Swagger document", http.StatusInternalServerError)
				return
			}
		default:
			// 尝试作为静态文件处理
			if strings.HasPrefix(r.URL.Path, "/doc/front/") {
				path := strings.TrimPrefix(r.URL.Path, "/doc")
				log.Printf("Trying to open file as static: %s", path)

				// 直接使用path，因为path已经包含了/front前缀
				file, err := front.Open(strings.TrimPrefix(path, "/"))
				if err != nil {
					log.Printf("Failed to open static file: %v, path: %s", err, strings.TrimPrefix(path, "/"))
					http.NotFound(w, r)
					return
				}
				defer file.Close()

				// 设置正确的Content-Type
				ext := filepath.Ext(path)
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

				// 设置缓存控制
				w.Header().Set("Cache-Control", "public, max-age=31536000")

				io.Copy(w, file)
				return
			}
			log.Printf("Not found: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	})
}

// 将OpenAPI转换为Swagger格式
func convertToSwagger(openapi *OpenAPI) map[string]interface{} {
	swagger := make(map[string]interface{})
	swagger["openapi"] = openapi.OpenAPI
	swagger["info"] = openapi.Info
	swagger["tags"] = openapi.Tags
	swagger["servers"] = []map[string]interface{}{
		{
			"url":         "http://localhost:8000",
			"description": "Inferred Url",
		},
	}

	components := make(map[string]interface{})
	var schemas map[string]interface{}
	if s, ok := openapi.Components["schemas"].(map[string]interface{}); ok {
		schemas = s
		components["schemas"] = schemas
	}
	swagger["components"] = components

	paths := make(map[string]interface{})
	for path, item := range openapi.Paths {
		if pathItem, ok := item.(map[string]interface{}); ok {
			newPathItem := make(map[string]interface{})
			for method, op := range pathItem {
				if opMap, ok := op.(map[string]interface{}); ok {
					// 处理 requestBody
					if requestBody, ok := opMap["requestBody"].(map[string]interface{}); ok {
						if content, ok := requestBody["content"].(map[string]interface{}); ok {
							for _, cval := range content {
								if cMap, ok := cval.(map[string]interface{}); ok {
									if schema, ok := cMap["schema"].(map[string]interface{}); ok {
										if ref, hasRef := schema["$ref"].(string); hasRef && schemas != nil {
											refName := ref[strings.LastIndex(ref, "/")+1:]
											if model, ok := schemas[refName].(map[string]interface{}); ok {
												param := buildReqParameterTree(
													strings.ToLower(refName), // 参数名称
													refName,                  // 参数说明
													refName,                  // type
													refName,                  // schemaValue
													true,                     // require
													"body",                   // in
													model, schemas,
												)
												opMap["reqParameters"] = []map[string]interface{}{param}
											}
										}
									}
								}
							}
						}
					}
					newPathItem[method] = opMap
				} else {
					newPathItem[method] = op
				}
			}
			paths[path] = newPathItem
		} else {
			paths[path] = item
		}
	}
	swagger["paths"] = paths

	return swagger
}

func buildReqParameterTree(
	name, description, schemaType, schemaValue string,
	required bool, in string, model map[string]interface{}, schemas map[string]interface{},
) map[string]interface{} {
	param := map[string]interface{}{
		"name":        name,
		"description": description,
		"type":        schemaType,
		"schemaValue": schemaValue,
		"in":          in,
		"require":     required,
	}
	// 递归 children
	children := []map[string]interface{}{}
	if props, ok := model["properties"].(map[string]interface{}); ok {
		requiredFields := map[string]bool{}
		if reqs, ok := model["required"].([]interface{}); ok {
			for _, r := range reqs {
				if s, ok := r.(string); ok {
					requiredFields[s] = true
				}
			}
		}
		for propName, prop := range props {
			if propMap, ok := prop.(map[string]interface{}); ok {
				childType := ""
				if t, ok := propMap["type"].(string); ok {
					childType = t
				}
				childDesc := ""
				if d, ok := propMap["description"].(string); ok {
					childDesc = d
				}
				childSchemaValue := ""
				if ref, ok := propMap["$ref"].(string); ok {
					childSchemaValue = ref[strings.LastIndex(ref, "/")+1:]
				}
				childModel := map[string]interface{}{}
				if childSchemaValue != "" && schemas != nil {
					if m, ok := schemas[childSchemaValue].(map[string]interface{}); ok {
						childModel = m
					}
				}
				child := buildReqParameterTree(
					propName, childDesc, childType, childSchemaValue,
					requiredFields[propName], "", childModel, schemas,
				)
				children = append(children, child)
			}
		}
	}
	param["children"] = children
	return param
}
