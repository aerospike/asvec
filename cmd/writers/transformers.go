package writers

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/text"
)

var removeNil = text.Transformer(func(val interface{}) string {
	switch v := val.(type) {
	case *string:
		if v == nil {
			return ""
		}

		return *v
	default:
		return fmt.Sprintf("%v", v)
	}
})
