// Web Server Example
// Run with: gots serve server.ts
// Visit: http://localhost:3000

import { createServer } from "../stdlib/http/server";
import { createApp } from "../stdlib/framework/index";

const app = createApp("WebServerExample");

// Define routes
app.get("/", (ctx) => {
    ctx.response.body = "Hello from GoTS Web Server!";
    ctx.response.status = 200;
});

app.get("/api/users", (ctx) => {
    const users = [
        { id: 1, name: "Alice" },
        { id: 2, name: "Bob" },
        { id: 3, name: "Charlie" },
    ];
    ctx.response.body = JSON.stringify(users);
    ctx.response.status = 200;
});

app.get("/api/users/:id", (ctx) => {
    const userId = ctx.request.params?.id || "unknown";
    const user = { id: userId, name: `User ${userId}` };
    ctx.response.body = JSON.stringify(user);
    ctx.response.status = 200;
});

app.post("/api/users", (ctx) => {
    const body = JSON.parse(ctx.request.body.toString());
    const newUser = { id: 4, ...body };
    ctx.response.body = JSON.stringify(newUser);
    ctx.response.status = 201;
});

// Error handler
app.setErrorHandler((ctx, error) => {
    console.error("Error:", error);
    ctx.response.status = 500;
    ctx.response.body = JSON.stringify({ error: error.message });
});

// Start server
app.listen(3000, () => {
    console.log("Server running on http://localhost:3000");
});

export { app };
