package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

var (
	maxWorkers = 5
	timeout    = 1
)

func main() {

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)

	//Создаем канал в который будем отсылать новые таски и запускаем функцию генерации тасков в отдельной горутине
	taskCh := make(chan Task, 10)
	go taskCreator(ctx, taskCh)

	//Создаем канал в который будем отсылать таски обработанные воркерами
	processedTask := make(chan Task, maxWorkers)
	errTaskCh := make(chan Task, maxWorkers)
	doneTaskCh := make(chan Task, maxWorkers)
	go taskSorter(ctx, processedTask, doneTaskCh, errTaskCh)

	wg := &sync.WaitGroup{}
	for i := 1; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			taskWorker(ctx, wg, taskCh, processedTask)
		}()
	}

	go func() {
		wg.Wait()
		close(processedTask)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-doneTaskCh:
			if !ok {
				continue
			}
			fmt.Printf("DONE TASK: %s WITH RESULT: %s\n", t.id, t.result)
		case t, ok := <-errTaskCh:
			if !ok {
				continue
			}
			fmt.Printf("TASK FAILED: %s ERROR: %v", t.id, t.err)
		}
	}
}

var someErr = errors.New("some error occured")

func taskCreator(ctx context.Context, taskCh chan<- Task) {
	for {
		select {
		case <-ctx.Done():
			close(taskCh)
			return
		default:
			var task Task
			task.id = uuid.NewString()
			task.createdAt = time.Now()
			if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
				task.err = someErr
			}
			taskCh <- task // передаем таск на выполнение
		}
	}
}

func taskWorker(ctx context.Context, wg *sync.WaitGroup, taskCh <-chan Task, processedTask chan<- Task) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-taskCh:
			if t.createdAt.After(time.Now().Add(-20 * time.Second)) {
				t.result = SuccessfulTaskStatus
			} else {
				t.result = FailureTaskStatus
				t.err = someErr
			}
			t.finishedAt = time.Now()

			processedTask <- t
		}
	}
}

func taskSorter(ctx context.Context, processedTaskCh <-chan Task, doneTaskCh, errTaskCh chan<- Task) {
	for {
		select {
		case <-ctx.Done():
			close(doneTaskCh)
			close(errTaskCh)
			return
		case t := <-processedTaskCh:
			if t.err != nil {
				errTaskCh <- t
			} else {
				doneTaskCh <- t
			}
		}
	}
}
