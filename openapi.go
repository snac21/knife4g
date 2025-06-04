package knife4g

// OpenAPI3 表示 OpenAPI 3.0 规范的结构
type OpenAPI3 struct {
	OpenAPI    string              `json:"openapi" yaml:"openapi"`
	Info       Info                `json:"info" yaml:"info"`
	Paths      map[string]PathItem `json:"paths" yaml:"paths"`
	Components Components          `json:"components" yaml:"components"`
	Tags       []Tag               `json:"tags" yaml:"tags"`
	Servers    []Server            `json:"servers" yaml:"servers"`
}

// Info 包含 API 的基本信息
type Info struct {
	Title          string   `json:"title" yaml:"title"`
	Description    string   `json:"description" yaml:"description"`
	Version        string   `json:"version" yaml:"version"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Name           string   `json:"name,omitempty" yaml:"name,omitempty"`
}

// Contact 包含联系信息
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// Server 表示服务器信息
type Server struct {
	URL         string                    `json:"url" yaml:"url"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// ServerVariable 表示服务器变量
type ServerVariable struct {
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
}

// PathItem 表示路径项
type PathItem struct {
	Ref         string      `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *Operation  `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *Operation  `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *Operation  `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *Operation  `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch       *Operation  `json:"patch,omitempty" yaml:"patch,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Operation 表示 API 操作
type Operation struct {
	Tags        []string              `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses" yaml:"responses"`
	Callbacks   map[string]Callback   `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security    []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Servers     []Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// Parameter 表示参数
type Parameter struct {
	Name            string               `json:"name" yaml:"name"`
	In              string               `json:"in" yaml:"in"`
	Description     string               `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                 `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool                 `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema              `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         interface{}          `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]Example   `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// RequestBody 表示请求体
type RequestBody struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content" yaml:"content"`
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
}

// Response 表示响应
type Response struct {
	Description string               `json:"description" yaml:"description"`
	Headers     map[string]Header    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Links       map[string]Link      `json:"links,omitempty" yaml:"links,omitempty"`
}

// Components 表示组件
type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       map[string]Response       `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]Parameter      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]Example        `json:"examples,omitempty" yaml:"examples,omitempty"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         map[string]Header         `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Links           map[string]Link           `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       map[string]Callback       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
}

// Schema 表示模式
type Schema struct {
	Type                 string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Format               string                 `json:"format,omitempty" yaml:"format,omitempty"`
	Title                string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Description          string                 `json:"description,omitempty" yaml:"description,omitempty"`
	FieldDoc             *FieldDoc              `json:"fieldDoc,omitempty" yaml:"fieldDoc,omitempty"`
	Default              interface{}            `json:"default,omitempty" yaml:"default,omitempty"`
	MultipleOf           *float64               `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum     bool                   `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum     bool                   `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength            *int                   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength            *int                   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems             *int                   `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems             *int                   `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems          bool                   `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties        *int                   `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        *int                   `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required             []string               `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty" yaml:"enum,omitempty"`
	Properties           map[string]*Schema     `json:"properties,omitempty" yaml:"properties,omitempty"`
	AllOf                []*Schema              `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                []*Schema              `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []*Schema              `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not                  *Schema                `json:"not,omitempty" yaml:"not,omitempty"`
	Items                *Schema                `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalItems      *Schema                `json:"additionalItems,omitempty" yaml:"additionalItems,omitempty"`
	AdditionalProperties *Schema                `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Ref                  string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Nullable             bool                   `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Discriminator        *Discriminator         `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	ReadOnly             bool                   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool                   `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	XML                  *XML                   `json:"xml,omitempty" yaml:"xml,omitempty"`
	ExternalDocs         *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Example              interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
	Deprecated           bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// MediaType 表示媒体类型
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// Example 表示示例
type Example struct {
	Summary       string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value         interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Header 表示头部
type Header struct {
	Description     string               `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                 `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool                 `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema              `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         interface{}          `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]Example   `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// Tag 表示标签
type Tag struct {
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// ExternalDocumentation 表示外部文档
type ExternalDocumentation struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
}

// SecurityRequirement 表示安全要求
type SecurityRequirement map[string][]string

// SecurityScheme 表示安全方案
type SecurityScheme struct {
	Type             string      `json:"type" yaml:"type"`
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
}

// OAuthFlows 表示 OAuth 流程
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow 表示 OAuth 流程
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"`
}

// Link 表示链接
type Link struct {
	OperationRef string                 `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	OperationID  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Server       *Server                `json:"server,omitempty" yaml:"server,omitempty"`
}

// Callback 表示回调
type Callback map[string]PathItem

// XML 表示 XML
type XML struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty" yaml:"wrapped,omitempty"`
}

// Discriminator 表示鉴别器
type Discriminator struct {
	PropertyName string            `json:"propertyName" yaml:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

// Encoding 表示编码
type Encoding struct {
	ContentType   string            `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string            `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool              `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool              `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// FieldDoc 表示字段文档信息
type FieldDoc struct {
	FieldDescription string `json:"fieldDescription"` // @description 后的内容
	FieldExample     string `json:"fieldExample"`     // @example 后的内容
	FieldFormat      string `json:"fieldFormat"`      // @format 后的内容
	FieldRequired    bool   `json:"fieldRequired"`    // @required 后的内容
	FieldMinLength   int    `json:"fieldMinLength"`   // @minLength 后的内容
	FieldMaxLength   int    `json:"fieldMaxLength"`   // @maxLength 后的内容
	FieldMinimum     int    `json:"fieldMinimum"`     // @minimum 后的内容
	FieldMaximum     int    `json:"fieldMaximum"`     // @maximum 后的内容
	FieldPattern     string `json:"fieldPattern"`     // @pattern 后的内容
	FieldEnum        []int  `json:"fieldEnum"`        // @enum 后的内容
	Raw              string `json:"raw"`              // 原始描述
}

// OperationDescription 表示操作描述信息
type OperationDescription struct {
	Summary     string            `json:"summary,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	OperationID string            `json:"operationId,omitempty"`
	Request     string            `json:"request,omitempty"`
	Responses   map[string]string `json:"responses,omitempty"`
}
