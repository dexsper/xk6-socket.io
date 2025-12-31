// Package socket.io contains the xk6-socket.io extension.
package socketio

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/grafana/sobek"
	"go.k6.io/k6/js/modules"
)

type rootModule struct{}

func (*rootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &module{vu}
}

type module struct {
	vu modules.VU
}

func (m *module) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]any{
			"io":  m.ConnectSocketIO,
		},
	}
}

var _ modules.Module = (*rootModule)(nil)

type Options struct {
	Path      string         `json:"path"`
	Namespace string         `json:"namespace"`
	Auth      map[string]any `json:"auth"`
	Query     map[string]any `json:"query"`
	Params    map[string]any `json:"params"`
}

func (a *API) ConnectSocketIO(host string, options map[string]any, handler sobek.Value) (sobek.Value, error) {
	runtime := a.vu.Runtime()

	var optionsMap Options
	if options != nil && !sobek.IsUndefined(options) && !sobek.IsNull(options) {
		if err := runtime.ExportTo(options, &optionsMap); err != nil {
			return nil, fmt.Errorf("invalid options: %w", err)
		}
	}

	if optionsMap.Path == "" { optionsMap.Path = "/socket.io/" }
	if optionsMap.Namespace == "" { optionsMap.Namespace = "/" }
	if optionsMap.Params == nil { optionsMap.Params = map[string]any{} }

	websocketURL, err := buildSocketIOWSURL(host, options)
	if err != nil { return nil, err }

	var handlerFunction sobek.Callable
	if handler != nil && !sobek.IsUndefined(handler) && !sobek.IsNull(handler) {
		_handlerFunction, ok := sobek.AssertFunction(handler)
		if !ok { panic(rt.ToValue("handler must be a function")) }
	
		if _, herr := hfn(sobek.Undefined(), socket); herr != nil { panic(herr) }
		handlerFunction = _handlerFunction
	}

	// this needs better naming
	requireFunction := requireMethod(rt, rt, "require")
	wsModule, err := requireFunction(sobek.Undefined(), rt.ToValue("k6/ws"))
	if err != nil { return nil, err }

	connectFunction := requireMethod(rt, wsModule, "connect")

	callback := rt.ToValue(func(callbackContext sobek.FunctionCall) sobek.Value {
		socketValue := callbackContext.Argument(0)
		socketObject := socketValue.ToObject(rt)

		onCallbackFunction := requireMethod(rt, socketObject, "on")
		sendFunction := requireMethod(rt, socketObject, "send")

		jsonObject := rt.Get("JSON").ToObject(rt)
		jsonStringifyFunction := requireMethod(rt, jsonObject, "stringify")

		emitFunction := func(event string, data sobek.Value) {
			array := rt.NewArray()
			_ = array.Push(rt.ToValue(event))
			_ = array.Push(data)

			stringifiedArray, err := jsonStringifyFunction(sobek.Undefined(), array)
			if err != nil { panic(err) }

			frame := "42" + stringifiedArray.String()

			if _, err := sendFunction(socketObject, rt.ToValue(frame)); err != nil {
				panic(err)
			}
		}

		msgHandlerFunction := func(msg string) {
			// Engine.IO ping -> pong
			if msg == "2" {
				_, _ := sendFunction(socketObject, rt.ToValue("3"))
			}

			if strings.HasPrefix(msg, "0") {
				_, _ := sendFunction(socketObject, rt.ToValue("40"))
			}
		}

		_, err := onCallbackFunction(socketObj, rt.ToValue("message"), func(messageHandlerContext sobek.FunctionCall) sobek.Value {
			msg := mc.Argument(0).String()
			msgHandlerFunction(msg)

			if userFn != nil {
				if _, err := userFn(sobek.Undefined(), socketVal); err != nil {
					panic(err)
				}
			}

			return sobek.Undefined()
		})
		if err != nil { panic(err) }

		socketObj.Set("emit", func(emitContext sobek.FunctionCall) sobek.Value {
			if len(emitContext.Arguments) == 0 { panic(rt.ToValue("emit(event, data): missing event")) }
			event := emitContext.Argument(0).String()

			if len(emitContext.Arguments > 1) { panic(rt.ToValue("emit(event, data): missing data")) }
			data := sobek.Argument(1)

			emitFunction(event, data)
			return sobek.Undefined()
		})

		socketObj.Set("send", func(sendContext sobek.FunctionCall) sobek.Value {
			if len(call.Arguments) == 0 { panic(rt.ToValue("send(data): missing data")) }
			data := sobek.Argument(0)

			emitInternal("message", data)
			return sobek.Undefined()
		})
		
		return sobek.Undefined()
	})

	return connectFn(
		sobek.Undefined(),
		rt.ToValue(wsURL),
		rt.ToValue(params),
		callback,
	)
}

func requireMethod(rt *sobek.Runtime, obj *sobek.Object, name string) sobek.Callable {
	v := obj.Get(name)
	function, ok := sobek.AssertFunction(v)

	if !ok {
		panic(rt.ToValue(fmt.Sprintf("method %q not found or not callable", name)))
	}
	return fn
}

func buildSocketIOWSURL(host string, opts Options) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return "", fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	path := opts.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u.Path = path

	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "websocket")
	for k, v := range opts.Query {
		q.Set(k, fmt.Sprint(v))
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}