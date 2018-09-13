package generator

import (
	"log"
	"strings"
)

type Returns []Return

type Return struct {
	Name string
	Type string
}

func (r Returns) HasLength() bool {
	return len(r) > 0
}

func (r Returns) WithPrefix(p string) string {
	if len(r) == 0 {
		return ""
	}

	rets := []string{}
	for i := range r {
		if p == "" {
			rets = append(rets, unexport(r[i].Name))
		} else {
			rets = append(rets, p+unexport(r[i].Name))
		}
	}
	return strings.Join(rets, ", ")
}

func (r Returns) AsArgs() string {
	if len(r) == 0 {
		return ""
	}

	rets := []string{}
	for i := range r {
		log.Println(r[i].Type)
		rets = append(rets, r[i].Type)
	}
	return strings.Join(rets, ", ")
}

func (r Returns) AsNamedArgsWithTypes() string {
	if len(r) == 0 {
		return ""
	}

	rets := []string{}
	for i := range r {
		rets = append(rets, unexport(r[i].Name)+" "+r[i].Type)
	}
	return strings.Join(rets, ", ")
}

func (r Returns) AsNamedArgs() string {
	if len(r) == 0 {
		return ""
	}

	rets := []string{}
	for i := range r {
		rets = append(rets, unexport(r[i].Name))
	}
	return strings.Join(rets, ", ")
}

func (r Returns) AsReturnSignature() string {
	if len(r) == 0 {
		return ""
	}
	if len(r) == 1 {
		return r[0].Type
	}
	result := "("
	for i := range r {
		result = result + r[i].Type
		if i < len(r) {
			result = result + ", "
		}
	}
	result = result + ")"
	return result
}
