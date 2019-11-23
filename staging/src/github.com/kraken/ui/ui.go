package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/johandry/log"
)

// ANSI colors
const (
	Reset     string = "\033[0m"
	BlackBold string = "\033[1;30m"
	Red       string = "\033[31m"
	Green     string = "\033[32m"
	Yellow    string = "\033[33m"
)

type action struct {
	location string
	name     string
	state    string
}

// UI handles the interactivity with the user
type UI struct {
	Tasks  *Tasks
	Log    *log.Logger
	scroll bool
	lines  int
	order  []string
	// Out *Output
	// cluster  string
}

// New creates a new UI
func New(scroll bool, logger *log.Logger) *UI {

	ui := UI{
		Tasks:  NewTasks(),
		Log:    logger,
		scroll: scroll,
	}
	if !scroll {

	}

	return &ui
}

// Copy makes a deep copy of this UI
func (ui *UI) Copy() *UI {
	tasks := ui.Tasks.Copy()

	logger := ui.Log.Copy()

	order := make([]string, len(ui.order))
	copy(order, ui.order)

	return &UI{
		Tasks:  tasks,
		Log:    logger,
		scroll: ui.scroll,
		lines:  ui.lines,
		order:  order,
	}
}

// SetLogPrefix changes the prefix log
func (ui *UI) SetLogPrefix(prefix string) {
	ui.Log.SetPrefix(prefix)
}

// Task return the task with the given node and name
func (ui *UI) Task(node, name string) *Task {
	return ui.Tasks.Task(node, name)
}

// Print prints the output to stdOut and to a logger
func (ui *UI) Print(task *Task, stdOut, logOut string) {
	resourceTitleLog := fmt.Sprintf("%s%s", Reset, BlackBold)

	if len(task.Name) != 0 {
		resourceTitleLog = fmt.Sprintf("%s%s: ", resourceTitleLog, task.Name)
	}

	// Log Output:
	if len(logOut) != 0 {
		ui.Log.Infof("%s%s", resourceTitleLog, logOut)
	}

	// CLI Output: Print only if log output is a file, not Stderr or Stdout
	if len(stdOut) == 0 {
		return
	}
	task.LastStatus = stdOut
	if !isRegular(ui.Log.Out) {
		return
	}

	if ui.scroll {
		fmt.Println(task.String())
		return
	}

	ui.Tasks.Render()
	// ui.output.Update(location, resource, strings.ToUpper(stdOut[:1])+stdOut[1:])
}

func isRegular(out io.Writer) bool {
	file := out.(*os.File)
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return stat.Mode().IsRegular()
}

// Notify prints to stdout and send to logs the given task
func (ui *UI) Notify(location, resource, description, details string, action ...TaskAction) {
	// Finish a started task
	if strings.Contains(description, "</") {
		task := ui.Tasks.Complete(location, resource)
		if task == nil {
			return
		}
		description := fmt.Sprintf("%s after %s", task.Action.Stage(Complete), task.EndTime.Sub(task.StartTime))
		details = fmt.Sprintf("%s%s%s", description, details, Reset)
		ui.Print(task, description, details)
		return
	}

	// Handle starting tasks
	if strings.Contains(description, "<") && len(action) != 0 {
		task := ui.Tasks.Start(location, resource, action[0])

		details = fmt.Sprintf("%s%s%s", task.Action.Stage(Start), details, Reset)
		ui.Print(task, task.Action.Stage(Start), details)
		go task.stillWorking(ui)
		return
	}

	if len(details) == 0 {
		details = description
	}

	// Handle regular or one time tasks
	task := ui.Tasks.Complete(location, resource)
	if task == nil {
		task = ui.Tasks.NewTask(location, resource)
	}
	ui.Print(task, description, details)
}

// TerminateNotificationFor terminate all the notifications for a given location
// Usefull when there is an error and the notification are still reporting
// incorrect status
func (ui *UI) TerminateNotificationFor(location, details string) {
	tasks := ui.Tasks.CompleteGroup(location)
	ui.terminate(details, tasks)
}

// TerminateAllNotifications terminate all the existing notifications in this UI
// Usefull when there is an error and the notification are still reporting
// incorrect status
func (ui *UI) TerminateAllNotifications(details string) {
	tasks := ui.Tasks.CompleteAll()
	ui.terminate(details, tasks)
}

func (ui *UI) terminate(details string, tasks []*Task) {
	for _, task := range tasks {
		description := fmt.Sprintf("Terminating notification after %s", task.EndTime.Sub(task.StartTime))
		details = fmt.Sprintf("%s%s%s", description, details, Reset)
		ui.Print(task, description, details)
	}
}
