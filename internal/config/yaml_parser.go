package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type yamlToken struct {
	indent int
	text   string
	line   int
}

type yamlParser struct {
	tokens []yamlToken
	index  int
}

func decodeYAML(data []byte, out interface{}) error {
	root, err := parseYAML(data)
	if err != nil {
		return err
	}
	buf, err := json.Marshal(root)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, out)
}

func parseYAML(data []byte) (interface{}, error) {
	tokens, err := tokenizeYAML(string(data))
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, nil
	}
	parser := &yamlParser{tokens: tokens}
	node, err := parser.parseNode(0)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func tokenizeYAML(input string) ([]yamlToken, error) {
	var tokens []yamlToken
	lines := strings.Split(input, "\n")
	for idx, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := countIndent(line)
		if indent%2 != 0 {
			return nil, fmt.Errorf("line %d: indentation must be multiples of two spaces", idx+1)
		}
		tokens = append(tokens, yamlToken{
			indent: indent,
			text:   strings.TrimSpace(line),
			line:   idx + 1,
		})
	}
	return tokens, nil
}

func countIndent(line string) int {
	count := 0
	for _, r := range line {
		if r == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

func (p *yamlParser) parseNode(indent int) (interface{}, error) {
	if p.index >= len(p.tokens) {
		return nil, nil
	}
	tok := p.tokens[p.index]
	if tok.indent < indent {
		return nil, nil
	}
	if strings.HasPrefix(tok.text, "- ") {
		return p.parseSequence(indent)
	}
	return p.parseMap(indent)
}

func (p *yamlParser) parseMap(indent int) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for p.index < len(p.tokens) {
		tok := p.tokens[p.index]
		if tok.indent < indent {
			break
		}
		if tok.indent > indent {
			return nil, fmt.Errorf("line %d: unexpected indentation", tok.line)
		}
		if strings.HasPrefix(tok.text, "- ") {
			break
		}
		key, value, hasValue := splitKeyValue(tok.text)
		if key == "" {
			return nil, fmt.Errorf("line %d: expected key", tok.line)
		}
		p.index++
		if hasValue {
			if value == "" {
				val, err := p.parseNode(indent + 2)
				if err != nil {
					return nil, err
				}
				result[key] = val
			} else {
				result[key] = parseScalar(value)
			}
		} else {
			val, err := p.parseNode(indent + 2)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
	}
	return result, nil
}

func (p *yamlParser) parseSequence(indent int) ([]interface{}, error) {
	var seq []interface{}
	for p.index < len(p.tokens) {
		tok := p.tokens[p.index]
		if tok.indent < indent || !strings.HasPrefix(tok.text, "- ") {
			break
		}
		line := strings.TrimSpace(tok.text[2:])
		p.index++
		if line == "" {
			val, err := p.parseNode(indent + 2)
			if err != nil {
				return nil, err
			}
			seq = append(seq, val)
			continue
		}
		if key, val, hasValue := splitKeyValue(line); hasValue {
			item := make(map[string]interface{})
			if val == "" {
				sub, err := p.parseNode(indent + 2)
				if err != nil {
					return nil, err
				}
				item[key] = sub
			} else {
				item[key] = parseScalar(val)
			}
			extra, err := p.parseMap(indent + 2)
			if err != nil {
				return nil, err
			}
			for k, v := range extra {
				item[k] = v
			}
			seq = append(seq, item)
		} else {
			seq = append(seq, parseScalar(line))
		}
	}
	return seq, nil
}

func splitKeyValue(line string) (string, string, bool) {
	colon := strings.Index(line, ":")
	if colon == -1 {
		key := strings.Trim(line, ` "'`)
		return key, "", false
	}
	key := strings.TrimSpace(line[:colon])
	key = strings.Trim(key, `"'`)
	value := strings.TrimSpace(line[colon+1:])
	return key, value, true
}

func parseScalar(value string) interface{} {
	lower := strings.ToLower(value)
	switch lower {
	case "true":
		return true
	case "false":
		return false
	}
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) && len(value) >= 2 {
		return strings.Trim(value, `"`)
	}
	if strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`) && len(value) >= 2 {
		return strings.Trim(value, `'`)
	}
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return value
}
