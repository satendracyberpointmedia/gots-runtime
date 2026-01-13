// Standard Library: Plugin System
// TypeScript definitions for runtime extensions

export interface PluginContext {
    runtimeID: string;
    config: Record<string, any>;
    logger: Logger;
}

export interface Logger {
    info(format: string, ...args: any[]): void;
    error(format: string, ...args: any[]): void;
}

export interface Plugin {
    name: string;
    version: string;
    initialize(ctx: PluginContext): Promise<void> | void;
    execute(ctx: PluginContext, args: Record<string, any>): Promise<any> | any;
    shutdown(): Promise<void> | void;
}

export interface PluginManager {
    register(plugin: Plugin): Promise<void>;
    unregister(name: string): Promise<void>;
    execute(name: string, args: Record<string, any>): Promise<any>;
    list(): string[];
    get(name: string): Plugin | undefined;
}

// Factory function to get plugin manager
export function getPluginManager(): PluginManager;

// Factory function to create a plugin
export function createPlugin(name: string, version: string, init: (ctx: PluginContext) => Promise<void> | void, exec: (ctx: PluginContext, args: Record<string, any>) => Promise<any> | any, shutdown?: () => Promise<void> | void): Plugin;

