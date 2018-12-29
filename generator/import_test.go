package generator

import (
	gΩ "github.com/onsi/gomega"
	"testing"
)

func TestImport_String(t *testing.T) {
	var testcases = []struct {
		name     string
		imp      Import
		expected string
	}{
		{
			name:     "stdlib package",
			imp:      Import{Alias: "os", PkgPath: "os"},
			expected: `"os"`,
		},
		{
			name:     "alias matches base name",
			imp:      Import{Alias: "foo", PkgPath: "example.com/goo/foo"},
			expected: `"example.com/goo/foo"`,
		},
		{
			name:     "custom package alias",
			imp:      Import{Alias: "thinga", PkgPath: "example.com/go-thing"},
			expected: `thinga "example.com/go-thing"`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			o := gΩ.NewGomegaWithT(t)
			o.Expect(tc.imp.String()).To(gΩ.Equal(tc.expected))
		})
	}
}
