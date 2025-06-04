package knife4g

import (
	"strconv"
	"strings"
)

// CommentParser 注释解析器
type CommentParser struct {
	tags         map[string]string
	arrayTags    map[string][]string
	numberTags   map[string]float64
	boolTags     map[string]bool
	responseTags map[string]string
}

// NewCommentParser 创建新的注释解析器
func NewCommentParser() *CommentParser {
	return &CommentParser{
		tags:         make(map[string]string),
		arrayTags:    make(map[string][]string),
		numberTags:   make(map[string]float64),
		boolTags:     make(map[string]bool),
		responseTags: make(map[string]string),
	}
}

// Parse 解析注释字符串
func (p *CommentParser) Parse(comment string) *CommentParser {
	if comment == "" {
		return p
	}

	// 按行分割
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理标签
		if strings.HasPrefix(line, "@") {
			parts := strings.SplitN(line[1:], ":", 2)
			if len(parts) != 2 {
				continue
			}

			tag := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// 根据标签类型进行不同的处理
			switch tag {
			case "enum":
				// 处理枚举值
				value = strings.Trim(value, "[]")
				values := strings.Split(value, ",")
				for i, v := range values {
					values[i] = strings.TrimSpace(v)
				}
				p.arrayTags[tag] = values

			case "minLength", "maxLength", "minimum", "maximum":
				// 处理数值类型
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					p.numberTags[tag] = num
				}

			case "required":
				// 处理布尔类型
				p.boolTags[tag] = value == "true"

			case "example":
				// 处理示例值，去除多余的引号
				value = strings.Trim(value, "\"")
				p.tags[tag] = value

			case "response":
				// 处理响应标签
				// 格式: "400: ErrorResponse"
				if strings.Contains(value, ":") {
					responseParts := strings.SplitN(value, ":", 2)
					if len(responseParts) == 2 {
						code := strings.TrimSpace(responseParts[0])
						responseType := strings.TrimSpace(responseParts[1])
						p.responseTags[code] = responseType
					}
				}

			default:
				// 处理字符串类型
				p.tags[tag] = value
			}
		} else {
			// 处理普通描述文本
			if _, exists := p.tags["description"]; !exists {
				p.tags["description"] = line
			}
		}
	}

	return p
}

// GetString 获取字符串类型的标签值
func (p *CommentParser) GetString(tag string) string {
	return p.tags[tag]
}

// GetArray 获取数组类型的标签值
func (p *CommentParser) GetArray(tag string) []string {
	return p.arrayTags[tag]
}

// GetNumber 获取数值类型的标签值
func (p *CommentParser) GetNumber(tag string) float64 {
	return p.numberTags[tag]
}

// GetBool 获取布尔类型的标签值
func (p *CommentParser) GetBool(tag string) bool {
	return p.boolTags[tag]
}

// GetResponse 获取响应标签值
func (p *CommentParser) GetResponse(code string) string {
	return p.responseTags[code]
}

// GetResponses 获取所有响应标签
func (p *CommentParser) GetResponses() map[string]string {
	return p.responseTags
}

// HasTag 检查是否存在指定标签
func (p *CommentParser) HasTag(tag string) bool {
	_, hasString := p.tags[tag]
	_, hasArray := p.arrayTags[tag]
	_, hasNumber := p.numberTags[tag]
	_, hasBool := p.boolTags[tag]
	_, hasResponse := p.responseTags[tag]
	return hasString || hasArray || hasNumber || hasBool || hasResponse
}

// ParseOperationDescription 解析操作描述信息
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
