# Go-based Multithreaded Runtime with Inbuilt TypeScript

## Roadmap

A comprehensive roadmap for building a next-generation runtime environment that combines Golang's multithreading capabilities with TypeScript as a first-class citizen, designed to address Node.js limitations while maintaining event-driven simplicity.

---

## 1. Problem Statement

Node.js JavaScript runtime relies on a single-threaded event-loop architecture, which presents limitations in:
- High CPU-bound workloads
- True parallelism
- Predictable performance in certain cases

**Goal**: Use Go to create a runtime that is multithreaded by default while preserving event-driven simplicity.

---

## 2. High-Level Vision

1. **Golang-based runtime** (native binary)
2. **Multithreaded execution** using goroutines
3. **Event-driven programming model** preserved
4. **TypeScript as first-class citizen** (no plain JS)
5. **Node.js alternative**, not replacement

---

## 3. Core Architecture

1. **Main Orchestrator (Go)**: Process lifecycle, scheduler
2. **Event Loop Layer**: Logical event queue (not single OS thread)
3. **Worker Pool**: Goroutine-based parallel execution
4. **TS Execution Engine**: Embedded JS/TS engine (AST or bytecode based)
5. **IPC via Channels**: Safe data passing

---

## 4. Event Loop vs Multithreading (Key Design)

The event loop maintains the logical abstraction for handling I/O events, while CPU-heavy tasks are offloaded to worker goroutines. Users experience the single-threaded illusion, while the runtime is internally multithreaded.

**Design Principle**: Event loop handles I/O events, worker goroutines handle CPU-intensive tasks. This hybrid approach provides both simplicity and performance.

---

## 5. TypeScript Integration

1. **TS compiler inbuilt** (no external build step)
2. **Runtime-level type validation hooks**
3. **Strict TS-only execution** (JS disabled)

### Advanced TypeScript Features

1. **Zero-build TypeScript execution** (no tsc, no dist)
2. **Runtime type enforcement** based on TS types
3. **Strict TS-only execution** (plain JS disabled)

---

## 6. Advanced Runtime Capabilities

1. **Hybrid Concurrency Model**: Event-loop + Goroutines auto scheduling
2. **Deterministic Scheduling Mode**: Debug & prod parity
3. **Zero-cost worker thread abstraction**

---

## 7. Memory & Safety Architecture

1. **Per-module memory isolation**
2. **Automatic memory-leak detection**
3. **Crash containment** (one module crash ≠ whole app crash)

---

## 8. Modular-to-Microservice Evolution

1. **Built-in modular architecture enforcement**
2. **Controlled module boundaries & permissions**
3. **Module → Service promotion** with same codebase

---

## 9. Performance Considerations

1. **Goroutine scheduling overhead**
2. **Cross-thread communication cost**
3. **Memory isolation vs shared memory tradeoff**

### Performance & Backpressure Handling

1. **Runtime-managed backpressure**
2. **Event queue overload protection**
3. **Adaptive load shedding**

---

## 10. Known Challenges / Risks

1. **Race conditions** if APIs expose shared state
2. **Debugging complexity** vs Node.js
3. **Ecosystem maturity** (packages, tooling)
4. **Developer mental model shift**

---

## 11. How Issues Are Managed

1. **Immutable data by default**
2. **Message-passing** instead of shared memory
3. **Structured concurrency**
4. **Built-in profiler & tracer**

---

## 12. Observability & Operations

1. **Built-in metrics, tracing & logs**
2. **Health & readiness endpoints** by default
3. **Production-grade profiling tools**

---

## 13. Security & Sandboxing

1. **Permission-based API access** (fs, net, env)
2. **Deterministic sandbox execution mode**
3. **Safe execution of untrusted TypeScript**

### Capability-based Security Model

- **Permission-based API access** (fs, net, env)
- **Deterministic sandbox execution mode**
- **Secrets & Config Vault** (runtime managed)

---

## 14. Cloud-Native & DevEx

1. **Single static binary deployment**
2. **Fast cold-start** (serverless friendly)
3. **Hot reload** with state preservation

### Cloud & Infrastructure

1. **Runtime-aware Load Balancer**
2. **Native Serverless Mode**
3. **Single Binary Deployment Model**

---

## 15. Core Platform (Runtime के علاوہ)

### Official CLI Tool
- Command-line interface for development and deployment

### Opinionated Standard Library
- Curated, runtime-optimized standard library

### Runtime-aware Official Framework
- Framework designed specifically for the runtime architecture

---

## 16. Developer Experience (DX)

1. **Runtime-native Debugger**
2. **Built-in Test Runner**
3. **Zero-config Observability** (logs, metrics, traces)

---

## 17. Architecture & Scalability

1. **Domain-Driven Modules (DDD enforcement)**
2. **Service Graph Awareness**
3. **Native RPC System** (HTTP optional)

---

## 18. Ecosystem

1. **Curated Package System** (TS-only, audited)
2. **Plugin System** (runtime extensions)

---

## 19. Advanced / Future-ready (Non-AI)

1. **Deterministic Replay Engine**
2. **Multi-runtime Federation**

---

## 20. Phased Roadmap

### Phase 1: Minimal runtime + TS execution
- Basic runtime infrastructure
- TypeScript execution engine
- Core type system integration

### Phase 2: Event loop + async I/O
- Event loop implementation
- Asynchronous I/O operations
- Basic event queue management

### Phase 3: Worker pool & multithreading
- Goroutine-based worker pool
- Task scheduling and distribution
- Cross-thread communication

### Phase 4: HTTP server, APIs, modules
- HTTP server implementation
- Core APIs and module system
- Standard library foundation

### Phase 5: Tooling, CLI, debugger
- Command-line interface
- Development tools
- Debugging capabilities
- Testing framework

---

## Final Vision

This runtime is not intended to replace Node.js, but rather to address its limitations. It is designed to be an ideal solution for high-performance, type-safe, multithreaded backend workloads, particularly in cloud-native environments.

The runtime represents an evolution beyond Node.js, fundamentally designed for high-performance, type-safe, multithreaded, and cloud-native backend systems.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│              Main Orchestrator (Go)                      │
│         (Process Lifecycle, Scheduler)                   │
└──────────────────┬──────────────────────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
┌───────▼────────┐   ┌────────▼──────────┐
│  Event Loop    │   │   Worker Pool      │
│  Layer         │   │   (Goroutines)      │
│  (Logical)     │   │                    │
└───────┬────────┘   └────────┬───────────┘
        │                     │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────┐
        │  TS Execution       │
        │  Engine             │
        │  (AST/Bytecode)     │
        └─────────────────────┘
```

---

## Key Design Principles

1. **Type Safety First**: TypeScript-only execution ensures type safety at runtime
2. **Concurrency by Default**: Multithreading without explicit thread management
3. **Event-Driven Simplicity**: Familiar async/await patterns with true parallelism
4. **Cloud-Native**: Built for modern deployment scenarios
5. **Developer Experience**: Zero-config, production-ready tooling out of the box
6. **Security by Design**: Capability-based permissions and sandboxing
7. **Observability Built-in**: Metrics, tracing, and logging without external dependencies

---

## Contributing

This is a roadmap document. Implementation details and contributions will be documented as the project progresses.

---

## License

[To be determined]

---

## Installation & Stdlib Layout (Current Implementation)

To run the `gots` CLI from a compiled binary (including future MSI installers), the TypeScript standard library **must** be available so the runtime can start correctly.

The runtime resolves the stdlib directory in the following order:

1. `GOTS_STDLIB_PATH` environment variable (if set and points to an existing directory).
2. `stdlib` directory located **next to the `gots` executable**.
3. `./stdlib` relative to the current working directory (for development).
4. `../../stdlib` as a legacy/dev fallback.

### Recommended layout

- Place `gots` / `gots.exe` and the `stdlib/` folder side‑by‑side, for example:

  - `C:\Program Files\GoTSRuntime\gots.exe`
  - `C:\Program Files\GoTSRuntime\stdlib\...`

or set:

```bash
GOTS_STDLIB_PATH=C:\Program Files\GoTSRuntime\stdlib
```

With this in place, commands like `gots run main.ts` and `gots serve main.ts` can be executed from any project directory without `stdlib directory not found` errors, and the same layout can be used by an MSI installer.

