package mimetype_test

import (
	"encoding/json"
	"fmt"

	"github.com/gabriel-vasile/mimetype"
)

func ExampleLookup_configuration() {
	// Store the MIME type as a string in a JSON (or YAML) configuration.
	// Resolve it separately, so unsupported values can be reported explicitly.
	type config struct {
		MIME string `json:"mime"`
	}
	for _, input := range []string{
		`{"mime":"application/json"}`,
		`{"mime":"application/x-not-registered"}`,
	} {
		var cfg config
		if err := json.Unmarshal([]byte(input), &cfg); err != nil {
			fmt.Println("invalid configuration:", err)
			continue
		}
		m := mimetype.Lookup(cfg.MIME)
		if m == nil {
			fmt.Printf("unsupported MIME type: %q\n", cfg.MIME)
			continue
		}
		fmt.Println(m.String(), m.Extension())
		encoded, err := json.Marshal(cfg)
		if err != nil {
			fmt.Println("cannot encode configuration:", err)
			continue
		}
		fmt.Println(string(encoded))
	}
	// Output:
	// application/json .json
	// {"mime":"application/json"}
	// unsupported MIME type: "application/x-not-registered"
}
