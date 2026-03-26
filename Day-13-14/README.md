# High-Performance Auction System: Scalability Report

A high-performance bidding engine designed for extreme write-throughput and hierarchical category browsing, optimized through extensive profiling and data-structure selection.

## 🏛 Architecture Overview
The system employs a decentralized synchronization model with fine-grained locking and specialized data structures to manage high-volume concurrent auctions.

```mermaid
graph TD
    AM[Auction Manager] --> MapItems{map[ID]ItemContext}
    MapItems --> IC[ItemContext]
    IC --> Heap[Max-Heap: Bids]
    IC --> LL[Linked List: History]
    IC --> ItemStruct[models.Item]
    ItemStruct --> Current[CurrentBid]
    
    AM --> UM[User Manager]
    UM --> UserMap{map[ID]User}
    UserMap --> User[UserStruct]
    User --> Balance[Balance]
    User --> UndoStack[Undo Stack: Bids]

    AM --> CT[Category Tree: Trie]
    CT --> Node[TrieNode]
    Node --> Child[Children]
```

## 🛠 Tech Stack & Implementation Analysis
Each data structure was selected to minimize time complexity for critical path operations.

| Data Structure | Use Case | Time Complexity | Trade-off Justification |
| :--- | :--- | :--- | :--- |
| **Max-Heap** | Bidding Leaderboard | O(log N) Build | Allows O(1) extraction of current winner and efficient leader tracking during high-concurrency bidding. |
| **Linked List** | Item Bid History | O(1) Append | Superior for "history" views where new bids are prepended. Facilitates O(1) retraction of the most recent bid. |
| **Trie** | Category Hierarchy | O(L) Search | Optimal for recursive category browsing (e.g., searching "Electronics" returns "Phones" and "Laptops" efficiently). |
| **Undo Stack** | Bid Retraction | O(1) Push/Pop | Prevents global traversal. Each user maintains their own chronological bid stack for immediate retraction. |
| **Concurrent Map**| Item Registry | O(1) Lookup | Fine-grained per-item locking (`sync.Mutex` inside the value) prevents the "stop-the-world" effect of a global lock. |

## 🚀 Benchmark Results: Throughput
We executed a stress test simulating 100 users bidding on 10 items across 8 CPU cores.

- **Peak Throughput**: **~1,860,000 Bids/Second**
- **Average Latency**: **~535 ns/op** per bid placement

### Optimization Notes
Original profiling showed **mutex contention** as the primary bottleneck (74% delay in `sync.Unlock`). By shrinking the critical section in `PlaceBid` and releasing the lock before side-effects (like SSE broadcasting), we achieved a **28x throughput improvement**.

## 🔌 API Documentation

### Users
- `POST /users`: Register a new user with initial balance.

### Auctions
- `POST /items`: Register a new item for auction.
- `GET /items`: List items (Optional filter: `?category=Phones`).
- `GET /items/{id}`: Fetch detailed item state and bid history.
- `GET /stats`: Dashboard for global auction metrics (Active Items, Revenue, Total Bids).

### Bidding
- `POST /items/{id}/bid`: Place a bid. Includes validation of balance, start time, and minimum increment.
- `DELETE /items/{id}/bid/last`: Retract your most recent bid for an item.
- `POST /items/{id}/end`: Finalize auction and retrieve winner.
- `GET /items/{id}/live`: **SSE Endpoint** for real-time bid updates.

## 📊 Profiling Visuals
Check the codebase for generated profile visual artifacts:
- [CPU Flame Graph](./cpu_opt_flamegraph.svg)
- [Mutex Delay Analysis](./mutex_opt_top.txt)
