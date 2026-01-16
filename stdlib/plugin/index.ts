// Standard Library: Plugin System
// TypeScript definitions for runtime extensions

export interface Logger {
    info(message: string, ...args: any[]): void;
    warn(message: string, ...args: any[]): void;
    error(message: string, ...args: any[]): void;
    debug(message: string, ...args: any[]): void;
}

export interface PluginContext {
    runtimeID: string;
    config: Record<string, any>;
    logger: Logger;
    setData(key: string, value: any): void;
    getData(key: string): any;
}

export interface Plugin {
    name: string;
    version: string;
    description?: string;
    dependencies?: string[];

    initialize(ctx: PluginContext): Promise<void> | void;
    execute(ctx: PluginContext, args: Record<string, any>): Promise<any> | any;
    shutdown(): Promise<void> | void;

    getMetadata(): PluginMetadata;
}

export interface PluginMetadata {
    name: string;
    version: string;
    description?: string;
    author?: string;
    license?: string;
    entryPoint?: string;
}

export interface PluginLoadOptions {
    path?: string;
    config?: Record<string, any>;
    priority?: number; // 0-100, higher = earlier execution
}

export interface PluginManager {
    register(plugin: Plugin, options?: PluginLoadOptions): Promise<void>;
    unregister(name: string): Promise<void>;
    load(path: string, config?: Record<string, any>): Promise<Plugin>;

    execute(name: string, args: Record<string, any>): Promise<any>;
    executeAll(hookName: string, args?: Record<string, any>): Promise<any[]>;

    list(): PluginMetadata[];
    get(name: string): Plugin | undefined;
    has(name: string): boolean;

    shutdown(): Promise<void>;
}

export interface HookManager {
    register(hookName: string, handler: (args: Record<string, any>) => Promise<any> | any): void;
    unregister(hookName: string, handler: Function): void;
    execute(hookName: string, args?: Record<string, any>): Promise<any[]>;
}

// Factory function to get plugin manager
export function getPluginManager(): PluginManager { throw new Error('Not implemented'); }

// Factory function to get hook manager
export function getHookManager(): HookManager { throw new Error('Not implemented'); }

// Factory function to create a plugin
export function createPlugin(
    name: string,
    version: string,
    init: (ctx: PluginContext) => Promise<void> | void,
    exec: (ctx: PluginContext, args: Record<string, any>) => Promise<any> | any,
    shutdown?: () => Promise<void> | void,
    metadata?: PluginMetadata
): Plugin { throw new Error('Not implemented'); }

// Utility functions
export function isPlugin(obj: any): boolean { throw new Error('Not implemented'); }
export function validatePlugin(plugin: Plugin): string[] { throw new Error('Not implemented'); } // Returns validation errors
