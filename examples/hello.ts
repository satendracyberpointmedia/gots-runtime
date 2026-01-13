// Example TypeScript file for GoTS Runtime
// This demonstrates basic TypeScript execution

interface Greeter {
    greet(name: string): string;
}

class HelloGreeter implements Greeter {
    greet(name: string): string {
        return `Hello, ${name}! Welcome to GoTS Runtime.`;
    }
}

const greeter: HelloGreeter = new HelloGreeter();
const message: string = greeter.greet("World");

console.log(message);

// Export for module system
export { greeter, message };

