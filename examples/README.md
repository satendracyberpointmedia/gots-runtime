# GoTS Runtime Examples

This directory contains example applications demonstrating the GoTS Runtime features.

## Quick Start

### 1. Hello World
Basic TypeScript program with strict typing:
```bash
gots run hello.ts
```

### 2. Web Server
HTTP server with routing and middleware:
```bash
gots serve server.ts
# Visit http://localhost:3000
```

Features:
- Dynamic routing with path parameters
- JSON API responses
- Error handling
- Hot reload support

### 3. Unit Testing
Comprehensive test suite examples:
```bash
gots test
```

Features:
- Multiple describe blocks
- Various assertion types
- Test isolation
- Async test support

### 4. Async/Await
Asynchronous programming patterns:
```bash
gots run async.ts
```

Features:
- Promise-based async operations
- Error handling with try/catch
- Parallel execution with Promise.all()
- Proper cleanup

### 5. Worker Pools
Concurrent task execution with workers:
```bash
gots run workers.ts
```

Features:
- Worker pool creation
- Task submission
- Parallel processing
- Result aggregation

### 6. Data Structures
Immutable data structure usage:
```bash
gots run datastructures.ts
```

Features:
- Immutable Map
- Immutable List
- Immutable Set
- Immutable Queue/Stack
- Record objects

### 7. Cryptography
Cryptographic operations:
```bash
gots run crypto.ts
```

Features:
- Hashing (MD5, SHA-256, SHA-512)
- Random generation
- HMAC operations
- Utility functions (Base64, Hex encoding)

## Running with Debugger

Debug any example:
```bash
gots debug <example.ts>
```

Debugger commands:
- `continue` - Continue execution
- `step` - Execute next line
- `break <line>` - Set breakpoint
- `inspect` - Inspect variables
- `quit` - Exit debugger

## Running with Profiler

Profile performance:
```bash
gots profile <example.ts>
```

## Running with Linter

Check code quality:
```bash
gots lint
```

## Running with Formatter

Auto-format code:
```bash
gots fmt
```

## Project Structure

Each example demonstrates different features:

- `hello.ts` - Basic types and functions
- `server.ts` - Web framework and routing
- `tests.test.ts` - Testing framework
- `async.ts` - Asynchronous patterns
- `workers.ts` - Concurrency and worker pools
- `datastructures.ts` - Immutable data structures
- `crypto.ts` - Cryptographic operations

## Development Workflow

1. **Write code** in `.ts` files
2. **Test with** `gots test` or `gots run`
3. **Debug with** `gots debug`
4. **Serve with** `gots serve` (with hot reload)
5. **Profile with** `gots profile`
6. **Lint/Format** with `gots lint` and `gots fmt`

## Documentation

View documentation:
```bash
gots doc
gots doc typescript
gots doc async
```

## API Documentation

Auto-generated API docs when serving:
```bash
gots serve server.ts
# Visit http://localhost:3000/api/docs
```

## Performance Tips

1. Use immutable data structures
2. Leverage worker pools for CPU-bound work
3. Use async/await for I/O operations
4. Enable profiling to identify bottlenecks
5. Watch for memory leaks with the crash detector

## Next Steps

- Explore the Framework documentation
- Build your own application
- Check out the Runtime API
- Learn about Security and Sandboxing
- Review Plugin System

Happy coding with GoTS Runtime!
