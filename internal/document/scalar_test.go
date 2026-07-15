package document

import "testing"

func TestEmitScalar(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"retry-semantics.md", "retry-semantics.md"},
		{"Retry Semantics (v2)", "Retry Semantics (v2)"},
		{"github:openziti/zrok#412", "github:openziti/zrok#412"},
		{"#weird.md", `"#weird.md"`},
		{"watch: the # signs", `"watch: the # signs"`},
		{"trailing:", `"trailing:"`},
		{"- leading dash", `"- leading dash"`},
		{"true", `"true"`},
		{"2026", `"2026"`},
		{"2026-07-13", `"2026-07-13"`},
		{".inf", `".inf"`},
		{"12:34", `"12:34"`},
		{"y", `"y"`},
		{"", `""`},
		{" padded ", `" padded "`},
	}
	for _, tt := range tests {
		if got := emitScalar(tt.in); got != tt.want {
			t.Errorf("emitScalar(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestHash(t *testing.T) {
	if got := Hash(nil); got != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("Hash(empty) = %s", got)
	}
	if Hash([]byte("a")) == Hash([]byte("b")) {
		t.Error("distinct content must hash distinctly")
	}
}
