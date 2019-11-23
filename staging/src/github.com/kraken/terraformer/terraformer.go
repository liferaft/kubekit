package terraformer

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/backend/local"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/plans"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/provisioners"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/terraform/tfdiags"
	"github.com/terraform-providers/terraform-provider-null/null"
	"github.com/zclconf/go-cty/cty"
)

// Terraformer is a struct that works as the interface to Terraform but using
// the library instead of the binary. The input parameters are Code, Vars,
// providers and provisioners, these last two are set with the methods
// AddProvider and AddProvisioner. State is an input and output. If the
// infrastructure is not new (has a current state) such state needs to be set
// (input) either with State or the method LoadState. When the changes are
// applied the State is the output of the final infrastructure state. This final
// output state could be the input state for the next change. Other output are
// the logs of the changes, these can be manipulated/filtered with a UI Hook.
type Terraformer struct {
	Code         []byte
	Vars         Variables
	State        *State
	Hooks        []Hook
	Stats        *Stats
	plan         *Plan
	lw           *LogWriter
	providers    map[string]providers.Factory
	provisioners map[string]provisioners.Factory
	context      *terraform.Context
}

// State is an alias for terraform.State
type State = states.State

// Plan is an alias for terraform.Plan
type Plan = plans.Plan

// Variables is a map of variables and their values
type Variables map[string]interface{}

// Hook is an alias for terraform.Hook
type Hook = terraform.Hook

// New returns an initialized Terraformer instance
func New(logger Logger, hooks ...terraform.Hook) (*Terraformer, error) {
	// initiate it with an empty state. It can be replaced with SetState
	// state := terraform.NewState()
	// According to NewContext() code, if the state is nill it will be set as a new state.

	// set default providers, if more are needed adds them with AddProvider
	providers := map[string]providers.Factory{
		"null": providersFactory(null.Provider()),
	}
	// default provisioners, if more are needed adds them with AddProvisioner
	provisioners := map[string]provisioners.Factory{}

	// initiate a Hook to send all logs to StdOut, this can and should be changed with Logger
	if logger == nil {
		logger = NewLogger(os.Stdout, "TERRAFORMER", DefLogLevel)
	}
	if len(hooks) == 0 {
		// hooks = []terraform.Hook{NewLogHook(logger)}
	}
	lw := NewLogWriter(logger)

	// the code and variables are inputs provided by the user, those cannot be initialized.
	tfr := Terraformer{
		Hooks:        hooks,
		lw:           lw,
		providers:    providers,
		provisioners: provisioners,
	}
	return &tfr, nil
}

// Apply do apply the changes and transform the infrastructure to a new state.
// If destroy is 'true' will destroy the infrastructure.
func (t *Terraformer) Apply(destroy bool) (err error) {
	// Do not set the log out before planning bc Plan also set it out and then
	// restore it. So, this will cause the following lines to log as TF does.
	t.lw.SetLogOut()
	defer t.lw.RestoreLogOut()

	action := "Apply"
	if destroy {
		action = "Destroy"
	}

	countHook := new(local.CountHook)
	stateHook := new(local.StateHook)
	ctx, err := t.NewContext(destroy, countHook, stateHook)
	if err != nil {
		return err
	}
	t.lw.Logger.Debugf("new context created and assigned")
	t.context = ctx

	// save the state before any action
	state := ctx.State()

	t.lw.Logger.Debugf("refreshing before applying")
	if err = t.refresh(); err != nil {
		return fmt.Errorf("error refreshing state before apply. %s", err)
	}

	plan, diag := ctx.Plan()
	if diag.HasErrors() {
		return fmt.Errorf("error getting plan before apply. %s", diag.Err())
	}
	t.lw.Logger.Debugf("plan to apply: %s", plan)
	t.Stats = NewStats(plan)
	t.lw.Logger.Infof("actions: %d to add, %d to change, %d to destroy", t.Stats.Add, t.Stats.Change, t.Stats.Destroy)

	// TODO: stateHook.State is not a terraform.State it is a state.State (terraform/state/state.go). Find how to get one of those, check backend/local/backend_apply.go L#56
	// stateHook.State =

	// Apply the changes and get the final state
	errorCh := make(chan error)

	go func() {
		t.lw.Logger.Debugf("%sing plan", action)
		_, diag = ctx.Apply()
		t.lw.Logger.Debugf("getting the state after %sing", strings.ToLower(action))
		state = ctx.State()
		errorCh <- diag.Err()
	}()

	// Wait until apply is completed
	err = <-errorCh
	close(errorCh)
	t.State = state

	// TODO: Save state to the file

	if err != nil {
		return fmt.Errorf("error applying changes. The state has been partially updated with successfully completed resources. %s", err)
	}

	t.lw.Logger.Infof("%s complete, resources: %d added, %d changed, %d destroyed", action, countHook.Added, countHook.Changed, countHook.Removed)
	if t.Stats.Add != countHook.Added || t.Stats.Change != countHook.Changed || t.Stats.Destroy != countHook.Removed {
		t.lw.Logger.Warnf("the planned changes are different to the applied: (%d,%d) added, (%d,%d) changed, (%d,%d) destroyed", t.Stats.Add, countHook.Added, t.Stats.Change, countHook.Changed, t.Stats.Destroy, countHook.Removed)
	} else {
		t.lw.Logger.Debugf("the planned changes and applied changes are the same: (%d==%d) added, (%d==%d) changed, (%d==%d) destroyed", t.Stats.Add, countHook.Added, t.Stats.Change, countHook.Changed, t.Stats.Destroy, countHook.Removed)
	}

	//  This is no needed:
	// var afterApplyPlan *terraform.Plan

	// afterApplyPlan, err = ctx.Plan()
	// if err != nil {
	// 	return err
	// }
	// if afterApplyPlan.Diff != nil && !afterApplyPlan.Diff.Empty() {
	// 	return fmt.Errorf("there are some actions that were not applied: %s", afterApplyPlan.Diff)
	// }

	// if !destroy {
	// 	if _, err = ctx.Refresh(); err != nil {
	// 		return err
	// 	}
	// 	// t.State = state

	// 	afterApplyPlan, err = ctx.Plan()
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if afterApplyPlan.Diff != nil && !afterApplyPlan.Diff.Empty() {
	// 		return fmt.Errorf("after refreshing, were found some actions that were not applied: %s", afterApplyPlan.Diff)
	// 	}
	// }

	return nil
}

// Refresh refreshes state of the existig (or not) infrastructure.
func (t *Terraformer) Refresh(destroy bool) (err error) {
	// Do not set the log out before planning bc Plan also set it out and then
	// restore it. So, this will cause the following lines to log as TF does.
	t.lw.SetLogOut()
	defer t.lw.RestoreLogOut()

	ctx := t.context
	if ctx != nil {
		ctx, err = t.NewContext(destroy)
		if err != nil {
			return err
		}
		t.lw.Logger.Debugf("new context created and assigned")
		t.context = ctx
	}

	return t.refresh()
}

func (t *Terraformer) refresh() error {
	_, diag := t.context.Refresh()
	return diag.Err()
}

// Plan do the planning of the changes and refresh/get the infrastructure current
// state. If destroy is 'true' will plan the destruction of the infrastructure.
// The result of Plan() is the current state of the infrastructure.
func (t *Terraformer) Plan(destroy bool) (plan *Plan, err error) {
	t.lw.SetLogOut()
	defer t.lw.RestoreLogOut()

	ctx, err := t.NewContext(destroy)
	if err != nil {
		return nil, err
	}
	t.lw.Logger.Debugf("new context created and assigned")
	t.context = ctx

	// Refreshing before get the plan
	if err := t.refresh(); err != nil {
		return nil, fmt.Errorf("error refreshing state before planning. %s", err)
	}

	var diag tfdiags.Diagnostics
	if plan, diag = ctx.Plan(); diag.HasErrors() {
		return nil, fmt.Errorf("error performing planing. %s", diag.Err())
	}
	t.plan = plan

	return plan, nil
}

// Stats contain the stats from a Plan
type Stats struct {
	Add, Change, Destroy int
}

// NewStats creates the stats from a Plan
func NewStats(plan *Plan) *Stats {
	stats := new(Stats)
	stats.Update(plan)
	return stats
}

// Update updates the stats
func (s *Stats) Update(plan *Plan) {
	if plan == nil || plan.Changes == nil || plan.Changes.Empty() {
		return
	}

	s.Add, s.Change, s.Destroy = 0, 0, 0

	for _, r := range plan.Changes.Resources {
		switch r.Action {
		case plans.Create:
			s.Add++
		case plans.Update:
			s.Change++
		case plans.DeleteThenCreate, plans.CreateThenDelete:
			s.Add++
			s.Destroy++
		case plans.Delete:
			if r.Addr.Resource.Resource.Mode != addrs.DataResourceMode {
				s.Destroy++
			}
		}
	}

	// if plan == nil || plan.Diff == nil || plan.Diff.Empty() {
	// 	return
	// }
	// s.Add, s.Change, s.Destroy = 0, 0, 0
	// for _, m := range plan.Diff.Modules {
	// 	for _, r := range m.Resources {
	// 		if r.Empty() {
	// 			continue
	// 		}
	// 		switch r.ChangeType() {
	// 		case terraform.DiffCreate:
	// 			s.Add++
	// 		case terraform.DiffUpdate:
	// 			s.Change++
	// 		case terraform.DiffDestroyCreate:
	// 			s.Add++
	// 			s.Destroy++
	// 		case terraform.DiffDestroy:
	// 			s.Destroy++
	// 		}
	// 	}
	// }
}

// Logger receives a Writer interface to send all the Terraform log or output
func (t *Terraformer) Logger(l Logger) {
	t.lw = NewLogWriter(l)
}

// Var adds a variable that is used in the code or Terraform template
func (t *Terraformer) Var(name string, value interface{}) {
	if len(t.Vars) == 0 {
		t.Vars = make(Variables)
	}
	t.Vars[name] = value
}

// AddVars adds all the given variable from a map to the variables to use in the TF code
func (t *Terraformer) AddVars(variables Variables) {
	for k, v := range variables {
		t.Var(k, v)
	}
}

// LoadState sets the current state of the infrastructure from a Reader such as
// a io.File, bytes.Buffer or something else that implements Read()
// The user still can get t.State to process it at will
func (t *Terraformer) LoadState(r io.Reader) error {
	state, err := LoadState(r)
	if err != nil {
		return err
	}
	t.State = state
	return nil
}

// LoadState sets the current state of the infrastructure from a Reader into the
// given reader
func LoadState(r io.Reader) (*State, error) {
	logW := log.Writer()
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(logW)

	sf, err := statefile.Read(r)
	if err != nil {
		return nil, err
	}
	return sf.State, nil
}

// SaveState writes the current state of the infrastructure into a Writer such as
// a io.File, bytes.Buffer or something else that implements Write()
// The user still can get t.State to process it at will
func (t *Terraformer) SaveState(w io.Writer) error {
	return SaveState(w, t.State)
}

// SaveState writes the given state of the infrastructure into the given writer
func SaveState(w io.Writer, state *State) error {
	logW := log.Writer()
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(logW)

	sf := statefile.New(state, "", 0)
	return statefile.Write(sf, w)
}

// AddProvider append a new provider to the list of providers to support in the
// Terraform templates
func (t *Terraformer) AddProvider(name string, provider terraform.ResourceProvider) {
	t.providers[name] = providersFactory(provider)
}

// AddProvisioner append a new provisioner to the list of provisioners to support
// in the Terraform templates
func (t *Terraformer) AddProvisioner(name string, provisioner terraform.ResourceProvisioner) {
	t.provisioners[name] = provisionersFactory(provisioner)
}

// NewContext return a Terraform context with the Terraform templates (Terraformer
// code), provisioners, providers, variables and the current infrastructure state
func (t *Terraformer) NewContext(destroy bool, hooks ...terraform.Hook) (*terraform.Context, error) {
	cfg, err := t.config()
	if err != nil {
		return nil, err
	}

	vars, err := t.variables(cfg.Module.Variables)
	if err != nil {
		return nil, err
	}

	if len(hooks) == 0 {
		hooks = []terraform.Hook{}
	}
	if len(t.Hooks) != 0 {
		hooks = append(hooks, t.Hooks...)
	}

	ctxOpts := terraform.ContextOpts{
		Config:           cfg,
		Destroy:          destroy,
		State:            t.State,
		Variables:        vars,
		Hooks:            hooks,
		ProviderResolver: providers.ResolverFixed(t.providers),
		Provisioners:     t.provisioners,
	}

	ctx, diags := terraform.NewContext(&ctxOpts)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error creating context. %s", diags.Err())
	}

	diags = ctx.Validate()
	var errCount int
	for _, d := range diags {
		if d.Severity() != tfdiags.Warning {
			errCount++
			continue
		}
		t.lw.Logger.Warnf("context validation: %s", d.Description().Summary)
	}
	if errCount > 0 {
		// if diags.HasErrors() {
		return nil, fmt.Errorf("found %d error(s) validating the context. %s", errCount, diags.Err())
	}

	return ctx, nil
}

// saveCode create a file in the given directory with the Terraform code in in
func (t *Terraformer) saveCode(dir string) error {
	// Create a file in the temporal directory to store the Terraform template
	filename := filepath.Join(dir, "main.tf")
	templatefile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer templatefile.Close()

	// Copy the Terraform template in t.Code into the new file
	if _, err = io.Copy(templatefile, bytes.NewReader(t.Code)); err != nil {
		return err
	}

	return nil
}

// config returns a Terraform config where the Terraform template or
// Terraformer Code will be loaded
func (t *Terraformer) config() (*configs.Config, error) {
	if len(t.Code) == 0 {
		return nil, fmt.Errorf("Code was not found")
	}

	tfrDir, err := ioutil.TempDir("", ".terraformer")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tfrDir)

	t.saveCode(tfrDir)

	cnfLoader, err := configload.NewLoader(&configload.Config{
		ModulesDir: filepath.Join(tfrDir, "modules"),
	})
	if err != nil {
		return nil, err
	}

	cnf, diags := cnfLoader.LoadConfig(tfrDir)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error loading the configuration. %s", diags.Error())
	}

	return cnf, nil
}

func (t *Terraformer) variables(v map[string]*configs.Variable) (terraform.InputValues, error) {
	iv := make(terraform.InputValues)
	for name, value := range t.Vars {
		if _, declared := v[name]; !declared {
			return iv, fmt.Errorf("variable %q is not declared in the code", name)
		}

		val := &terraform.InputValue{
			Value:      cty.StringVal(fmt.Sprintf("%v", value)),
			SourceType: terraform.ValueFromCaller,
		}

		iv[name] = val
	}

	return iv, nil
}
