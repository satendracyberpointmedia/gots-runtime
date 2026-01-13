package plugin

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
)

// TypeScriptPlugin wraps a TypeScript plugin implementation
type TypeScriptPlugin struct {
	name      string
	version   string
	initFunc  goja.Callable
	execFunc  goja.Callable
	shutdownFunc goja.Callable
	engine    *goja.Runtime
	mu        sync.RWMutex
}

// NewTypeScriptPlugin creates a new TypeScript plugin
func NewTypeScriptPlugin(engine *goja.Runtime, name, version string, initFunc, execFunc, shutdownFunc goja.Callable) *TypeScriptPlugin {
	return &TypeScriptPlugin{
		name:        name,
		version:     version,
		initFunc:    initFunc,
		execFunc:    execFunc,
		shutdownFunc: shutdownFunc,
		engine:      engine,
	}
}

// Name returns the plugin name
func (tp *TypeScriptPlugin) Name() string {
	return tp.name
}

// Version returns the plugin version
func (tp *TypeScriptPlugin) Version() string {
	return tp.version
}

// Initialize initializes the plugin
func (tp *TypeScriptPlugin) Initialize(ctx *PluginContext) error {
	if tp.initFunc == nil {
		return nil
	}
	
	tsCtx := tp.createContextObject(ctx)
	_, err := tp.initFunc(nil, tsCtx)
	if err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}
	return nil
}

// Execute executes the plugin
func (tp *TypeScriptPlugin) Execute(ctx *PluginContext, args map[string]interface{}) (interface{}, error) {
	if tp.execFunc == nil {
		return nil, fmt.Errorf("plugin has no execute function")
	}
	
	tsCtx := tp.createContextObject(ctx)
	argsObj := tp.engine.ToValue(args)
	
	result, err := tp.execFunc(nil, tsCtx, argsObj)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}
	
	return result.Export(), nil
}

// Shutdown shuts down the plugin
func (tp *TypeScriptPlugin) Shutdown() error {
	if tp.shutdownFunc == nil {
		return nil
	}
	
	_, err := tp.shutdownFunc(nil)
	if err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}
	return nil
}

// createContextObject creates a TypeScript context object
func (tp *TypeScriptPlugin) createContextObject(ctx *PluginContext) *goja.Object {
	ctxObj := tp.engine.NewObject()
	ctxObj.Set("runtimeID", ctx.RuntimeID)
	ctxObj.Set("config", tp.engine.ToValue(ctx.Config))
	
	// Logger object
	loggerObj := tp.engine.NewObject()
	loggerObj.Set("info", func(format string, args ...goja.Value) {
		// Convert goja.Value to interface{}
		interfaceArgs := make([]interface{}, len(args))
		for i, arg := range args {
			interfaceArgs[i] = arg.Export()
		}
		ctx.Logger.Info(format, interfaceArgs...)
	})
	loggerObj.Set("error", func(format string, args ...goja.Value) {
		// Convert goja.Value to interface{}
		interfaceArgs := make([]interface{}, len(args))
		for i, arg := range args {
			interfaceArgs[i] = arg.Export()
		}
		ctx.Logger.Error(format, interfaceArgs...)
	})
	ctxObj.Set("logger", loggerObj)
	
	return ctxObj
}

// TypeScriptPluginManager wraps PluginManager for TypeScript
type TypeScriptPluginManager struct {
	manager *PluginManager
	engine  *goja.Runtime
	mu      sync.RWMutex
}

// NewTypeScriptPluginManager creates a new TypeScript-wrapped plugin manager
func NewTypeScriptPluginManager(engine *goja.Runtime, manager *PluginManager) *TypeScriptPluginManager {
	return &TypeScriptPluginManager{
		manager: manager,
		engine:  engine,
	}
}

// ToJSObject converts the plugin manager to a JavaScript object
func (tpm *TypeScriptPluginManager) ToJSObject() *goja.Object {
	obj := tpm.engine.NewObject()
	
	// Register method
	obj.Set("register", func(pluginObj goja.Value) *goja.Promise {
		promise, resolve, reject := tpm.engine.NewPromise()
		
		go func() {
			if pluginObj == nil || goja.IsUndefined(pluginObj) {
				reject(tpm.engine.ToValue("plugin object is required"))
				return
			}
			
			plugin, ok := pluginObj.(*goja.Object)
			if !ok {
				reject(tpm.engine.ToValue("plugin must be an object"))
				return
			}
			
			name := plugin.Get("name").String()
			version := plugin.Get("version").String()
			
			initFunc, _ := goja.AssertFunction(plugin.Get("initialize"))
			execFunc, ok := goja.AssertFunction(plugin.Get("execute"))
			if !ok {
				reject(tpm.engine.ToValue("plugin must have an execute function"))
				return
			}
			
			shutdownFunc, _ := goja.AssertFunction(plugin.Get("shutdown"))
			
			tsPlugin := NewTypeScriptPlugin(tpm.engine, name, version, initFunc, execFunc, shutdownFunc)
			
			if err := tpm.manager.Register(tsPlugin); err != nil {
				reject(tpm.engine.ToValue(err.Error()))
			} else {
				resolve(tpm.engine.ToValue(true))
			}
		}()
		
		return promise
	})
	
	// Unregister method
	obj.Set("unregister", func(name string) *goja.Promise {
		promise, resolve, reject := tpm.engine.NewPromise()
		
		go func() {
			if err := tpm.manager.Unregister(name); err != nil {
				reject(tpm.engine.ToValue(err.Error()))
			} else {
				resolve(tpm.engine.ToValue(true))
			}
		}()
		
		return promise
	})
	
	// Execute method
	obj.Set("execute", func(name string, args goja.Value) *goja.Promise {
		promise, resolve, reject := tpm.engine.NewPromise()
		
		go func() {
			var argsMap map[string]interface{}
			if args != nil && !goja.IsUndefined(args) {
				argsMap = args.Export().(map[string]interface{})
			} else {
				argsMap = make(map[string]interface{})
			}
			
			// Create a basic context
			ctx := &PluginContext{
				RuntimeID: "ts-runtime",
				Config:    make(map[string]interface{}),
				Logger:    &TypeScriptLogger{engine: tpm.engine},
			}
			
			result, err := tpm.manager.Execute(name, ctx, argsMap)
			if err != nil {
				reject(tpm.engine.ToValue(err.Error()))
			} else {
				resolve(tpm.engine.ToValue(result))
			}
		}()
		
		return promise
	})
	
	// List method
	obj.Set("list", func() []string {
		return tpm.manager.ListPlugins()
	})
	
	// Get method
	obj.Set("get", func(name string) goja.Value {
		plugin, ok := tpm.manager.GetPlugin(name)
		if !ok {
			return goja.Undefined()
		}
		
		// Return plugin info
		pluginObj := tpm.engine.NewObject()
		pluginObj.Set("name", plugin.Name())
		pluginObj.Set("version", plugin.Version())
		return pluginObj
	})
	
	return obj
}

// TypeScriptLogger implements Logger for TypeScript
type TypeScriptLogger struct {
	engine *goja.Runtime
}

// Info logs an info message
func (tl *TypeScriptLogger) Info(format string, args ...interface{}) {
	// Log implementation - for now just a placeholder
	// In a full implementation, this would use the observability logger
	_ = format
	_ = args
}

// Error logs an error message
func (tl *TypeScriptLogger) Error(format string, args ...interface{}) {
	// Log implementation - for now just a placeholder
	// In a full implementation, this would use the observability logger
	_ = format
	_ = args
}

