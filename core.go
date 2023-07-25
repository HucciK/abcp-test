package main

import "time"

const (
	SuccessfulTaskStatus = "task has been successed"
	FailureTaskStatus    = "something went wrong"
)

type Task struct {
	id         string
	createdAt  time.Time // время создания
	finishedAt time.Time // время выполнения
	result     string
	err        error
}
