package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

type selectOptions struct {
	commonOptions
	runRoot, runID, tier, attemptTier, taskFamily, risk, override, alias, prefer, currentReason string
	tools                                                                                       repeatFlag
	contextSize                                                                                 int
	sensitive                                                                                   bool
	allowCurrent                                                                                bool
}

type selectOutput struct {
	Status         modelrouting.SelectionStatus        `json:"status"`
	PlannedTier    modelrouting.Tier                   `json:"planned_tier"`
	AttemptTier    modelrouting.Tier                   `json:"attempt_tier"`
	Preference     modelrouting.RoutePreference        `json:"preference,omitempty"`
	Aliases        []string                            `json:"aliases,omitempty"`
	CurrentModel   string                              `json:"current_model,omitempty"`
	ExecutionOwner string                              `json:"execution_owner,omitempty"`
	CurrentReason  modelrouting.CurrentExecutionReason `json:"current_reason,omitempty"`
	Fallback       string                              `json:"fallback,omitempty"`
	ErrorClass     string                              `json:"error_class,omitempty"`
}

func runModelsSelect(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("models select", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	opts := selectOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.runRoot, "run-root", "", "marked KB run root")
	fs.StringVar(&opts.runID, "run-id", "", "KB run id")
	fs.StringVar(&opts.tier, "tier", "", "small, medium, or large")
	fs.StringVar(&opts.attemptTier, "attempt-tier", "", "optional caller-vetted lower initial tier")
	fs.StringVar(&opts.taskFamily, "task-family", "", "task family")
	fs.Var(&opts.tools, "tool", "required tool; repeatable")
	fs.IntVar(&opts.contextSize, "context-size", 0, "required context size")
	fs.StringVar(&opts.risk, "risk", "", "normal or broad")
	fs.BoolVar(&opts.sensitive, "sensitive-data", false, "work contains sensitive data")
	fs.BoolVar(&opts.allowCurrent, "allow-current", false, "allow current reasoner execution when no worker qualifies")
	fs.StringVar(&opts.currentReason, "current-reason", "", "reasoner-required or no-qualified-route")
	fs.StringVar(&opts.override, "override", "", "run-only use, require, or ignore")
	fs.StringVar(&opts.alias, "alias", "", "run-only override alias")
	fs.StringVar(&opts.prefer, "prefer", "", "run-only self-hosted or native preference")
	if err := fs.Parse(args); err != nil {
		return flagError(stderr, err)
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(stderr, "unexpected argument %q\n", fs.Arg(0))
		return 2
	}
	if customUserRootRejected(fs) {
		fmt.Fprintln(stderr, "selection uses the fixed user-local trust root; custom --user-root is test-only")
		return 2
	}
	if opts.runRoot == "" || opts.runID == "" || opts.tier == "" || opts.taskFamily == "" || len(opts.tools) == 0 || opts.contextSize <= 0 || opts.risk == "" {
		fmt.Fprintln(stderr, "select requires complete run and work-request bindings")
		return 2
	}
	prepared, err := prepareRunRoot(opts.projectRoot, opts.runRoot)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if prepared.marker.RunID != opts.runID {
		fmt.Fprintln(stderr, "run id does not match prepared marker")
		return 1
	}
	state, err := dispatchTrustedStateProvider(opts.userRoot, prepared.runPath)
	if err != nil {
		fmt.Fprintln(stderr, "router-unavailable: "+err.Error())
		return 1
	}
	validated, policy, err := loadDispatchCatalog(prepared, dispatchOptions{commonOptions: opts.commonOptions}, state)
	if err != nil {
		fmt.Fprintln(stderr, "router-unavailable: "+err.Error())
		return 1
	}
	mode := modelrouting.OverrideMode(strings.TrimSpace(opts.override))
	if mode != "" && mode != modelrouting.OverrideUse && mode != modelrouting.OverrideRequire && mode != modelrouting.OverrideIgnore {
		fmt.Fprintln(stderr, "unsupported run override")
		return 2
	}
	if (mode == modelrouting.OverrideUse || mode == modelrouting.OverrideRequire) && opts.alias == "" {
		fmt.Fprintln(stderr, "use/require override needs --alias")
		return 2
	}
	preference := modelrouting.RoutePreference(strings.TrimSpace(opts.prefer))
	switch preference {
	case "self-hosted":
		preference = modelrouting.PreferenceSelfHostedFirst
	case "native":
		preference = modelrouting.PreferenceNativeFirst
	case "":
		if mode == "" {
			priorities, loadErr := loadProjectPriorities(opts.userRoot)
			if loadErr != nil {
				fmt.Fprintln(stderr, "router-unavailable: load project priority: "+loadErr.Error())
				return 1
			}
			preference = priorities.priorityFor(policy.Project.ProjectID)
		} else {
			preference = modelrouting.PreferenceAutomatic
		}
	}
	if !validStoredPriority(preference) {
		fmt.Fprintln(stderr, "unsupported route preference")
		return 2
	}
	request := modelrouting.WorkRequest{PlannedTier: modelrouting.Tier(opts.tier), AttemptTier: modelrouting.Tier(opts.attemptTier), AllowCurrent: opts.allowCurrent, CurrentReason: modelrouting.CurrentExecutionReason(opts.currentReason), TaskFamily: opts.taskFamily, Tools: []string(opts.tools), ContextSize: opts.contextSize, Risk: modelrouting.RiskLevel(opts.risk), SensitiveData: opts.sensitive, ProjectID: policy.Project.ProjectID}
	decision, selectErr := modelrouting.SelectRoute(validated, request, policy, modelrouting.RunOverride{Mode: mode, Alias: opts.alias, Prefer: preference}, modelrouting.AttemptLedger{}, time.Now())
	out := selectOutput{Status: decision.Status, PlannedTier: decision.PlannedTier, AttemptTier: decision.AttemptTier, Preference: decision.Preference, CurrentModel: decision.Current.ModelID, ExecutionOwner: decision.ExecutionOwner, CurrentReason: decision.CurrentReason}
	for _, route := range decision.Routes {
		out.Aliases = append(out.Aliases, route.Alias)
	}
	if decision.Status == modelrouting.SelectionDegraded {
		out.Fallback = "current-model-degraded"
	}
	if selectErr != nil {
		switch {
		case errors.Is(selectErr, modelrouting.ErrRequiredRouteUnavailable):
			out.ErrorClass = "required-route-unavailable"
		case errors.Is(selectErr, modelrouting.ErrInvalidWorkRequest):
			out.ErrorClass = "invalid-work-request"
		default:
			out.ErrorClass = "selection-error"
		}
	}
	if opts.json {
		if printResult(stdout, stderr, out, true, nil) != 0 {
			return 1
		}
	} else {
		fmt.Fprintf(stdout, "selection: %s planned-tier=%s attempt-tier=%s preference=%s aliases=%s current=%s fallback=%s error=%s\n", out.Status, out.PlannedTier, out.AttemptTier, out.Preference, strings.Join(out.Aliases, ","), out.CurrentModel, out.Fallback, out.ErrorClass)
	}
	if selectErr != nil {
		return 1
	}
	return 0
}
