// Data Structures Example
// Run with: gots run datastructures.ts

import {
    createMap,
    createList,
    createSet,
    createQueue,
    createStack,
    createRecord,
} from "../stdlib/data/index";

// Immutable Map
const map = createMap<string, number>([
    ["one", 1],
    ["two", 2],
    ["three", 3],
]);

console.log("Map operations:");
console.log("Get 'one':", map.get("one"));
console.log("Has 'two':", map.has("two"));
console.log("Size:", map.size());

// Immutable List
const list = createList<number>([1, 2, 3, 4, 5]);

console.log("\nList operations:");
console.log("Get at index 2:", list.get(2));
console.log("Length:", list.size());
console.log("Contains 3:", list.includes(3));

// Immutable Set
const set = createSet<string>(["apple", "banana", "cherry"]);

console.log("\nSet operations:");
console.log("Has 'apple':", set.has("apple"));
console.log("Size:", set.size());

// Immutable Queue
const queue = createQueue<number>([1, 2, 3]);

console.log("\nQueue operations:");
console.log("Peek:", queue.peek());
console.log("Size:", queue.size());

// Immutable Stack
const stack = createStack<string>(["first", "second", "third"]);

console.log("\nStack operations:");
console.log("Peek:", stack.peek());
console.log("Size:", stack.size());

// Immutable Record
const config = createRecord({
    host: "localhost",
    port: 3000,
    debug: true,
});

console.log("\nRecord operations:");
console.log("Config object created successfully");

export { };
