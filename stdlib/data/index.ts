// Standard Library: Immutable Data Structures
// TypeScript definitions for immutable data structures

export interface ImmutableMap<K, V> {
    get(key: K): V | undefined;
    has(key: K): boolean;
    set(key: K, value: V): ImmutableMap<K, V>;
    delete(key: K): ImmutableMap<K, V>;
    clear(): ImmutableMap<K, V>;
    size(): number;
    isEmpty(): boolean;
    keys(): K[];
    values(): V[];
    entries(): Array<[K, V]>;
    forEach(callback: (value: V, key: K) => void): void;
    map<U>(callback: (value: V, key: K) => U): ImmutableList<U>;
    filter(callback: (value: V, key: K) => boolean): ImmutableMap<K, V>;
    toJS(): Map<K, V>;
    toObject(): Record<string | number | symbol, V>;
    equals(other: ImmutableMap<K, V>): boolean;
}

export interface ImmutableList<T> {
    get(index: number): T | undefined;
    set(index: number, value: T): ImmutableList<T>;
    push(...values: T[]): ImmutableList<T>;
    pop(): [ImmutableList<T>, T | undefined];
    unshift(...values: T[]): ImmutableList<T>;
    shift(): [ImmutableList<T>, T | undefined];
    insert(index: number, ...values: T[]): ImmutableList<T>;
    delete(index: number): ImmutableList<T>;
    splice(index: number, deleteCount: number, ...items: T[]): [ImmutableList<T>, T[]];
    size(): number;
    isEmpty(): boolean;
    first(): T | undefined;
    last(): T | undefined;
    slice(start?: number, end?: number): ImmutableList<T>;
    concat(...lists: ImmutableList<T>[]): ImmutableList<T>;
    forEach(callback: (value: T, index: number) => void): void;
    map<U>(callback: (value: T, index: number) => U): ImmutableList<U>;
    filter(callback: (value: T, index: number) => boolean): ImmutableList<T>;
    reduce<U>(callback: (acc: U, value: T, index: number) => U, initial: U): U;
    find(callback: (value: T, index: number) => boolean): T | undefined;
    findIndex(callback: (value: T, index: number) => boolean): number;
    includes(value: T): boolean;
    indexOf(value: T): number;
    reverse(): ImmutableList<T>;
    sort(compareFn?: (a: T, b: T) => number): ImmutableList<T>;
    join(separator?: string): string;
    toJS(): T[];
    toArray(): T[];
    equals(other: ImmutableList<T>): boolean;
}

export interface ImmutableSet<T> {
    has(value: T): boolean;
    add(...values: T[]): ImmutableSet<T>;
    delete(value: T): ImmutableSet<T>;
    clear(): ImmutableSet<T>;
    size(): number;
    isEmpty(): boolean;
    values(): T[];
    entries(): Array<[T, T]>;
    forEach(callback: (value: T) => void): void;
    map<U>(callback: (value: T) => U): ImmutableSet<U>;
    filter(callback: (value: T) => boolean): ImmutableSet<T>;
    union(other: ImmutableSet<T>): ImmutableSet<T>;
    intersection(other: ImmutableSet<T>): ImmutableSet<T>;
    difference(other: ImmutableSet<T>): ImmutableSet<T>;
    isSubsetOf(other: ImmutableSet<T>): boolean;
    isSupersetOf(other: ImmutableSet<T>): boolean;
    toJS(): Set<T>;
    toArray(): T[];
    equals(other: ImmutableSet<T>): boolean;
}

export interface ImmutableQueue<T> {
    enqueue(...values: T[]): ImmutableQueue<T>;
    dequeue(): [ImmutableQueue<T>, T | undefined];
    peek(): T | undefined;
    size(): number;
    isEmpty(): boolean;
    toArray(): T[];
    equals(other: ImmutableQueue<T>): boolean;
}

export interface ImmutableStack<T> {
    push(...values: T[]): ImmutableStack<T>;
    pop(): [ImmutableStack<T>, T | undefined];
    peek(): T | undefined;
    size(): number;
    isEmpty(): boolean;
    toArray(): T[];
    equals(other: ImmutableStack<T>): boolean;
}

export interface ImmutableRecord {
    get(key: string): any;
    set(key: string, value: any): ImmutableRecord;
    delete(key: string): ImmutableRecord;
    merge(other: ImmutableRecord): ImmutableRecord;
    toJS(): Record<string, any>;
    equals(other: ImmutableRecord): boolean;
}

// Factory functions
export function createMap<K, V>(entries?: Array<[K, V]> | Record<string, V>): ImmutableMap<K, V> { throw new Error('Not implemented'); }
export function createList<T>(items?: T[] | Iterable<T>): ImmutableList<T> { throw new Error('Not implemented'); }
export function createSet<T>(items?: T[] | Iterable<T>): ImmutableSet<T> { throw new Error('Not implemented'); }
export function createQueue<T>(items?: T[] | Iterable<T>): ImmutableQueue<T> { throw new Error('Not implemented'); }
export function createStack<T>(items?: T[] | Iterable<T>): ImmutableStack<T> { throw new Error('Not implemented'); }
export function createRecord(values?: Record<string, any>): ImmutableRecord { throw new Error('Not implemented'); }
