package internal

import "github.com/nextgen-training-kushal/Day-10/models"

type Scheduler interface {
	AddTask(models.Task)
	Schedule()
	Shutdown()
}
