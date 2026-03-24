# Product Store Performance Analysis

This project implements a high-performance product store using a hybrid storage approach: a **B-Tree** for range indexing by price and a **Map** for O(1) primary key lookups.

## Architecture Overview

The `ProductStore` is designed to handle large datasets (100K+ products) with efficient retrieval patterns:

-   **Primary Lookup**: A `map[string]*Product` provides instantaneous access by Product ID.
-   **Range Index**: A **B-Tree (Order 50)** indexes products by `Price`. This allows for O(log N) range queries, which are significantly faster than linear scans as the dataset grows.
-   **Concurrency**: Protected by `sync.RWMutex` to allow concurrent reads while ensuring thread-safe writes.

## Middleware Stack

The server utilizes a robust middleware pipeline to enhance observability and reliability:

1.  **Error Recovery**: `ErrorRecoveryMiddleware` catches panics and returns a 500 Internal Server Error, preventing the server from crashing.
2.  **Logging**: `LoggingMiddleware` uses **Uber-Zap** for high-performance structured logging of every request (method, path, status, duration).
3.  **Request Timing**: `RequestTimingMiddleware` adds an `X-Response-Time` header to every response for client-side latency tracking.
4.  **JSON Content-Type**: `JSONContentTypeMiddleware` ensures all responses have the `application/json` header set.

---

## Profiling & Performance Results

### Range Query Performance (100K Products, 1000 Queries)

| Metric | B-Tree (Order 50) | Linear Scan | Difference |
| :--- | :--- | :--- | :--- |
| **Total Time (1000 queries)** | 20.32s | 24.88s | **19% Faster** |
| **Average Query Latency** | 20.3ms | 24.9ms | - |

> [!IMPORTANT]
> **Bottleneck Analysis**: Profiling revealed that ~85% of query time is spent in **JSON Serialization** (`json.Marshal`) and network I/O, rather than the search itself. While the B-Tree search is orders of magnitude faster than a linear scan, the overall API latency is dominated by the overhead of returning large result sets (~5% of the 100K products per query).

### Memory Footprint (100K Products)

-   **Map + B-Tree Store**: ~65MB (Measured via Heap Profile)
-   **Flat Slice (Estimated)**: ~25MB (100K pointers + structs)
-   **Trade-off**: The indexing structures introduce ~2.5x memory overhead to provide rapid multi-dimensional lookups.

---

## Observability & pprof

### How to Read pprof Profiles

-   **Heap Profile**: `heap profile: A: B [C: D] @ heap/E`
    -   `A/B`: Current memory usage (`inuse_objects` / `inuse_space`).
    -   `C/D`: Lifetime allocations (`alloc_objects` / `alloc_space`).
-   **Mutex Profile**: Identifies lock contention. Look for functions above the center line (Callers) and below (Callees) to trace where threads are blocking.

### Technical Setup

1.  **Start Server**:
    ```bash
    go run cmd/server/main.go
    ```
2.  **Seed Data** (e.g., 100,000 products):
    ```bash
    curl "http://localhost:8080/seed?count=100000"
    ```
3.  **Run Load Test**:
    ```bash
    go run scripts/load.go
    ```
4.  **Capture Profiles**:
    -   CPU Profile: `go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30`
    -   Heap Profile: `go tool pprof http://localhost:6060/debug/pprof/heap`
    -   Mutex Profile: `go tool pprof http://localhost:6060/debug/pprof/mutex`
