package templates

import (
	"encoding/json"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/speps/go-hashids/v2"
)

func IntToStr(input int) string {
	return strconv.FormatInt(int64(input), 10)
}

func DateToStr(input time.Time) string {
	return input.Format("02.01.2006")
}

// Since we only use it for generating unique form IDs, we can use the default implementation as global state.
var hid, _ = hashids.New()

func FormId(parameters ...any) string {
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
		return ""
	}

	return "id-" + s
}
