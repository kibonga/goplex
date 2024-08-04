package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonVal := fmt.Sprintf("%d mins", r)

	quotedJsonVal := strconv.Quote(jsonVal)

	return []byte(quotedJsonVal), nil
}

func (r *Runtime) MarshalJSONPtr() ([]byte, error) {
	jsonVal := fmt.Sprintf("%d mins", r)

	quotedJsonVal := strconv.Quote(jsonVal)

	return []byte(quotedJsonVal), nil
}

func (r *Runtime) UnmarshalJSON(jsonVal []byte) error {
	fmt.Printf("unmarshal json runtime = %d", r)
	fmt.Printf("unmarshal json val = %v", jsonVal)

	unquotedJsonVal, err := strconv.Unquote(string(jsonVal))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	values := strings.Split(unquotedJsonVal, " ")
	if len(values) != 2 {
		return ErrInvalidRuntimeFormat
	}
	num, err := strconv.Atoi(values[0])
	mins := values[1]
	if err != nil || mins != "mins" {
		return ErrInvalidRuntimeFormat
	}

	var runtime Runtime = Runtime(num)
	*r = runtime

	return nil
}
