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
	runRoot, runID, tier, attemptTier, executionOwner, ownerReason, tierReason string
	taskFamily, risk, override, alias, prefer                                  string
	tools                                                                      repeatFlag
	contextSize                                                                int
	sensitive                                                                  bool
}

type selectOutput struct {
	Status         modelrouting.SelectionStatus `json:"status"`
	PlannedTier    modelrouting.Tier            `json:"planned_tier"`
	AttemptTier    modelrouting.Tier            `json:"attempt_tier"`
	ExecutionOwner modelrouting.ExecutionOwner  `json:"execution_owner"`
	OwnerReason    string                       `json:"owner_reason"`
	TierReason     string                       `json:"tier_reason"`
	Preference     modelrouting.RoutePreference `json:"preference,omitempty"`
	Alias          string                       `json:"alias,omitempty"`
	Aliases        []string                     `json:"aliases,omitempty"`
	CurrentModel   string                       `json:"current_model,omitempty"`
	Fallback       string                       `json:"fallback,omitempty"`
	ErrorClass     string                       `json:"error_class,omitempty"`
}

func runModelsSelect(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("models select", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	opts := selectOptions{}
	opts.commonOptions.bind(fs)
	fs.StringVar(&opts.runRoot, "run-root", "", "marked KB run root")
	fs.StringVar(&opts.runID, "run-id", "", "KB run id")
	fs.StringVar(&opts.tier, "tier", "", "small, medium, or large")
	fs.StringVar(&opts.attemptTier, "attempt-tier", "", "experimental AMR compatibility only; omitted by normal DDR")
	fs.StringVar(&opts.executionOwner, "execution-owner", "", "orchestrator decision: current or delegated")
	fs.StringVar(&opts.ownerReason, "owner-reason", "", "delegated rationale, or current: reasoning-required|context-required|tool-required|authority-required|trust-required|user-required|no-qualified-route (optional ': explanation')")
	fs.StringVar(&opts.tierReason, "tier-reason", "", "why this capability tier is required")
	fs.StringVar(&opts.taskFamily, "task-family", "", "task family")
	fs.Var(&opts.tools, "tool", "required tool; repeatable")
	fs.IntVar(&opts.contextSize, "context-size", 0, "required context size")
	fs.StringVar(&opts.risk, "risk", "", "normal or broad")
	fs.BoolVar(&opts.sensitive, "sensitive-data", false, "work contains sensitive data")
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
	if opts.runRoot == "" || opts.runID == "" || opts.tier == "" || opts.executionOwner == "" ||
		strings.TrimSpace(opts.ownerReason) == "" || strings.TrimSpace(opts.tierReason) == "" ||
		opts.taskFamily == "" || len(opts.tools) == 0 || opts.contextSize <= 0 || opts.risk == "" {
		fmt.Fprintln(stderr, "select requires complete run, ownership, and work-request bindings")
		return 2
	}
	owner := modelrouting.ExecutionOwner(strings.TrimSpace(opts.executionOwner))
	if owner != modelrouting.ExecutionOwnerCurrent && owner != modelrouting.ExecutionOwnerDelegated {
		fmt.Fprintln(stderr, "execution owner must be current or delegated")
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
	request := modelrouting.WorkRequest{
		PlannedTier: modelrouting.Tier(opts.tier), AttemptTier: modelrouting.Tier(opts.attemptTier),
		ExecutionOwner: owner, OwnerReason: strings.TrimSpace(opts.ownerReason), TierReason: strings.TrimSpace(opts.tierReason),
		TaskFamily: opts.taskFamily, Tools: []string(opts.tools), ContextSize: opts.contextSize,
		Risk: modelrouting.RiskLevel(opts.risk), SensitiveData: opts.sensitive, ProjectID: policy.Project.ProjectID,
	}
	decision, selectErr := modelrouting.SelectRoute(validated, request, policy, modelrouting.RunOverride{Mode: mode, Alias: opts.alias, Prefer: preference}, modelrouting.AttemptLedger{}, time.Now())
	out := selectOutput{
		Status: decision.Status, PlannedTier: decision.PlannedTier, AttemptTier: decision.AttemptTier,
		ExecutionOwner: decision.ExecutionOwner, OwnerReason: decision.OwnerReason, TierReason: decision.TierReason,
		Preference: decision.Preference, CurrentModel: decision.Current.ModelID,
	}
	for _, route := range decision.Routes {
		out.Aliases = append(out.Aliases, route.Alias)
	}
	if len(out.Aliases) == 1 {
		out.Alias = out.Aliases[0]
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
		fmt.Fprintf(stdout, "selection: %s owner=%s owner-reason=%q planned-tier=%s tier-reason=%q attempt-tier=%s preference=%s alias=%s current=%s fallback=%s error=%s\n",
			out.Status, out.ExecutionOwner, out.OwnerReason, out.PlannedTier, out.TierReason, out.AttemptTier,
			out.Preference, out.Alias, out.CurrentModel, out.Fallback, out.ErrorClass)
	}
	if selectErr != nil {
		return 1
	}
	return 0
}
