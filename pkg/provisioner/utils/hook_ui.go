package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/plans"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/terraform"
	"github.com/kraken/ui"
	"github.com/zclconf/go-cty/cty"
)

// Won't compile if UIHook can't be realized by a Hook
var _ terraform.Hook = (*UIHook)(nil)

// Hook is the interface that must be implemented to hook into various
// parts of Terraform, allowing you to inspect or change behavior at runtime.
//
// There are MANY hook points into Terraform. If you only want to implement
// some hook points, but not all (which is the likely case), then embed the
// NilHook into your struct, which implements all of the interface but does
// nothing. Then, override only the functions you want to implement.

// UIHook inherit from terraform.NilHook and implement terraform.Hook which is
// the interface that must be implemented to hook into various parts of
// Terraform, allowing you to inspect or change behavior at runtime.
type UIHook struct {
	terraform.NilHook
	platform string
	l        sync.Mutex
	once     sync.Once
	ui       *ui.UI
}

// NewUIHook returns a new LogHook which is the default Hook
func NewUIHook(platform string, ui *ui.UI) *UIHook {
	hook := &UIHook{
		platform: platform,
		ui:       ui,
	}

	return hook
}

// // PreApply and PostApply are called before and after a single
// // resource is applied. The error argument in PostApply is the
// // error, if any, that was returned from the provider Apply call itself.
// PreApply(*InstanceInfo, *InstanceState, *InstanceDiff) (HookAction, error)
// PostApply(*InstanceInfo, *InstanceState, error) (HookAction, error)

// PreApply is called before a single resource is applied.
func (h *UIHook) PreApply(addr addrs.AbsResourceInstance, gen states.Generation, action plans.Action, priorState, plannedNewState cty.Value) (terraform.HookAction, error) {
	// dispAddr := addr.String()
	// if gen != states.CurrentGen {
	// 	dispAddr = fmt.Sprintf("%s (%s)", dispAddr, gen)
	// }
	idKey, idValue := format.ObjectValueIDOrName(priorState)

	var uiOp ui.TaskAction
	switch action {
	case plans.Delete:
		uiOp = ui.Destroy
	case plans.Create:
		uiOp = ui.Create
	case plans.Update:
		uiOp = ui.Modify
	default:
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

	// attrString and stateIDSuffix are only used in the logs, not in the CLI
	attrString := strings.TrimSpace(attrBuf.String())
	if attrString != "" {
		attrString = "\n  " + attrString
	}

	var stateIDSuffix string
	if idKey != "" && idValue != "" {
		stateIDSuffix = fmt.Sprintf(" [%s=%s]", idKey, idValue)
	}

	h.ui.Notify(
		h.platform,
		addr.String(),
		"<"+addr.String()+">",
		fmt.Sprintf("%s%s%s", stateIDSuffix, ui.Reset, attrString),
		uiOp,
	)

	return terraform.HookActionContinue, nil
}

// PostApply is called after a single resource is applied. The error argument in
// PostApply is the error, if any, that was returned from the provIDer Apply
// call itself.
func (h *UIHook) PostApply(addr addrs.AbsResourceInstance, gen states.Generation, newState cty.Value, applyerr error) (terraform.HookAction, error) {
	// dispAddr := addr.String()
	// if gen != states.CurrentGen {
	// 	dispAddr = fmt.Sprintf("%s (%s)", dispAddr, gen)
	// }

	var stateIDSuffix string
	if k, v := format.ObjectValueID(newState); k != "" && v != "" {
		stateIDSuffix = fmt.Sprintf(" [%s=%s]", k, v)
	}

	t := h.ui.Task(h.platform, addr.String())

	var uiOp ui.TaskAction
	switch t.Action {
	case ui.Modify:
		uiOp = ui.Modify
	case ui.Destroy:
		uiOp = ui.Destroy
	case ui.Create:
		uiOp = ui.Create
	default:
		return terraform.HookActionContinue, nil
	}

	if applyerr != nil {
		// Errors are collected and printed in ApplyCommand, no need to duplicate
		return terraform.HookActionContinue, nil
	}

	h.ui.Notify(
		h.platform,
		addr.String(),
		"</"+addr.String()+">",
		stateIDSuffix,
		uiOp,
	)

	return terraform.HookActionContinue, nil
}

// // PreDiff and PostDiff are called before and after a single resource
// // resource is diffed.
// PreDiff(*InstanceInfo, *InstanceState) (HookAction, error)
// PostDiff(*InstanceInfo, *InstanceDiff) (HookAction, error)

// // Provisioning hooks
// //
// // All should be self-explanatory. ProvisionOutput is called with
// // output sent back by the provisioners. This will be called multiple
// // times as output comes in, but each call should represent a line of
// // output. The ProvisionOutput method cannot control whether the
// // hook continues running.
// PreProvisionResource(*InstanceInfo, *InstanceState) (HookAction, error)
// PostProvisionResource(*InstanceInfo, *InstanceState) (HookAction, error)
// PreProvision(*InstanceInfo, string) (HookAction, error)
// PostProvision(*InstanceInfo, string, error) (HookAction, error)
// ProvisionOutput(*InstanceInfo, string, string)

// PreProvisionInstanceStep should be self-explanatory
func (h *UIHook) PreProvisionInstanceStep(addr addrs.AbsResourceInstance, typeName string) (terraform.HookAction, error) {
	// dispAddr := addr.String()
	// if gen != states.CurrentGen {
	// 	dispAddr = fmt.Sprintf("%s (%s)", dispAddr, gen)
	// }

	msgLog := fmt.Sprintf("Provisioning with '%s'...%s", typeName, ui.Reset)
	h.ui.Notify(h.platform, addr.String(), msgLog, msgLog)

	return terraform.HookActionContinue, nil
}

// ProvisionOutput is called with output sent back by the provisioners. This
// will be called multiple times as output comes in, but each call should
// represent a line of output. The ProvisionOutput method cannot control whether
// the hook continues running.
func (h *UIHook) ProvisionOutput(addr addrs.AbsResourceInstance, typeName string, msg string) {
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

	// There is no h.ui.Notify() here, it only prints to the logs and it's different
	h.ui.Log.Infof(strings.TrimSpace(buf.String()))
}

// // PreRefresh and PostRefresh are called before and after a single
// // resource state is refreshed, respectively.
// PreRefresh(*InstanceInfo, *InstanceState) (HookAction, error)
// PostRefresh(*InstanceInfo, *InstanceState) (HookAction, error)

// PreRefresh is called before a single resource state is refreshed.
func (h *UIHook) PreRefresh(addr addrs.AbsResourceInstance, gen states.Generation, priorState cty.Value) (terraform.HookAction, error) {
	var stateIDSuffix string
	if k, v := format.ObjectValueID(priorState); k != "" && v != "" {
		stateIDSuffix = fmt.Sprintf(" [%s=%s]", k, v)
	}

	h.ui.Notify(
		h.platform,
		addr.String(),
		fmt.Sprintf("Refreshing state..."),
		fmt.Sprintf("Refreshing state...%s", stateIDSuffix),
	)
	return terraform.HookActionContinue, nil
}

// // PostStateUpdate is called after the state is updated.
// PostStateUpdate(*State) (HookAction, error)

// // PreImportState and PostImportState are called before and after
// // a single resource's state is being improted.
// PreImportState(*InstanceInfo, string) (HookAction, error)
// PostImportState(*InstanceInfo, []*InstanceState) (HookAction, error)

// PreImportState is called before a single resource's state is being improted.
func (h *UIHook) PreImportState(addr addrs.AbsResourceInstance, importID string) (terraform.HookAction, error) {
	msgLog := fmt.Sprintf("Importing from ID %q...", importID)
	h.ui.Notify(h.platform, addr.String(), msgLog, msgLog)

	return terraform.HookActionContinue, nil
}

// PostImportState is called after a single resource's state is being improted.
func (h *UIHook) PostImportState(addr addrs.AbsResourceInstance, imported []providers.ImportedResource) (terraform.HookAction, error) {
	msgLog := "Import complete!"
	h.ui.Notify(h.platform, addr.String(), msgLog, msgLog)
	for _, s := range imported {
		msgLog := fmt.Sprintf("%s  Prepared %s for import", ui.Green, s.TypeName)
		h.ui.Notify(h.platform, "", msgLog, msgLog)
	}

	return terraform.HookActionContinue, nil
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

// func truncateID(id string, maxLen int) string {
// 	totalLength := len(id)
// 	if totalLength <= maxLen {
// 		return id
// 	}
// 	if maxLen < 5 {
// 		// We don't shorten to less than 5 chars
// 		// as that would be pointless with ... (3 chars)
// 		maxLen = 5
// 	}

// 	dots := "..."
// 	partLen := maxLen / 2

// 	leftIdx := partLen - 1
// 	leftPart := id[0:leftIdx]

// 	rightIdx := totalLength - partLen - 1

// 	overlap := maxLen - (partLen*2 + len(dots))
// 	if overlap < 0 {
// 		rightIdx -= overlap
// 	}

// 	rightPart := id[rightIdx:]

// 	return leftPart + dots + rightPart
// }
