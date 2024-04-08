package templates

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"html/template"
	"strings"

	"github.com/l1f/mahlzeit/web/assets/icons"
	"github.com/google/uuid"
	"github.com/speps/go-hashids/v2"
)

// Since we only use it for generating unique form IDs, we can use the default implementation as global state.
var hid, _ = hashids.New()

var appFunctions = template.FuncMap{
	"icon": func(iconName string) (template.HTML, error) {
		if !strings.HasSuffix(iconName, ".svg") {
			iconName += ".svg"
		}

		content, err := icons.Icons.ReadFile(iconName)
		if err != nil {
			return "", fmt.Errorf("reading %q from icons: %w", iconName, err)
		}

		return template.HTML(content), nil //nolint:gosec // all icons are predefined by us
	},
	"random": func() string {
		return uuid.New().String()
	},
	// formID generates a unique ID, based on the parameters as a hash.
	"formID": func(parameters ...any) (string, error) {
		h := fnv.New32a()

		for _, p := range parameters {
			switch v := p.(type) {
			case string:
				_, _ = h.Write([]byte(v))
			case int:
				_, _ = h.Write([]byte{byte(v)})
			case nil:
				continue
			default:
				marshalled, _ := json.Marshal(v)
				_, _ = h.Write(marshalled)
			}
		}

		s, err := hid.Encode([]int{int(h.Sum32())})
		if err != nil {
			return "", fmt.Errorf("generating hash failed: %w", err)
		}

		return "id-" + s, nil
	},
}
