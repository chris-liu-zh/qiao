package qiao

import (
	"encoding/json"
	"fmt"
	"sync"
)

type WG struct {
	Wg        sync.WaitGroup
	MaxTask   int
	TaskCount int
	Tasks     chan Task
	Results   chan Result
}

type Task struct {
	Id     int
	F      func(v ...any) error
	Params []any
}

type Result struct {
	TaskId   int
	WorkId   int
	Success  bool
	TaskTime int64
	Message  string
}

func (w *WG) worker(id int) {
	defer w.Wg.Done()
	for task := range w.Tasks {
		var message string
		success := true
		s := TimeStamp()
		if err := task.F(task.Params...); err != nil {
			message = err.Error()
			success = false
		}
		t := TimeStamp()
		w.Results <- Result{TaskId: task.Id, WorkId: id, TaskTime: (t - s) / 1000, Success: success, Message: message}
		w.TaskCount--
		fmt.Println("任务ID：", task.Id, "已完成，耗时：", (t-s)/1000, "秒")
		fmt.Println("当前任务数：", w.TaskCount)
	}
}

func (w *WG) AddTask(id int, f func(...any) error, params ...any) {
	w.Tasks <- Task{Id: id, F: f, Params: params}
	w.TaskCount++
	fmt.Println("当前任务数：", w.TaskCount)
	fmt.Println("添加任务ID：", id)
}

func (w *WG) Result() {
	for result := range w.Results {
		r, _ := json.Marshal(result)
		fmt.Println(string(r))
	}
}

func TaskStart(MaxTask int) *WG {
	w := &WG{}
	w.MaxTask = MaxTask
	w.Wg = sync.WaitGroup{}
	w.Tasks = make(chan Task, MaxTask)
	w.Results = make(chan Result, MaxTask)
	w.Wg.Add(MaxTask)

	for i := 1; i <= MaxTask; i++ {
		go w.worker(i)
	}
	return w
}
