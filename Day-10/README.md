# Day 10: Scheduling Algorithms Implementation

This project implements three fundamental CPU scheduling algorithms in Go, demonstrating priority management, starvation prevention, and multi-level feedback queue concepts.

## Implemented Algorithms

### 1. Priority Scheduling
- **Logic**: Always selects the task with the highest priority (lowest numerical value) from the ready queue.
- **Data Structure**: Uses a Min-Heap for $O(\log n)$ insertion and extraction of the highest priority task.
- **Problem**: Lower priority tasks may suffer from starvation if higher priority tasks keep arriving.

### 2. Priority Scheduling with Aging
- **Logic**: Prevents starvation by gradually increasing the priority of tasks that have been waiting in the queue for a long time.
- **Implementation**: A background goroutine runs a ticker every 500ms, which decrements the priority value (effectively increasing priority) of all waiting tasks.
- **Goal**: Ensures that every task eventually reaches the highest priority (1) and gets executed.

### 3. Round-Robin with Priority Classes
- **Logic**: Divides tasks into three priority classes, each with a different time quantum.
  - **High Priority** (Priority ≤ 3): 50ms quantum.
  - **Medium Priority** (Priority 4-7): 100ms quantum.
  - **Low Priority** (Priority > 7): 200ms quantum.
- **Pre-emption**: If a High-priority task arrives while a lower priority task is running, the running task is immediately pre-empted (interrupted) and returned to its queue, allowing the urgent task to run.

## Project Structure

```text
Day-10/
├── cmd/
│   └── main.go           # Simulation and verification entry point
├── internal/
│   ├── heap/
│   │   └── heap.go       # Generic Min-Heap implementation
│   ├── aging_scheduler.go
│   ├── priority_scheduler.go
│   ├── round_robin_scheduler.go
│   └── scheduler.go      # Shared interface definition
└── models/
    └── task.go           # Task data structures
```

## How to Run

To run the simulation and see all three schedulers in action:

```bash
cd Day-10/cmd
go run .
```

The simulation will output the execution order, quantum usage, and pre-emption events for each scheduler.
