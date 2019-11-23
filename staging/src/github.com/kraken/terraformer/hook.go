package terraformer

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/plans"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

// Won't compile if LogHook can't be realized by a Hook
var _ Hook = (*LogHook)(nil)

// LogHook inherit from terraform.NilHook and implement terraform.Hook. It's
// required by the terraform.ContextOpts struct.
//
// Hook is the interface that must be implemented to hook into various
// parts of Terraform, allowing you to inspect or change behavior at runtime.
//
// There are MANY hook points into Terraform. If you only want to implement
// some hook points, but not all (which is the likely case), then embed the
// NilHook into your struct, which implements all of the interface but does
// nothing. Then, override only the functions you want to implement.
type LogHook struct {
	terraform.NilHook

	log       Logger
	l         sync.Mutex
	resources map[string]uiResourceState
	uiTimer   time.Duration
}

// uiResourceState tracks the state of a single resource
type uiResourceState struct {
	DispAddr       string
	IDKey, IDValue string
	Op             uiResourceOp
	Start          time.Time

	DoneCh chan struct{} // To be used for cancellation

	done chan struct{} // used to coordinate tests
}

// uiResourceOp is an enum for operations on a resource
type uiResourceOp byte

const (
	uiResourceUnknown uiResourceOp = iota
	uiResourceCreate
	uiResourceModify
	uiResourceDestroy
)

const defUITimer = 10 * time.Second
const maxIDLen = 80

// // SetLog assign a logger to the LogHook. A logger is what prints the output to
// // a LogWriter
// func (h *LogHook) SetLog(log *log.Logger)  {
//   log.Out = &LogWriter{
//     logger: log,
//   }
// }

// NewLogHook returns a new LogHook which is the default Hook
func NewLogHook(l Logger) *LogHook {
	if l == nil {
		l = NewStdLogger()
	}
	return &LogHook{
		log:       l,
		uiTimer:   defUITimer,
		resources: make(map[string]uiResourceState),
	}
}

// SetLog replaces the log of the LogHook
func (h *LogHook) SetLog(log Logger) {
	h.log = log
}

// PreApply is called before a single resource is applied.
func (h *LogHook) PreApply(addr addrs.AbsResourceInstance, gen states.Generation, action plans.Action, priorState, plannedNewState cty.Value) (terraform.HookAction, error) {
	dispAddr := addr.String()
	if gen != states.CurrentGen {
		dispAddr = fmt.Sprintf("%s (%s)", dispAddr, gen)
	}

	var operation string
	var op uiResourceOp
	idKey, idValue := format.ObjectValueIDOrName(priorState)
	switch action {
	case plans.Delete:
		operation = "Destroying..."
		op = uiResourceDestroy
	case plans.Create:
		operation = "Creating..."
		op = uiResourceCreate
	case plans.Update:
		operation = "Modifying..."
		op = uiResourceModify
	default:
		h.log.Infof("(Unknown action %s for %s)", action, dispAddr)
		return terraform.HookActionContinue, nil
	}

	attrBuf := new(bytes.Buffer)

	// Get all the attributes that are changing, and sort them. Also
	// determine the longest key so that we can align them all.
	keyLen := 0

	dAttrs := map[string]terraform.ResourceAttrDiff{}
	keys := make([]string, 0, len(dAttrs))
	for key := range dAttrs {
		keys = append(keys, key)
		if len(key) > keyLen {
			keyLen = len(key)
		}
	}
	sort.Strings(keys)

	// Go through and output each attribute
	for _, attrK := range keys {
		attrDiff := dAttrs[attrK]

		v := attrDiff.New
		u := attrDiff.Old
		if attrDiff.NewComputed {
			v = "<computed>"
		}

		if attrDiff.Sensitive {
			u = "<sensitive>"
			v = "<sensitive>"
		}

		attrBuf.WriteString(fmt.Sprintf("  %s:%s %#v => %#v\n", attrK, strings.Repeat(" ", keyLen-len(attrK)), u, v))
	}

	attrString := strings.TrimSpace(attrBuf.String())
	if attrString != "" {
		attrString = "\n  " + attrString
	}

	var stateIDSuffix string
	if idKey != "" && idValue != "" {
		stateIDSuffix = fmt.Sprintf(" [%s=%s]", idKey, idValue)
	} else {
		// Make sure they are both empty so we can deal with this more
		// easily in the other hook methods.
		idKey = ""
		idValue = ""
	}

	h.log.Infof("%s: %s%s %s", dispAddr, operation, stateIDSuffix, attrString)

	key := addr.String()
	uiState := uiResourceState{
		DispAddr: key,
		IDKey:    idKey,
		IDValue:  idValue,
		Op:       op,
		Start:    time.Now().Round(time.Second),
		DoneCh:   make(chan struct{}),
		done:     make(chan struct{}),
	}

	h.l.Lock()
	h.resources[key] = uiState
	h.l.Unlock()

	// Start goroutine that shows progress
	go h.stillApplying(uiState)

	return terraform.HookActionContinue, nil
}

// stillApplying shows the apply progress for every single resource
func (h *LogHook) stillApplying(state uiResourceState) {
	defer close(state.done)
	for {
		select {
		case <-state.DoneCh:
			return

		case <-time.After(h.uiTimer):
			// Timer up, show status
		}

		var msg string
		switch state.Op {
		case uiResourceModify:
			msg = "Still modifying..."
		case uiResourceDestroy:
			msg = "Still destroying..."
		case uiResourceCreate:
			msg = "Still creating..."
		case uiResourceUnknown:
			return
		}

		idSuffix := ""
		if state.IDKey != "" {
			idSuffix = fmt.Sprintf("%s=%s, ", state.IDKey, truncateStr(state.IDValue, maxIDLen))
		}

		h.log.Infof("%s: %s (%s%s elapsed)", state.DispAddr, msg, idSuffix, time.Now().Round(time.Second).Sub(state.Start))
	}
}

// PostApply is called after a single resource is applied. The error argument in
// PostApply is the error, if any, that was returned from the provider Apply
// call itself.
func (h *LogHook) PostApply(addr addrs.AbsResourceInstance, gen states.Generation, newState cty.Value, applyerr error) (terraform.HookAction, error) {

	id := addr.String()

	h.l.Lock()
	state := h.resources[id]
	if state.DoneCh != nil {
		close(state.DoneCh)
	}

	delete(h.resources, id)
	h.l.Unlock()

	var stateIDSuffix string
	if k, v := format.ObjectValueID(newState); k != "" && v != "" {
		stateIDSuffix = fmt.Sprintf(" [%s=%s]", k, truncateStr(v, maxIDLen))
	}

	var msg string
	switch state.Op {
	case uiResourceModify:
		msg = "Modifications complete"
	case uiResourceDestroy:
		msg = "Destruction complete"
	case uiResourceCreate:
		msg = "Creation complete"
	case uiResourceUnknown:
		return terraform.HookActionContinue, nil
	}

	if applyerr != nil {
		// Errors are collected and printed in ApplyCommand, no need to duplicate
		return terraform.HookActionContinue, nil
	}

	h.log.Infof("%s: %s after %s%s", addr, msg, time.Now().Round(time.Second).Sub(state.Start), stateIDSuffix)

	return terraform.HookActionContinue, nil
}

// PreDiff and PostDiff are called before and after a single resource
// resource is diffed.
func (h *LogHook) PreDiff(addr addrs.AbsResourceInstance, gen states.Generation, priorState, proposedNewState cty.Value) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

// PreProvisionInstanceStep should be self-explanatory
func (h *LogHook) PreProvisionInstanceStep(addr addrs.AbsResourceInstance, typeName string) (terraform.HookAction, error) {
	h.log.Infof("%s: Provisioning with '%s'...", addr, typeName)
	return terraform.HookActionContinue, nil
}

// PreProvision should be self-explanatory
// func (h *LogHook) PreProvision(n *terraform.InstanceInfo, provID string) (terraform.HookAction, error) {
// 	id := n.HumanId()
// 	h.log.Infof("%s: Provisioning with '%s'...", id, provID)
// 	return terraform.HookActionContinue, nil
// }

// ProvisionOutput is called with output sent back by the provisioners. This
// will be called multiple times as output comes in, but each call should
// represent a line of output. The ProvisionOutput method cannot control whether
// the hook continues running.
func (h *LogHook) ProvisionOutput(addr addrs.AbsResourceInstance, typeName string, msg string) {
	var buf bytes.Buffer

	prefix := fmt.Sprintf("%s (%s): ", addr, typeName)
	s := bufio.NewScanner(strings.NewReader(msg))
	s.Split(scanLines)
	for s.Scan() {
		line := strings.TrimRightFunc(s.Text(), unicode.IsSpace)
		if line != "" {
			buf.WriteString(fmt.Sprintf("%s%s\n", prefix, line))
		}
	}

	h.log.Printf("> %s", strings.TrimSpace(buf.String()))
}

// PreRefresh is called before a single resource state is refreshed.
func (h *LogHook) PreRefresh(addr addrs.AbsResourceInstance, gen states.Generation, priorState cty.Value) (terraform.HookAction, error) {
	var stateIDSuffix string
	if k, v := format.ObjectValueID(priorState); k != "" && v != "" {
		stateIDSuffix = fmt.Sprintf(" [%s=%s]", k, truncateStr(v, maxIDLen))
	}

	h.log.Infof("%s: Refreshing state...%s", addr, stateIDSuffix)
	return terraform.HookActionContinue, nil
}

// PreImportState is called before a single resource's state is being improted.
func (h *LogHook) PreImportState(addr addrs.AbsResourceInstance, importID string) (terraform.HookAction, error) {
	h.log.Infof("%s: Importing from ID %q...", addr, importID)
	return terraform.HookActionContinue, nil
}

// PostImportState is called after a single resource's state is being improted.
func (h *LogHook) PostImportState(addr addrs.AbsResourceInstance, imported []providers.ImportedResource) (terraform.HookAction, error) {
	h.log.Infof("%s: Import prepared!", addr)
	for _, s := range imported {
		h.log.Infof("  Prepared %s for import", s.TypeName)
	}

	return terraform.HookActionContinue, nil
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

// scanLines is basically copied from the Go standard library except
// we've modified it to also fine `\r`.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
