package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// TaskAction is a possible action that a task can do
type TaskAction int

// All the possible actions a task can do
const (
	Regular TaskAction = iota
	Configure
	Upload
	Create
	Setup
	Modify
	Destroy
)

// Stage are the different stages of a task action
type stage int

// All the stages of an task action
const (
	Start stage = iota
	Working
	Complete
	Fail
)

// Stage returns the text of an action at the given stage
func (a TaskAction) Stage(stage stage) string {
	switch a {
	case Regular:
		switch stage {
		case Start:
			return "Starting..."
		case Working:
			return "Still working..."
		case Complete:
			return "Complete"
		case Fail:
			return "Fail"
		}
	case Configure:
		switch stage {
		case Start:
			return "Configuring..."
		case Working:
			return "Still configuring..."
		case Complete:
			return "Configuration complete"
		case Fail:
			return "Configuration fail"
		}
	case Upload:
		switch stage {
		case Start:
			return "Uploading..."
		case Working:
			return "Still uploading..."
		case Complete:
			return "Upload complete"
		case Fail:
			return "Upload fail"
		}
	case Create:
		switch stage {
		case Start:
			return "Creating..."
		case Working:
			return "Still creating..."
		case Complete:
			return "Creation complete"
		case Fail:
			return "Creation fail"
		}
	case Setup:
		switch stage {
		case Start:
			return "Setting up..."
		case Working:
			return "Still setting up..."
		case Complete:
			return "Setup complete"
		case Fail:
			return "Setuo fail"
		}
	case Modify:
		switch stage {
		case Start:
			return "Modifying..."
		case Working:
			return "Still modifying..."
		case Complete:
			return "Modifications complete"
		case Fail:
			return "Modifications fail"
		}
	case Destroy:
		switch stage {
		case Start:
			return "Destroying..."
		case Working:
			return "Still destroying..."
		case Complete:
			return "Destruction complete"
		case Fail:
			return "Destruction fail"
		}
	}

	panic(fmt.Errorf("unknown action %v and stage %v", a, stage))
}

// Task tracks the status of a single configuration task
type Task struct {
	Name       string
	Node       string
	StartTime  time.Time
	EndTime    time.Time
	l          sync.Mutex
	DoneCh     chan struct{}
	Action     TaskAction
	LastStatus string
}

// Done close safely the Done Channel
func (t *Task) Done() bool {
	var closed bool

	t.l.Lock()
	defer t.l.Unlock()

	if t.DoneCh != nil && t.EndTime.IsZero() {
		close(t.DoneCh)
		t.EndTime = time.Now().Round(time.Second)
		closed = true
	}
	return closed
}

func truncateStr(str string, maxLen int) string {
	totalLength := len(str)
	if totalLength <= maxLen {
		return str
	}
	if maxLen < 5 {
		// We don't shorten to less than 5 chars
		// as that would be pointless with ... (3 chars)
		maxLen = 5
	}

	dots := "..."
	partLen := maxLen / 2

	leftStrx := partLen - 1
	leftPart := str[0:leftStrx]

	rightStrx := totalLength - partLen - 1

	overlap := maxLen - (partLen*2 + len(dots))
	if overlap < 0 {
		rightStrx -= overlap
	}

	rightPart := str[rightStrx:]

	return leftPart + dots + rightPart
}

func (t *Task) String() string {
	var location string
	locLen := len(t.Node)
	if locLen != 0 {
		// truncates to 10 characters, pads if needed, and left justifies the text
		location = fmt.Sprintf("[ %-10.10s ] ", truncateStr(t.Node, 10))
	}
	var status string
	if len(t.LastStatus) > 0 {
		status = strings.ToUpper(t.LastStatus[:1]) + strings.ToLower(t.LastStatus[1:])
	}
	return fmt.Sprintf("%s%s%s%s: %s%s%s", Reset, Green, location, t.Name, BlackBold, status, Reset)
}

func (t *Task) stillWorking(i *UI) {
	// When the task is terminated/completed the channel is closed. So, this defer fail
	// defer close(t.DoneCh)
	for {
		select {
		case <-t.DoneCh:
			return

		case <-time.After(defaultPeriodicUITimer):
			desc := fmt.Sprintf("%s (%s elapsed)", t.Action.Stage(Working), time.Now().Round(time.Second).Sub(t.StartTime))
			i.Print(t, desc, desc)
		}
	}
}
