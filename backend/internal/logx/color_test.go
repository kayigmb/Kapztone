package logx

import (
	"bytes"
	"testing"
)

func TestColorLevel(t *testing.T) {
	for in, want := range map[string]string{
		"time=x level=INFO msg=ok\n":           "time=x \x1b[32mlevel=INFO\x1b[0m msg=ok\n",
		"time=x level=WARN msg=4xx\n":          "time=x \x1b[33mlevel=WARN\x1b[0m msg=4xx\n",
		"time=x level=ERROR msg=5xx\n":         "time=x \x1b[31mlevel=ERROR\x1b[0m msg=5xx\n",
		"time=x level=DEBUG msg=d\n":           "time=x level=DEBUG msg=d\n",
		`{"level":"INFO","msg":"ok"}` + "\n":   "{\x1b[32m\"level\":\"INFO\"\x1b[0m,\"msg\":\"ok\"}\n",
		`{"level":"WARN","msg":"4xx"}` + "\n":  "{\x1b[33m\"level\":\"WARN\"\x1b[0m,\"msg\":\"4xx\"}\n",
		`{"level":"ERROR","msg":"5xx"}` + "\n": "{\x1b[31m\"level\":\"ERROR\"\x1b[0m,\"msg\":\"5xx\"}\n",
		`{"level":"DEBUG","msg":"d"}` + "\n":   `{"level":"DEBUG","msg":"d"}` + "\n",
	} {
		var buf bytes.Buffer
		n, err := colorLevel{&buf}.Write([]byte(in))
		if err != nil || n != len(in) || buf.String() != want {
			t.Errorf("%q: got %q n=%d err=%v, want %q", in, buf.String(), n, err, want)
		}
	}
}
