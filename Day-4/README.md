# Router Packet Priority System

A high-performance Go-based router simulation comparing Linked List and Slice implementations for priority-based packet queuing.

## Benchmark Findings

| Implementation | 100 Packets | 1,000 Packets | 10,000 Packets | 100,000 Packets |
|----------------|-------------|---------------|----------------|-----------------|
| **Linked List** | 15,011 ns   | 185,251 ns    | 1,878,119 ns   | 18,250,099 ns   |
| **Slices**      | 12,898 ns   | 159,222 ns    | 2,030,993 ns   | 18,015,857 ns   |

### Why Slices are generally faster for larger values?

As we scale to larger numbers of packets (100K+), Slices tend to outperform Linked Lists for several architectural reasons:

1.  **Cache Locality**: Slices use contiguous memory. When the CPU fetches a packet pointer from a slice, it likely fetches adjacent pointers into the cache simultaneously. Linked list nodes are scattered in memory, causing frequent "cache misses" and forcing the CPU to wait for RAM.
2.  **Allocation Overhead**: The Linked List implementation requires a new `Node` allocation for every single `Enqueue`. This results in roughly **33% more allocations** compared to Slices (which grow their capacity geometrically).
3.  **Pointer Chasing**: To traverse or manage a linked list, the CPU must follow (`dereference`) multiple pointers (`next`, `prev`). Each jump is overhead that Slices avoid by using simple index offsets.
4.  **Garbage Collection**: More objects (Nodes) mean more work for the Go Garbage Collector. Slices keep data in a single large block, making GC sweeps faster.
