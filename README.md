# Task Management Service (Go) — Cursor Pagination

This repository implements a Task Management microservice in Go, featuring CRUD operations, **cursor-based pagination**, and status filtering. It demonstrates sound separation of concerns, Go module best practices, and microservice-ready design.

---

## Features

- Create, read, update, and delete tasks (CRUD)
- Cursor-based pagination on `GET /tasks` using `cursor` and `limit`
- Filter by `status` (e.g., `?status=Completed`)
- In-memory repository (easily swappable for a database like Postgres)
- Clear API documentation and runnable example

---

## API Reference

### Health Check  
- **Request**: `GET /health`  
- **Response**: `200 OK`  
```json
{ "status": "ok" }
```
### Create Task
Request: POST /tasks

```json
Body:
{
  "title": "Write docs",
  "description": "First draft",
  "status": "Pending"
}
```
- **Response**: 201 Created → Returns created task JSON

### Get Task
- **Request**: GET /tasks/{id}

- **Response**: 200 OK → Returns the task JSON

### Update Task
- **Request**: PUT /tasks/{id}

Body :
```json
{
  "title": "Updated title",
  "status": "InProgress"
}
```
- **Response**: 200 OK → Returns updated task JSON

### Delete Task
- **Request**: DELETE /tasks/{id}

- **Response**: 204 No Content

### List Tasks — Cursor Pagination + Filter

- **Request**: GET /tasks?limit=10&cursor=<opaque>&status=Completed

- **Response**: 200 OK
```json
{
  "data": [ /* array of task objects */ ],
  "next_cursor": "<base64-encoded cursor>",
  "limit": 10
}
```
#### Cursor Format
The next_cursor is an opaque string—a Base64-encoded JSON object:
```json
{
  "t": "<RFC3339Nano timestamp>",
  "id": <int64>
}
```
t: The createdAt timestamp of the last task in the response
id: The unique ID of that task
Clients pass this cursor back via GET /tasks to fetch the next page, with filtering applied first.

## Getting Started
### Prerequisites
Go 1.22+

### Running the Service
```json
# Ensure module path is set to your GitHub repo, e.g. github.com/alle-assignment
go mod tidy
go run ./cmd/server
```
The service will run on `localhost:8080`

## Sample `curl` Commands

Use `curl` to interact with the API. The examples below demonstrate each action:

### 1. Health Check
```json
curl -i http://localhost:8080/health
```

Should return `200 OK` with `{ "status": "ok" }`

----------

### 2. Create Task
```json
curl -i -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Write README","description":"Draft documentation","status":"Pending"}'
```
-   **Response**: `201 Created` with JSON of created task
    

----------

### 3. Get Task by ID

```json
curl -i http://localhost:8080/tasks/1
```

-   **Response**: `200 OK` with task JSON, or `404 Not Found` if it doesn't exist
    

----------

### 4. Update Task

```json
curl -i -X PUT http://localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"status":"InProgress"}
```  

-   **Response**: `200 OK` with updated task JSON
    

----------

### 5. Delete Task
```json
curl -i -X DELETE http://localhost:8080/tasks/1
```

-   **Response**: `204 No Content`; subsequent GET returns `404`
    

----------

### 6. List Tasks (Cursor Pagination with Optional Filtering)

Fetch first page:
```json
curl -i "http://localhost:8080/tasks?limit=5"
``` 

Response includes:

-   `data`: array of tasks
    
-   `next_cursor`: Base64-encoded pagination token
    
-   `limit`: echoed number of items returned
    

Fetch next page using `next_cursor`:

```json
curl -i "http://localhost:8080/tasks?limit=5&cursor=<paste-next_cursor-here>"
```


## High-Level Design

### Single Responsibility Principle (SRP)
- **httpapi**: Handles HTTP routing and request/response lifecycle.
- **task**: Encapsulates domain logic, models, business rules, and pagination logic.
- **platform/response**: Manages JSON response formatting and error handling.

---

### Cursor Pagination
- **Benefits**:
  - More efficient and stable than offset pagination, especially for dynamic datasets.
  - Uses opaque encoding to abstract pagination details from clients.

---

### Stateless Architecture & Scalability
- Stateless HTTP server allows **horizontal scaling**.
- Repository interface makes it easy to **swap in a shared database** (e.g., Postgres) for production.
- **Containerization + Kubernetes + HPA** can be used for elastic scalability.

---
### Inter-Service Communication: Detailed Steps
- **REST** or **gRPC** for synchronous APIs.
- **Kafka/NATS** for asynchronous, event-driven use cases such as `task.created` events.

#### 1. RESTful Communication (synchronous HTTP JSON)
**Use case**: Task Service needs to fetch User details from a User Service on-demand.

**Steps**:
1. **Expose a REST endpoint** in User Service, e.g.:
```json
GET /users/{userId}  
```
  **Response**: { "id": "...", "name": "...", ... }
2. **Task Service** makes an HTTP GET request using `net/http`, e.g., `http.Get("http://user-svc/users/123")`.
3. **Parse JSON response** into a Go struct and handle errors or missing users.
4. **Use the user data** to populate task owner information or validate permissions during task creation.
5. **Add retry logic, timeouts**, and **circuit breakers** to handle API failures gracefully and improve resilience.  

#### 2. gRPC Communication (synchronous, high-performance)
**Use case**: Task Service needs frequent, low-latency interaction with User Service (e.g., permission check per request).
**Steps**:
1. **Define a `.proto` file** with service and messages, e.g.:
```json
proto
service UserService {
  rpc GetUser (GetUserRequest) returns (User);
}
```
2. **Generate Go client and server stubs** using `protoc`.
    
3. **Implement gRPC server** in User Service and **gRPC client** in Task Service.
    
4. **Call `GetUser(...)`** directly within Task handlers gRPC ensures efficient binary communication over HTTP/2.
    
5. **Handle context** timeouts, retries, and errors for robustness.
    
6. **Beneficial for** strong typing, streaming, backward compatibility, and high-throughput needs.

#### Message Queue Communication (asynchronous, event-driven)

**Use case**: After task creation, Task Service notifies other services (e.g., Notification or Audit Services) without waiting for them to respond.

**Steps**:

1.  **Publish an event** to a message broker (e.g., Kafka, RabbitMQ) with a topic like `task.created`, including payload `{ "taskId": 1, "userId": 123, ... }`.
    
2.  **Consumer services** (e.g., Notification Service) **subscribe** to the topic and process events asynchronously.
    
3.  The Task Service continues without delay, decoupling creation from downstream tasks.
    
4.  Implement **idempotency**, retries, and ordering safeguards to ensure reliable processing.
    
5.  **Scales well** and improves resilience, particularly for non-critical or bulk workflows.

---

## How It Satisfies Assignment Requirements
- Implements **cursor-based pagination** with opaque cursors.
- Supports **status-based filtering**.
- Follows **clean module and dependency design** using idiomatic Go.
- Provides **full API documentation** and **design explanations** in the README.

---

## Future Enhancements
- Migrate to **Postgres** with schema migrations.
- Add **Dockerfile**, **Kubernetes manifests**.
- Integrate **authentication/authorization** (JWT, API keys).
- Add **integration tests** and a **CI workflow**.
- Include **metrics** and **distributed tracing** for observability.

### Inter service communication
#### Hybrid Architecture — Combine All Three Styles
  Most scalable systems mix communication styles:
- **REST**: For external or team-facing APIs.
- **gRPC**: For fast, internal microservice-to-microservice calls.
- **Message queues**: For asynchronous processing, event streams, and decoupling.


## FAQs

### Do Clients Need to Store the Cursor Between Requests?

**Yes.** With cursor-based pagination, the client receives a `next_cursor` value in each paginated API response. Clients must save this cursor and include it in their next request to fetch the subsequent page.

#### Why This Is Required

- The cursor serves as a **marker** identifying where the last fetch ended. The server uses it to return the next batch of records consistently. ([StackOverflow example](https://stackoverflow.com/questions/18314687/how-to-implement-cursors-for-pagination-in-an-api))  
- Unlike offset-based pagination (e.g. `page`, `offset`), cursor pagination doesn't reset positions. The cursor ensures reliability even in dynamic datasets. ([StackOverflow explanation](https://stackoverflow.com/questions/55744926/offset-pagination-vs-cursor-pagination))  

---

## References & Further Reading
- [Cursor vs Offset Pagination Pros & Cons (StackOverflow)](https://stackoverflow.com/questions/29901252/what-are-the-differences-between-offset-and-cursor-pagination)
- [Pagination Best Practices for APIs (Speakeasy)](https://www.speakeasy.com/blog/api-pagination)
- [REST API Pagination Guide & Keyset Patterns (Medium)](https://medium.com/swlh/keyset-pagination-implementation-in-rest-apis-using-spring-boot-and-jpa-563a09dcf6d9)
