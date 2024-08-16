package writers

import (
	"fmt"
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
