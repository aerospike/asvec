package writers

import (
	"fmt"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
)

var removeNil = text.Transformer(func(val interface{}) string {
	switch v := val.(type) {
	case *string:
		if v == nil {
			return ""
		}

		return *v
	case *time.Time:
		if v == nil {
			return ""
		}

		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
})

func RemoveTrailingZeros(s string) string {
	lastNoneZeroIdx := len(s) - 1

	for ; s[lastNoneZeroIdx] == '0'; lastNoneZeroIdx-- {
		if lastNoneZeroIdx-1 >= 0 && s[lastNoneZeroIdx-1] == '.' {
			break
		}
	}

	return s[:lastNoneZeroIdx+1]
}

var vectorFormat = text.Transformer(func(val interface{}) string {

	switch v := val.(type) {
	case []float32:
		ss := make([]string, len(v))

		for i, f := range v {
			// Can't use strconv.FormatFloat(f, 'f', -1, 32). It will turn 1.0
			// into 1. We need 1.0 incase the user copies it to do another query.
			ss[i] = RemoveTrailingZeros(fmt.Sprintf("%f", f))
		}

		return fmt.Sprintf("[%s]", strings.Join(ss, ","))
	case []bool:
		ss := make([]string, len(v))

		for i, f := range v {
			if f {
				ss[i] = "1"
			} else {
				ss[i] = "0"
			}
		}

		return fmt.Sprintf("[%s]", strings.Join(ss, ","))
	default:
		return fmt.Sprintf("%v", v)
	}
})
