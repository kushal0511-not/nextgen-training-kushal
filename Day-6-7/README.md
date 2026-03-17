# Ride-Sharing Dispatch System (Go)

A high-performance, in-memory ride-sharing dispatch system implemented in Go, featuring custom data structures and efficient spatial matching logic.

## Architecture Diagram

```ascii
+-------------------+       +-----------------------+
|  CLI Interface    | <---> |    Dispatcher Logic   |
|   (cmd/main)      |       |  (internal/dispatch)  |
+-------------------+       +-----------+-----------+
                                        |
               +------------------------+------------------------+
               |                        |                        |
    +----------v----------+  +----------v----------+  +----------v----------+
    |    Min-Heap PQ      |  |   Driver Store      |  |   Ride Tracker      |
    |  (internal/queue)   |  |  (internal/driver)  |  |   (internal/rides)  |
    |  [Oldest First]     |  |  [HashMap + Grid]   |  |   [Doubly L-List]   |
    +---------------------+  +---------------------+  +---------------------+
```

## Data Structure Trade-offs & Complexity Analysis

### 1. Priority Queue (Min-Heap)
- **Choice**: Custom Min-Heap based on `RequestTime`.
- **Complexity**:
  - Enqueue: O(log N)
  - Dequeue: O(log N)
  - Peek: O(1)
- **Trade-off**: Memory efficient as it uses a slice-based binary heap. Provides strict ordering by wait time.

### 2. Driver Store (HashMap + Grid Index)
- **Choice**: `map[string]*Driver` for ID lookups and a string-keyed grid map for spatial indexing.
- **Complexity**:
  - Lookup by ID: O(1) average
  - Location Update: O(1) (updates HashMap and moves driver between grid cells)
  - Spatial Search: O(D) where D is the number of drivers in the search radius.
- **Trade-off**: Grid indexing significantly reduces the search space compared to a full linear scan of all drivers, especially in dense areas.

### 3. Active Ride Tracker (Doubly Linked List)
- **Choice**: Doubly Linked List with an auxiliary HashMap for index.
- **Complexity**:
  - Add: O(1)
  - Remove by ID: O(1)
  - List all: O(N)
- **Trade-off**: O(1) removal is critical for high-frequency "complete ride" operations where we have the ride ID.

## Benchmark Results

- **Min-Heap (Enqueue/Dequeue)**: ~102 ns/op
- **Spatial Search (FindNearest among 1000 drivers)**: ~71.5 µs/op

## CLI Command Reference

Once the application is running, you can use the following commands:

### Driver Management
- `register-driver <id> <name> <lat> <lng>`: Registers a new driver in the system.
    - **Example**: `register-driver D1 Alice 12.97 77.59`
    - **Effect**: Adds the driver to the HashMap and Grid Index. Sets status to `available`.
- `online <driverID>`: Toggles a driver's status to `available`.
- `offline <driverID>`: Toggles a driver's status to `offline`. Offline drivers are not considered for dispatch.
- `move <driverID> <lat> <lng>`: Simulates a driver moving to a new location.
    - **Effect**: Updates the Grid Index (moves driver between zones if necessary).

### Ride Management
- `request <riderID> <plat> <plng> <dlat> <dlng>`: Rider requests a ride from a Pickup (plat/plng) to a Dropoff (dlat/dlng).
    - **Example**: `request U1 12.97 77.59 12.90 77.50`
    - **Effect**: Enters the request into the Min-Heap Priority Queue (prioritized by request time).
- `dispatch`: Processes the ride queue.
    - **Logic**: Picks the oldest request and finds the nearest available driver within 5km. If found, assigns the driver and moves the ride to the Active Tracker (Linked List).
- `complete <rideID>`: Marks an active ride as completed.
    - **Effect**: Calculates fare, updates driver earnings, resets driver status to `available`, and moves driver to the dropoff location.

### Queries & Analytics
- `query-nearest <lat> <lng> <n>`: Finds the `n` nearest available drivers to the given coordinates.
- `query-earnings <driverID>`: Displays total earnings for the specified driver.
- `query-waits`: Calculates and displays the average wait time across all successfully dispatched rides.
- `query-zones`: Displays a list of zones and the number of requests originating from each (heat-map data).

---

## How to Run

1. **Build**: `go build -o rideshare cmd/main.go`
2. **Run Standard Interactive Shell**: `./rideshare`
3. **Run Individual Commands**: `./rideshare register-driver D1 Alice 10 10`
4. **Help**: `./rideshare --help` or `./rideshare help <command>`
5. **Exit Shell**: Type `exit` to close the interactive CLI.

### Command Format
The system now uses a standard subcommand format:
- `<subcommand> [flags] [args]`
- **Examples**:
  - `register-driver D1 Alice 12.97 77.59`
  - `request U1 12.97 77.59 12.90 77.50`
  - `dispatch`
  - `query-nearest 12.97 77.59 5`
