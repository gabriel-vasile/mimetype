package mimetype

import "testing"

func TestExtendNormalisesAliases(t *testing.T) {
	aliases := []string{" Application/X-Example-Alias; version=1 "}
	Extend(func([]byte, uint32) bool { return false }, "application/x-example", ".example", aliases...)
	defer func() { root.children = root.children[1:] }()
	m := Lookup("application/x-example")
	if !m.Is("application/x-example-alias") {
		t.Error("extended alias did not match its normalised media type")
	}
	if Lookup("APPLICATION/X-EXAMPLE-ALIAS; version=2") != m {
		t.Error("alias lookup did not ignore case and parameters")
	}
	if aliases[0] != " Application/X-Example-Alias; version=1 " {
		t.Error("Extend modified the caller's aliases")
	}
}
