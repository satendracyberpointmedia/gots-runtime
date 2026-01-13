package tsengine

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"sync"

	"github.com/dop251/goja"

	"gots-runtime/internal/api"
	"gots-runtime/internal/data"
	"gots-runtime/internal/eventloop"
	"gots-runtime/internal/framework"
	"gots-runtime/internal/observability"
	"gots-runtime/internal/plugin"
	"gots-runtime/internal/rpc"
	"gots-runtime/internal/security"
	"gots-runtime/internal/worker"
)

// RuntimeBindings provides TypeScript bindings for runtime APIs
type RuntimeBindings struct {
	engine      *Engine
	eventLoop   *eventloop.Loop
	permManager *security.PermissionManager
	moduleID    string
	mu          sync.RWMutex
}

// NewRuntimeBindings creates new runtime bindings
func NewRuntimeBindings(engine *Engine, eventLoop *eventloop.Loop, permManager *security.PermissionManager, moduleID string) *RuntimeBindings {
	return &RuntimeBindings{
		engine:      engine,
		eventLoop:   eventLoop,
		permManager: permManager,
		moduleID:    moduleID,
	}
}

// RegisterAPIs registers all runtime APIs to the TypeScript engine
func (rb *RuntimeBindings) RegisterAPIs() error {
	// Register FS API
	if err := rb.registerFS(); err != nil {
		return fmt.Errorf("failed to register FS API: %w", err)
	}
	
	// Register Net API
	if err := rb.registerNet(); err != nil {
		return fmt.Errorf("failed to register Net API: %w", err)
	}
	
	// Register Env API
	if err := rb.registerEnv(); err != nil {
		return fmt.Errorf("failed to register Env API: %w", err)
	}
	
	// Register HTTP API
	if err := rb.registerHTTP(); err != nil {
		return fmt.Errorf("failed to register HTTP API: %w", err)
	}
	
	// Register Crypto API
	if err := rb.registerCrypto(); err != nil {
		return fmt.Errorf("failed to register Crypto API: %w", err)
	}
	
	// Register Worker API
	if err := rb.registerWorker(); err != nil {
		return fmt.Errorf("failed to register Worker API: %w", err)
	}
	
	// Register Immutable Data API
	if err := rb.registerImmutableData(); err != nil {
		return fmt.Errorf("failed to register Immutable Data API: %w", err)
	}
	
	// Register Framework API
	if err := rb.registerFramework(); err != nil {
		return fmt.Errorf("failed to register Framework API: %w", err)
	}
	
	// Register RPC API
	if err := rb.registerRPC(); err != nil {
		return fmt.Errorf("failed to register RPC API: %w", err)
	}
	
	// Register Plugin API
	if err := rb.registerPlugin(); err != nil {
		return fmt.Errorf("failed to register Plugin API: %w", err)
	}
	
	// Register Profiler API
	if err := rb.registerProfiler(); err != nil {
		return fmt.Errorf("failed to register Profiler API: %w", err)
	}
	
	return nil
}

// registerFS registers file system API
func (rb *RuntimeBindings) registerFS() error {
	secureFS := api.NewSecureFS(rb.eventLoop, rb.permManager, rb.moduleID)
	
	// Create FS object for TypeScript
	fsObj := rb.engine.VM().NewObject()
	
	// Register async methods with promise-like callbacks
	fsObj.Set("readFile", func(path string, callback goja.Callable) {
		secureFS.ReadFile(path, func(data []byte, err error) {
			if callback != nil {
				if err != nil {
					_, _ = callback(nil, rb.engine.VM().ToValue(err.Error()))
				} else {
					_, _ = callback(rb.engine.VM().ToValue(string(data)), nil)
				}
			}
		})
	})
	
	fsObj.Set("writeFile", func(path string, data string, callback goja.Callable) {
		secureFS.WriteFile(path, []byte(data), 0644, func(err error) {
			if callback != nil {
				if err != nil {
					_, _ = callback(nil, rb.engine.VM().ToValue(err.Error()))
				} else {
					_, _ = callback(nil, nil)
				}
			}
		})
	})
	
	fsObj.Set("readDir", func(path string, callback goja.Callable) {
		secureFS.ReadDir(path, func(entries []fs.DirEntry, err error) {
			if callback != nil {
				if err != nil {
					_, _ = callback(nil, rb.engine.VM().ToValue(err.Error()))
				} else {
					entriesArray := rb.engine.VM().NewArray()
					for i, entry := range entries {
						entryObj := rb.engine.VM().NewObject()
						entryObj.Set("name", entry.Name())
						entryObj.Set("isDir", entry.IsDir())
						entriesArray.Set(fmt.Sprintf("%d", i), entryObj)
					}
					_, _ = callback(entriesArray, nil)
				}
			}
		})
	})
	
	// Register sync methods
	fsObj.Set("readFileSync", func(path string) string {
		data, err := secureFS.ReadFileSync(path)
		if err != nil {
			panic(rb.engine.VM().ToValue(err.Error()))
		}
		return string(data)
	})
	
	fsObj.Set("writeFileSync", func(path, data string) {
		if err := secureFS.WriteFileSync(path, []byte(data), 0644); err != nil {
			panic(rb.engine.VM().ToValue(err.Error()))
		}
	})
	
	rb.engine.Set("fs", fsObj)
	return nil
}

// registerNet registers network API
func (rb *RuntimeBindings) registerNet() error {
	secureNet := api.NewSecureNet(rb.eventLoop, rb.permManager, rb.moduleID)
	
	netObj := rb.engine.VM().NewObject()
	
	netObj.Set("dial", func(network, address string, callback goja.Callable) {
		secureNet.Dial(network, address, func(conn net.Conn, err error) {
			if callback != nil {
				if err != nil {
					_, _ = callback(nil, rb.engine.VM().ToValue(err.Error()))
				} else {
					connObj := rb.createConnObject(conn)
					_, _ = callback(connObj, nil)
				}
			}
		})
	})
	
	netObj.Set("listen", func(network, address string, callback goja.Callable) {
		secureNet.Listen(network, address, func(listener net.Listener, err error) {
			if callback != nil {
				if err != nil {
					_, _ = callback(nil, rb.engine.VM().ToValue(err.Error()))
				} else {
					listenerObj := rb.createListenerObject(listener)
					_, _ = callback(listenerObj, nil)
				}
			}
		})
	})
	
	rb.engine.Set("net", netObj)
	return nil
}

// registerHTTP registers HTTP API
func (rb *RuntimeBindings) registerHTTP() error {
	httpAPI := api.NewHTTP(rb.eventLoop)
	
	httpObj := rb.engine.VM().NewObject()
	
	// HTTP server
	httpObj.Set("createServer", func(callback goja.Callable) interface{} {
		server := httpAPI.NewServer(":0") // Dynamic port
		
		serverObj := rb.engine.VM().NewObject()
		serverObj.Set("listen", func(port int, callback goja.Callable) {
			server.ListenAndServe(func(err error) {
				if callback != nil {
					if err != nil {
						_, _ = callback(rb.engine.VM().ToValue(err.Error()))
					} else {
						_, _ = callback(nil)
					}
				}
			})
		})
		
		return serverObj
	})
	
	rb.engine.Set("http", httpObj)
	return nil
}

// registerEnv registers environment API
func (rb *RuntimeBindings) registerEnv() error {
	secureEnv := api.NewSecureEnv(rb.permManager, rb.moduleID)
	
	envObj := rb.engine.VM().NewObject()
	
	envObj.Set("get", func(key string) (string, error) {
		return secureEnv.Get(key)
	})
	
	envObj.Set("set", func(key, value string) error {
		return secureEnv.Set(key, value)
	})
	
	envObj.Set("lookup", func(key string) (interface{}, error) {
		value, ok, err := secureEnv.LookupEnv(key)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, nil
		}
		return value, nil
	})
	
	rb.engine.Set("env", envObj)
	return nil
}

// registerCrypto registers crypto API
func (rb *RuntimeBindings) registerCrypto() error {
	cryptoAPI := api.NewCrypto()
	
	cryptoObj := rb.engine.VM().NewObject()
	
	cryptoObj.Set("md5", func(data string) string {
		return cryptoAPI.MD5([]byte(data))
	})
	
	cryptoObj.Set("sha256", func(data string) string {
		return cryptoAPI.SHA256([]byte(data))
	})
	
	cryptoObj.Set("randomBytes", func(n int) string {
		bytes, err := cryptoAPI.RandomBytes(n)
		if err != nil {
			panic(rb.engine.VM().ToValue(err.Error()))
		}
		return string(bytes)
	})
	
	cryptoObj.Set("randomUUID", func() string {
		uuid, err := cryptoAPI.RandomUUID()
		if err != nil {
			panic(rb.engine.VM().ToValue(err.Error()))
		}
		return uuid
	})
	
	rb.engine.Set("crypto", cryptoObj)
	return nil
}

// createConnObject creates a connection object for TypeScript
func (rb *RuntimeBindings) createConnObject(conn net.Conn) *goja.Object {
	connObj := rb.engine.VM().NewObject()
	// Connection methods would be implemented here
	return connObj
}

// createListenerObject creates a listener object for TypeScript
func (rb *RuntimeBindings) createListenerObject(listener net.Listener) *goja.Object {
	listenerObj := rb.engine.VM().NewObject()
	// Listener methods would be implemented here
	return listenerObj
}

// registerWorker registers worker thread API
func (rb *RuntimeBindings) registerWorker() error {
	vm := rb.engine.VM()
	// Get context from orchestrator via runtime integration
	// For now, use background context - this should be passed from runtime integration
	ctx := context.Background()
	
	// Create default worker pool (min 2, max 10 workers)
	defaultWorker := worker.NewTypeScriptWorker(ctx, vm, 2, 10)
	
	// Create worker namespace
	workerObj := vm.NewObject()
	
	// Create worker pool factory
	workerObj.Set("createPool", func(minWorkers, maxWorkers int) *goja.Object {
		if minWorkers <= 0 {
			minWorkers = 2
		}
		if maxWorkers <= 0 {
			maxWorkers = 10
		}
		if maxWorkers < minWorkers {
			maxWorkers = minWorkers
		}
		
		pool := worker.NewTypeScriptWorker(ctx, vm, minWorkers, maxWorkers)
		
		poolObj := vm.NewObject()
		poolObj.Set("spawn", func(taskID string, handler goja.Callable, data goja.Value) *goja.Promise {
			return pool.Spawn(taskID, handler, data)
		})
		poolObj.Set("spawnBatch", func(tasks goja.Value) *goja.Promise {
			if tasksArray, ok := tasks.(*goja.Object); ok {
				length := tasksArray.Get("length").ToInteger()
				taskSlice := make([]interface{}, length)
				for i := int64(0); i < length; i++ {
					taskSlice[i] = tasksArray.Get(fmt.Sprintf("%d", i))
				}
				return pool.SpawnBatch(taskSlice)
			}
			promise, _, reject := vm.NewPromise()
			reject(vm.ToValue("tasks must be an array"))
			return promise
		})
		poolObj.Set("getStats", func() map[string]interface{} {
			return pool.GetStats()
		})
		poolObj.Set("close", func() *goja.Promise {
			promise, resolve, reject := vm.NewPromise()
			go func() {
				if err := pool.Close(); err != nil {
					reject(vm.ToValue(err.Error()))
				} else {
					resolve(vm.ToValue(true))
				}
			}()
			return promise
		})
		
		return poolObj
	})
	
	// Create spawnWorker convenience function
	workerObj.Set("spawn", func(taskID string, handler goja.Callable, data goja.Value) *goja.Promise {
		return defaultWorker.Spawn(taskID, handler, data)
	})
	
	// Expose worker API
	rb.engine.Set("worker", workerObj)
	
	return nil
}

// registerImmutableData registers immutable data structures API
func (rb *RuntimeBindings) registerImmutableData() error {
	vm := rb.engine.VM()
	
	// Create data namespace
	dataObj := vm.NewObject()
	
	// Create Map factory
	dataObj.Set("createMap", func(entries goja.Value) *goja.Object {
		im := data.NewImmutableMap()
		
		// If entries provided, add them
		if entries != nil && !goja.IsUndefined(entries) {
			if entriesArray, ok := entries.(*goja.Object); ok {
				length := entriesArray.Get("length").ToInteger()
				for i := int64(0); i < length; i++ {
					entry := entriesArray.Get(fmt.Sprintf("%d", i))
					if entryArray, ok := entry.(*goja.Object); ok {
						key := entryArray.Get("0").Export()
						value := entryArray.Get("1").Export()
						im = im.Set(fmt.Sprintf("%v", key), value)
					}
				}
			}
		}
		
		return data.NewTypeScriptImmutableMap(vm, im).ToJSObject()
	})
	
	// Create List factory
	dataObj.Set("createList", func(items goja.Value) *goja.Object {
		il := data.NewImmutableList()
		
		// If items provided, add them
		if items != nil && !goja.IsUndefined(items) {
			if itemsArray, ok := items.(*goja.Object); ok {
				length := itemsArray.Get("length").ToInteger()
				for i := int64(0); i < length; i++ {
					item := itemsArray.Get(fmt.Sprintf("%d", i)).Export()
					il = il.Append(item)
				}
			}
		}
		
		return data.NewTypeScriptImmutableList(vm, il).ToJSObject()
	})
	
	// Create Set factory
	dataObj.Set("createSet", func(items goja.Value) *goja.Object {
		is := data.NewImmutableSet()
		
		// If items provided, add them
		if items != nil && !goja.IsUndefined(items) {
			if itemsArray, ok := items.(*goja.Object); ok {
				length := itemsArray.Get("length").ToInteger()
				for i := int64(0); i < length; i++ {
					item := itemsArray.Get(fmt.Sprintf("%d", i)).Export()
					is = is.Add(item)
				}
			}
		}
		
		return data.NewTypeScriptImmutableSet(vm, is).ToJSObject()
	})
	
	// Expose data API
	rb.engine.Set("data", dataObj)
	
	return nil
}

// registerFramework registers the runtime-aware framework API
func (rb *RuntimeBindings) registerFramework() error {
	vm := rb.engine.VM()
	
	// Create framework namespace
	frameworkObj := vm.NewObject()
	
	// Create app factory
	frameworkObj.Set("createApp", func(name goja.Value) *goja.Object {
		appName := "gots-app"
		if name != nil && !goja.IsUndefined(name) {
			appName = name.String()
		}
		
		tsApp := framework.NewTypeScriptApp(vm, rb.eventLoop, appName)
		return tsApp.ToJSObject()
	})
	
	// Expose framework API
	rb.engine.Set("framework", frameworkObj)
	
	return nil
}

// registerRPC registers the native RPC system API
func (rb *RuntimeBindings) registerRPC() error {
	vm := rb.engine.VM()
	ctx := context.Background()
	
	// Create RPC namespace
	rpcObj := vm.NewObject()
	
	// Create server factory
	rpcObj.Set("createServer", func() *goja.Object {
		server := rpc.NewTypeScriptRPCServer(vm, ctx)
		return server.ToJSObject()
	})
	
	// Create client factory
	rpcObj.Set("createClient", func(address string) *goja.Promise {
		promise, resolve, reject := vm.NewPromise()
		
		go func() {
			client, err := rpc.NewTypeScriptRPCClient(vm, address)
			if err != nil {
				reject(vm.ToValue(err.Error()))
			} else {
				resolve(client.ToJSObject())
			}
		}()
		
		return promise
	})
	
	// Expose RPC API
	rb.engine.Set("rpc", rpcObj)
	
	return nil
}

// registerPlugin registers the plugin system API
func (rb *RuntimeBindings) registerPlugin() error {
	vm := rb.engine.VM()
	
	// Create plugin manager
	manager := plugin.NewPluginManager()
	tsManager := plugin.NewTypeScriptPluginManager(vm, manager)
	
	// Create plugin namespace
	pluginObj := vm.NewObject()
	
	// Get plugin manager
	pluginObj.Set("getPluginManager", func() *goja.Object {
		return tsManager.ToJSObject()
	})
	
	// Create plugin factory
	pluginObj.Set("createPlugin", func(name, version string, initFunc, execFunc, shutdownFunc goja.Value) *goja.Object {
		// Validate execFunc is a function
		_, ok := goja.AssertFunction(execFunc)
		if !ok {
			panic(vm.ToValue("execute function is required"))
		}
		
		pluginObj := vm.NewObject()
		pluginObj.Set("name", name)
		pluginObj.Set("version", version)
		pluginObj.Set("initialize", initFunc)
		pluginObj.Set("execute", execFunc)
		pluginObj.Set("shutdown", shutdownFunc)
		
		return pluginObj
	})
	
	// Expose plugin API
	rb.engine.Set("plugin", pluginObj)
	
	return nil
}

// registerProfiler registers the profiler API
func (rb *RuntimeBindings) registerProfiler() error {
	vm := rb.engine.VM()
	
	// Create profiler
	profiler := observability.NewTypeScriptProfiler(vm)
	
	// Create profiler namespace
	profilerObj := vm.NewObject()
	
	// Get profiler
	profilerObj.Set("getProfiler", func() *goja.Object {
		return profiler.ToJSObject()
	})
	
	// Expose profiler API
	rb.engine.Set("profiler", profilerObj)
	
	return nil
}

