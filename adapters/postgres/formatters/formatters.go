package formatters

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// FormatArray format slice to a postgres array format
// today support a slice of string, int and fmt.Stringer
func FormatArray(value interface{}) string {
	var aux string
	var check = func(aux string, value interface{}) (ret string) {
		if aux != "" {
			aux += ","
		}
		ret = aux + FormatArray(value)
		return
	}
	switch value := value.(type) {
	case []fmt.Stringer:
		for _, v := range value {
			aux = check(aux, v)
		}
		return "{" + aux + "}"
	case []interface{}:
		for _, v := range value {
			aux = check(aux, v)
		}
		return "{" + aux + "}"
	case []string:
		for _, v := range value {
			aux = check(aux, v)
		}
		return "{" + aux + "}"
	case []int:
		for _, v := range value {
			aux = check(aux, v)
		}
		return "{" + aux + "}"
	case string:
		aux := value
		aux = strings.Replace(aux, `\`, `\\`, -1)
		aux = strings.Replace(aux, `"`, `\"`, -1)
		return `"` + aux + `"`
	case int:
		return strconv.Itoa(value)
	case fmt.Stringer:
		return FormatArray(value.String())
	}
	return ""
}

// FormatVector format slice to PostgreSQL vector literal
// supports []float64, []float32, and []interface{} containing numeric values
func FormatVector(value interface{}) string {
	switch v := value.(type) {
	case []float64:
		// Format as: [1.0,2.0,3.0]
		strValues := make([]string, len(v))
		for i, f := range v {
			strValues[i] = strconv.FormatFloat(f, 'f', -1, 64)
		}
		return "[" + strings.Join(strValues, ",") + "]"
	case []float32:
		// Format as: [1.0,2.0,3.0]
		strValues := make([]string, len(v))
		for i, f := range v {
			strValues[i] = strconv.FormatFloat(float64(f), 'f', -1, 32)
		}
		return "[" + strings.Join(strValues, ",") + "]"
	case []interface{}:
		// JSON decoder produces []interface{} for numeric arrays
		// Check if all elements are numeric types
		strValues := make([]string, len(v))
		for i, item := range v {
			switch num := item.(type) {
			case float64:
				strValues[i] = strconv.FormatFloat(num, 'f', -1, 64)
			case float32:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 32)
			case int:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case int64:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case int32:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case int16:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case int8:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case uint:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case uint64:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case uint32:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case uint16:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case uint8:
				strValues[i] = strconv.FormatFloat(float64(num), 'f', -1, 64)
			case json.Number:
				f, err := num.Float64()
				if err != nil {
					return ""
				}
				strValues[i] = strconv.FormatFloat(f, 'f', -1, 64)
			default:
				// Non-numeric element, not a valid vector
				return ""
			}
		}
		return "[" + strings.Join(strValues, ",") + "]"
	default:
		return ""
	}
}
