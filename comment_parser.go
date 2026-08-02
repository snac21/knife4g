package knife4g

import (
	"strconv"
	"strings"
)

// CommentParser 注释解析器，用于解析 Proto 注释以及 OpenAPI 描述信息中的自定义标注扩展指令（如 @tags, @summary, @description 等）
type CommentParser struct {
	tags         map[string]string   // 存储字符串类型的单值标签映射（如 key="tags", value="基础服务"）
	arrayTags    map[string][]string // 存储切片/数组类型的多值标签映射（如 key="tags", value=["基础服务", "通用服务"]）
	numberTags   map[string]float64  // 存储数值类型的标签映射（如 key="minLength", value=1）
	boolTags     map[string]bool     // 存储布尔类型的标签映射（如 key="required", value=true）
	responseTags map[string]string   // 存储 HTTP 响应状态码与响应类型的映射（如 key="400", value="ErrorResponse"）
}

// NewCommentParser 创建并初始化一个新的注释解析器实例
func NewCommentParser() *CommentParser {
	return &CommentParser{
		tags:         make(map[string]string),
		arrayTags:    make(map[string][]string),
		numberTags:   make(map[string]float64),
		boolTags:     make(map[string]bool),
		responseTags: make(map[string]string),
	}
}

// Parse 解析输入的多行注释字符串，按行提取包含 @ 前缀的标签及其参数
func (p *CommentParser) Parse(comment string) *CommentParser {
	if comment == "" {
		return p
	}

	// 按换行符分割多行文本
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理以 @ 开头的标签格式，如 "@tags: 基础服务"
		if strings.HasPrefix(line, "@") {
			parts := strings.SplitN(line[1:], ":", 2)
			if len(parts) != 2 {
				continue
			}

			tag := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// 根据不同的标签名称进行分类
			switch tag {
			case "consumes":
				p.boolTags[tag] = true
				p.tags[tag] = value

			case "tags":
				// 处理标签列表，支持单个标签以及按逗号分隔的多标签切片
				value = strings.TrimSpace(value)
				p.tags[tag] = value
				values := strings.Split(value, ",")
				for i, v := range values {
					values[i] = strings.TrimSpace(v)
				}
				p.arrayTags[tag] = values

			case "enum":
				// 处理枚举值列表，去除方括号并按逗号分割
				value = strings.Trim(value, "[]")
				values := strings.Split(value, ",")
				for i, v := range values {
					values[i] = strings.TrimSpace(v)
				}
				p.arrayTags[tag] = values

			case "minLength", "maxLength", "minimum", "maximum":
				// 处理数值类型的校验约束
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					p.numberTags[tag] = num
				}

			case "required":
				// 处理布尔类型的必填状态
				p.boolTags[tag] = value == "true"

			case "example":
				// 处理示例值，去除包裹的双引号
				value = strings.Trim(value, "\"")
				p.tags[tag] = value

			case "response":
				// 处理响应标签，标准格式如 "400: ErrorResponse"
				if strings.Contains(value, ":") {
					responseParts := strings.SplitN(value, ":", 2)
					if len(responseParts) == 2 {
						code := strings.TrimSpace(responseParts[0])
						responseType := strings.TrimSpace(responseParts[1])
						p.responseTags[code] = responseType
					}
				}

			default:
				// 其他未特殊处理的标签，统一作为字符串类型存储
				p.tags[tag] = value
			}
		} else {
			// 处理无 @ 前缀的普通描述文本，作为默认的 description
			if _, exists := p.tags["description"]; !exists {
				p.tags["description"] = line
			}
		}
	}

	return p
}

// GetString 获取指定标签名称对应的字符串类型值
func (p *CommentParser) GetString(tag string) string {
	return p.tags[tag]
}

// GetArray 获取指定标签名称对应的字符串数组/切片类型值
func (p *CommentParser) GetArray(tag string) []string {
	return p.arrayTags[tag]
}

// GetNumber 获取指定标签名称对应的数值类型值
func (p *CommentParser) GetNumber(tag string) float64 {
	return p.numberTags[tag]
}

// GetBool 获取指定标签名称对应的布尔类型值
func (p *CommentParser) GetBool(tag string) bool {
	return p.boolTags[tag]
}

// GetResponse 获取指定响应 HTTP 状态码对应的响应数据结构名称
func (p *CommentParser) GetResponse(code string) string {
	return p.responseTags[code]
}

// GetResponses 获取解析出的所有 HTTP 状态码与响应结构的映射
func (p *CommentParser) GetResponses() map[string]string {
	return p.responseTags
}

// HasTag 检查解析器中是否存在指定名称的标签（覆盖字符串、数组、数值、布尔与响应类型）
func (p *CommentParser) HasTag(tag string) bool {
	_, hasString := p.tags[tag]
	_, hasArray := p.arrayTags[tag]
	_, hasNumber := p.numberTags[tag]
	_, hasBool := p.boolTags[tag]
	_, hasResponse := p.responseTags[tag]
	return hasString || hasArray || hasNumber || hasBool || hasResponse
}

// ParseOperationDescription 解析 API 操作的核心描述信息并封装为 OperationDescription 对象
func (p *CommentParser) ParseOperationDescription(comment string) *OperationDescription {
	p.Parse(comment)
	return &OperationDescription{
		Summary:     p.GetString("summary"),
		Description: p.GetString("description"),
		Tags:        p.GetArray("tags"),
		OperationID: p.GetString("operationId"),
		Request:     p.GetString("request"),
		Responses:   p.GetResponses(),
	}
}
