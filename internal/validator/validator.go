package validator

import "regexp"

var (
	EmailRegExp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, val string) {
	if _, exist := v.Errors[key]; !exist {
		v.Errors[key] = val
	}
}

func (v *Validator) Check(ok bool, key, val string) {
	if ok {
		return
	}
	v.AddError(key, val)
}

func (v *Validator) In(val string, list ...string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

func Matches(val string, rx *regexp.Regexp) bool {
	return rx.MatchString(val)
}

func Unique(values ...string) bool {
	valuesMap := make(map[string]bool)
	for _, v := range values {
		valuesMap[v] = true
	}

	return len(values) == len(valuesMap)
}

func GreaterThan(x, y int) bool {
	return x > y
}

func Unique2(vals ...string) bool {
	valsMap := make(map[string]bool)

	for _, val := range vals {
		if _, exist := valsMap[val]; !exist {
			valsMap[val] = true
		} else {
			return true
		}
	}

	return false
}
