package interpreter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/v4nsh0x/pengu/runtime"
)

func createHttpModule() *runtime.Value {
	om := runtime.NewOrderedMap()

	om.Set("get", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		return doHttpRequest("GET", args)
	}))
	om.Set("post", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		return doHttpRequest("POST", args)
	}))
	om.Set("put", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		return doHttpRequest("PUT", args)
	}))
	om.Set("delete", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		return doHttpRequest("DELETE", args)
	}))

	return runtime.NewObject(om)
}

func doHttpRequest(method string, args []*runtime.Value) (*runtime.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("http.%s expects at least a URL", method)
	}

	urlVal := args[0]
	if urlVal.Type != runtime.VAL_STRING {
		return nil, fmt.Errorf("URL must be a string")
	}
	url := urlVal.Str

	var reqBody io.Reader
	var options *runtime.Value

	if method == "POST" || method == "PUT" {
		if len(args) >= 2 && args[1].Type != runtime.VAL_NULL {
			dataVal := args[1]
			if dataVal.Type == runtime.VAL_STRING {
				reqBody = bytes.NewBuffer([]byte(dataVal.Str))
			} else {
				// Convert object/array to JSON
				jsonData, err := json.Marshal(valueToInterface(dataVal))
				if err != nil {
					return nil, fmt.Errorf("invalid data for JSON body: %v", err)
				}
				reqBody = bytes.NewBuffer(jsonData)
			}
		}
		if len(args) >= 3 {
			options = args[2]
		}
	} else {
		if len(args) >= 2 {
			options = args[1]
		}
	}

	timeout := 10 * time.Second
	headers := make(map[string]string)

	if options != nil && options.Type == runtime.VAL_OBJECT {
		if t, ok := options.Object.Get("timeout"); ok && t.Type == runtime.VAL_NUMBER {
			timeout = time.Duration(t.Number) * time.Second
		}
		if h, ok := options.Object.Get("headers"); ok && h.Type == runtime.VAL_OBJECT {
			for _, key := range h.Object.Keys {
				if val, exists := h.Object.Get(key); exists && val.Type == runtime.VAL_STRING {
					headers[key] = val.Str
				}
			}
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Default content type for object data
	if reqBody != nil && (args[1].Type == runtime.VAL_OBJECT || args[1].Type == runtime.VAL_ARRAY) {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Build response object
	respObj := runtime.NewOrderedMap()
	respObj.Set("status", runtime.NewNumber(float64(resp.StatusCode), true))
	respObj.Set("text", runtime.NewString(string(bodyBytes)))

	// Build headers object
	respHeaders := runtime.NewOrderedMap()
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders.Set(k, runtime.NewString(v[0])) // Just take the first header value for simplicity
		}
	}
	respObj.Set("headers", runtime.NewObject(respHeaders))

	// JSON method
	respObj.Set("json", runtime.NewBuiltin(func(jsonArgs []*runtime.Value) (*runtime.Value, error) {
		var result interface{}
		err := json.Unmarshal(bodyBytes, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON response: %v", err)
		}
		return interfaceToValue(result), nil
	}))

	return runtime.NewObject(respObj), nil
}

func valueToInterface(v *runtime.Value) interface{} {
	switch v.Type {
	case runtime.VAL_NUMBER:
		return v.Number
	case runtime.VAL_STRING:
		return v.Str
	case runtime.VAL_BOOL:
		return v.Bool
	case runtime.VAL_NULL:
		return nil
	case runtime.VAL_ARRAY:
		arr := make([]interface{}, len(v.Array))
		for i, item := range v.Array {
			arr[i] = valueToInterface(item)
		}
		return arr
	case runtime.VAL_OBJECT:
		obj := make(map[string]interface{})
		for _, key := range v.Object.Keys {
			if val, ok := v.Object.Get(key); ok {
				obj[key] = valueToInterface(val)
			}
		}
		return obj
	default:
		return nil
	}
}

func interfaceToValue(v interface{}) *runtime.Value {
	switch val := v.(type) {
	case float64:
		// simple heuristic for int check
		isInt := val == float64(int64(val))
		return runtime.NewNumber(val, isInt)
	case string:
		return runtime.NewString(val)
	case bool:
		return runtime.NewBool(val)
	case nil:
		return runtime.NewNull()
	case []interface{}:
		arr := make([]*runtime.Value, len(val))
		for i, item := range val {
			arr[i] = interfaceToValue(item)
		}
		return runtime.NewArray(arr)
	case map[string]interface{}:
		om := runtime.NewOrderedMap()
		for k, item := range val {
			om.Set(k, interfaceToValue(item))
		}
		return runtime.NewObject(om)
	default:
		return runtime.NewNull()
	}
}
