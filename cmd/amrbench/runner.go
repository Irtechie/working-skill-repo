package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var errPaidRunnerDisabled = errors.New("paid runner is disabled")

type runnerRequest struct{}
type runnerResponse struct{}

type modelRunner interface {
	Run(context.Context, runnerRequest) (runnerResponse, error)
	Name() string
}

type disabledRunner struct {
	mu      sync.Mutex
	started int
}

func newDisabledRunner(_ func(), _ func(string) string) *disabledRunner {
	return &disabledRunner{}
}

func (runner *disabledRunner) Run(context.Context, runnerRequest) (runnerResponse, error) {
	return runnerResponse{}, errPaidRunnerDisabled
}

func (runner *disabledRunner) Name() string { return "disabled" }

func (runner *disabledRunner) Started() int {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	return runner.started
}

var constructLiveRunner = func() modelRunner { return nil }

type creditBudget struct {
	perCall    int64
	perArm     int64
	experiment int64
	armTotal   int64
	total      int64
	arms       map[string]int64
}

func newCreditBudget(perCall, perArm, experiment int) *creditBudget {
	return &creditBudget{perCall: int64(perCall), perArm: int64(perArm), experiment: int64(experiment), arms: map[string]int64{}}
}

func (budget *creditBudget) Reserve(arm string, credits int) error {
	value := int64(credits)
	if value <= 0 || value > budget.perCall {
		return fmt.Errorf("credit request %d exceeds per-call ceiling %d", credits, budget.perCall)
	}
	if budget.armTotal > budget.perArm-value {
		return fmt.Errorf("arm would exceed cumulative credit ceiling %d before %q", budget.perArm, arm)
	}
	if budget.total > budget.experiment-value {
		return fmt.Errorf("experiment would exceed credit ceiling %d", budget.experiment)
	}
	budget.arms[arm] += value
	budget.armTotal += value
	budget.total += value
	return nil
}

func (budget *creditBudget) Reserved() int64 { return budget.total }

func (budget *creditBudget) BeginArm() {
	budget.armTotal = 0
}
