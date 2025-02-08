package magic

import "testing"

func TestGltf(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{
			name: "valid_gltf",
			raw:  `{"asset":{"version":"2.0"}}`,
			want: true,
		},
		{
			name: "not_json",
			raw:  "not json",
			want: false,
		},
		{
			name: "json_but_not_gltf",
			raw:  `{"foo":"bar"}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Gltf([]byte(tt.raw), 0)
			if got != tt.want {
				t.Errorf("Gltf() = %v, want %v", got, tt.want)
			}
		})
	}
}
