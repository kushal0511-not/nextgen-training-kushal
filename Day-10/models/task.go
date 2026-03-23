package models

import "time"

type TaskStatus string

const (
	TaskStatusReady     TaskStatus = "READY"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusCompleted TaskStatus = "COMPLETED"
	TaskStatusStarved   TaskStatus = "STARVED"
)

type Task struct {
	PID                int
	Name               string
	Priority           int
	CPUBurst           time.Duration
	Deadline           time.Time
	ArrivalTime        time.Time
	WaitTime           time.Duration
	TurnAroundTime     time.Duration
	Status             TaskStatus
	FirstScheduledTime time.Time
	CompletionTime     time.Time
}
