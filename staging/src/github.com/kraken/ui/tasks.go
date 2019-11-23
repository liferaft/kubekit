package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	defaultPeriodicUITimer = 10 * time.Second
	defaultRenderTimer     = 5 * time.Second
)

// Tasks is a set of Task's
type Tasks struct {
	tasks        map[string]map[string]*Task
	l            sync.Mutex
	DoneCh       chan struct{}
	order        []string
	printedTasks int
}

// NewTasks creates a Tasks object with a collection of tasks executed and the
// UI to report them
func NewTasks() *Tasks {
	return &Tasks{
		tasks: make(map[string]map[string]*Task),
		order: make([]string, 0),
	}
}

// Copy makes a deep copy of this Tasks
func (ct *Tasks) Copy() *Tasks {
	tasks := make(map[string]map[string]*Task)
	for k, v := range ct.tasks {
		tasks[k] = make(map[string]*Task)
		for tk, tv := range v {
			tasks[k][tk] = &Task{
				Name:       tv.Name,
				Node:       tv.Node,
				StartTime:  tv.StartTime,
				EndTime:    tv.EndTime,
				Action:     tv.Action,
				LastStatus: tv.LastStatus,
			}
		}
	}

	order := make([]string, len(ct.order))
	copy(order, ct.order)

	return &Tasks{
		tasks: tasks,
		order: order,
	}
}

// NewTask creates a new task
func (ct *Tasks) NewTask(node, name string) *Task {
	task := &Task{
		Name: strings.Replace(name, "/", ".", -1),
		Node: node,
	}

	ct.add(node, name, task)

	return task
}

// Start makes a configuration tasks to start reporting status
func (ct *Tasks) Start(node, name string, action TaskAction) *Task {
	task := ct.NewTask(node, name)

	task.StartTime = time.Now().Round(time.Second)
	task.DoneCh = make(chan struct{})
	task.Action = action

	return task
}

// Task return the task with the given node and name
func (ct *Tasks) Task(node, name string) *Task {
	var ts map[string]*Task
	var t *Task
	var ok bool

	// ct.l.Lock()
	// defer ct.l.Unlock()

	if ts, ok = ct.tasks[node]; !ok {
		return nil
	}
	if t, ok = ts[name]; !ok {
		return nil
	}
	return t
}

func (ct *Tasks) add(node, name string, task *Task) {
	ct.l.Lock()
	if _, ok := ct.tasks[node]; !ok {
		ct.tasks[node] = make(map[string]*Task)
	}
	if _, ok := ct.tasks[node][name]; !ok {
		ct.order = append(ct.order, node+":"+name)
	}
	ct.tasks[node][name] = task

	ct.l.Unlock()
}

func (ct *Tasks) delete(node, name string) {
	ct.l.Lock()
	delete(ct.tasks[node], name)
	if len(ct.tasks[node]) == 0 {
		delete(ct.tasks, node)
	}
	ct.l.Unlock()
}

// Complete makes a configuration task to finish the status report
func (ct *Tasks) Complete(node, name string) *Task {
	task := ct.Task(node, name)
	if task == nil {
		return nil
	}

	task.Done()

	return task
	// ct.delete(node, name)
}

// CompleteGroup makes all the configuration tasks of the given node to finish the
// status report
func (ct *Tasks) CompleteGroup(node string) []*Task {
	tasks := []*Task{}

	var ts map[string]*Task
	var ok bool

	if ts, ok = ct.tasks[node]; !ok {
		return tasks
	}

	for _, task := range ts {
		if closed := task.Done(); closed {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// CompleteAll makes all the configuration tasks to finish the status report
func (ct *Tasks) CompleteAll() []*Task {
	tasks := []*Task{}

	for node := range ct.tasks {
		if t := ct.CompleteGroup(node); t != nil {
			tasks = append(tasks, t...)
		}
	}

	return tasks
}

// Render print out all the tasks without scrolling
func (ct *Tasks) Render() {
	ct.l.Lock()
	defer ct.l.Unlock()

	if ct.printedTasks != 0 {
		fmt.Printf("\033[%dA", ct.printedTasks)
	}
	ct.printedTasks = len(ct.order)

	for _, key := range ct.order {
		location, resource := split(key)
		task := ct.Task(location, resource)
		if task == nil {
			ct.printedTasks--
			continue
		}
		fmt.Println("\033[2K" + task.String())
	}

}

func split(key string) (location string, resource string) {
	keys := strings.Split(key, ":")
	if len(keys) > 1 {
		location = keys[0]
		resource = strings.Join(keys[1:], ":")
	} else {
		resource = key
	}
	return
}
