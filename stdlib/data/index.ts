// Standard Library: Immutable Data Structures
// TypeScript definitions for immutable data structures

export interface ImmutableMap<K, V> {
    get(key: K): V | undefined;
    has(key: K): boolean;
    set(key: K, value: V): ImmutableMap<K, V>;
    delete(key: K): ImmutableMap<K, V>;
    size(): number;
    keys(): K[];
    values(): V[];
    entries(): Array<[K, V]>;
    forEach(callback: (value: V, key: K) => void): void;
    toJS(): Map<K, V>;
}

export interface ImmutableList<T> {
    get(index: number): T | undefined;
    set(index: number, value: T): ImmutableList<T>;
    push(value: T): ImmutableList<T>;
    pop(): [ImmutableList<T>, T | undefined];
    unshift(value: T): ImmutableList<T>;
    shift(): [ImmutableList<T>, T | undefined];
    size(): number;
    forEach(callback: (value: T, index: number) => void): void;
    map<U>(callback: (value: T, index: number) => U): ImmutableList<U>;
    filter(callback: (value: T, index: number) => boolean): ImmutableList<T>;
    toJS(): T[];
}

export interface ImmutableSet<T> {
    has(value: T): boolean;
    add(value: T): ImmutableSet<T>;
    delete(value: T): ImmutableSet<T>;
    size(): number;
    values(): T[];
    forEach(callback: (value: T) => void): void;
    toJS(): Set<T>;
}

// Factory functions
export function createMap<K, V>(entries?: Array<[K, V]>): ImmutableMap<K, V>;
export function createList<T>(items?: T[]): ImmutableList<T>;
export function createSet<T>(items?: T[]): ImmutableSet<T>;

