package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type pairedSample struct {
	TaskID string    `json:"task_id"`
	Seed   int       `json:"seed"`
	Family string    `json:"family"`
	Direct pairedArm `json:"direct"`
	AMR    pairedArm `json:"amr"`
}

type pairedArm struct {
	Correct           bool  `json:"correct"`
	AIUNano           int64 `json:"aiu_nano"`
	LatencyMS         int64 `json:"latency_ms"`
	Interventions     int64 `json:"interventions"`
	FallbackUsed      bool  `json:"fallback_used"`
	RepairTailAIUNano int64 `json:"repair_tail_aiu_nano"`
	TelemetryComplete bool  `json:"telemetry_complete"`
	RouteMatch        bool  `json:"route_match"`
	OracleIntact      bool  `json:"oracle_intact"`
	IsolationReady    bool  `json:"isolation_ready"`
	ProofPresent      bool  `json:"proof_present"`
	Calls             int64 `json:"calls"`
}

type pairedGradeReport struct {
	SchemaVersion   int                          `json:"schema_version"`
	ReleaseDecision string                       `json:"release_decision"`
	PaidCalls       int64                        `json:"paid_calls"`
	Families        map[string]pairedFamilyGrade `json:"families"`
}

type pairedRunRef struct {
	SchemaVersion int           `json:"schema_version"`
	TaskID        string        `json:"task_id"`
	Seed          int           `json:"seed"`
	Family        string        `json:"family"`
	Direct        pairedFileRef `json:"direct"`
	AMR           pairedFileRef `json:"amr"`
}

type pairedFileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type pairedFamilyGrade struct {
	Family                string  `json:"family"`
	Samples               int     `json:"samples"`
	DirectCorrect         int     `json:"direct_correct"`
	AMRCorrect            int     `json:"amr_correct"`
	DirectAIUNano         int64   `json:"direct_aiu_nano"`
	AMRAIUNano            int64   `json:"amr_aiu_nano"`
	PairedMedianReduction float64 `json:"paired_median_reduction"`
	ConfidenceLower       float64 `json:"confidence_lower"`
	ConfidenceUpper       float64 `json:"confidence_upper"`
	DirectLatencyP90MS    int64   `json:"direct_latency_p90_ms"`
	AMRLatencyP90MS       int64   `json:"amr_latency_p90_ms"`
	FallbackCount         int     `json:"fallback_count"`
	RepairTailP90Nano     int64   `json:"repair_tail_p90_aiu_nano"`
	DirectInterventions   int64   `json:"direct_interventions"`
	AMRInterventions      int64   `json:"amr_interventions"`
	ZeroRightToWrong      bool    `json:"zero_right_to_wrong"`
	PromotionEligible     bool    `json:"promotion_eligible"`
}

func runPairedGrade(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("grade-paired", flag.ContinueOnError)
	resultsPath := fs.String("results", "", "paired run references JSONL")
	configPath := fs.String("config", defaultConfig, "benchmark config")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || *resultsPath == "" {
		return fmt.Errorf("grade-paired requires --results <jsonl>")
	}
	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	file, err := os.Open(*resultsPath)
	if err != nil {
		return err
	}
	defer file.Close()
	var samples []pairedSample
	base := filepath.Dir(*resultsPath)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		var reference pairedRunRef
		if err := json.Unmarshal(scanner.Bytes(), &reference); err != nil {
			return err
		}
		sample, err := derivePairedSample(base, cfg, reference)
		if err != nil {
			return err
		}
		samples = append(samples, sample)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	report, err := gradePairedSamples(samples)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func derivePairedSample(base string, cfg config, reference pairedRunRef) (pairedSample, error) {
	if reference.SchemaVersion != 1 || reference.TaskID == "" || reference.Family == "" {
		return pairedSample{}, fmt.Errorf("invalid paired run reference")
	}
	task, ok := findTask(cfg, reference.TaskID)
	if !ok || task.Family != reference.Family {
		return pairedSample{}, fmt.Errorf("paired run reference is outside configured task family")
	}
	direct, err := loadBoundRunResult(base, reference.Direct)
	if err != nil {
		return pairedSample{}, err
	}
	amr, err := loadBoundRunResult(base, reference.AMR)
	if err != nil {
		return pairedSample{}, err
	}
	if direct.Mode != "direct" || amr.Mode != "amr" ||
		direct.TaskID != reference.TaskID || amr.TaskID != reference.TaskID ||
		direct.TaskFamily != reference.Family || amr.TaskFamily != reference.Family ||
		direct.Seed != reference.Seed || amr.Seed != reference.Seed ||
		direct.ContextContractHash == "" || direct.ContextContractHash != amr.ContextContractHash ||
		direct.ProofClosureHash == "" || direct.ProofClosureHash != amr.ProofClosureHash ||
		direct.ExperimentID == "" || direct.ExperimentID != amr.ExperimentID ||
		direct.RouteCatalogHash == "" || direct.RouteCatalogHash != amr.RouteCatalogHash ||
		direct.DriverModel == "" || direct.DriverModel != amr.DriverModel ||
		direct.PlannedTier != task.PlannedTier || amr.PlannedTier != task.PlannedTier ||
		amr.AttemptTier != task.AttemptTier {
		return pairedSample{}, fmt.Errorf("paired run artifacts do not share frozen task, seed, route-tier, context, and proof identities")
	}
	return pairedSample{
		TaskID: reference.TaskID, Seed: reference.Seed, Family: reference.Family,
		Direct: armFromRunResult(direct), AMR: armFromRunResult(amr),
	}, nil
}

func loadBoundRunResult(base string, reference pairedFileRef) (runResult, error) {
	var result runResult
	clean := filepath.Clean(filepath.FromSlash(reference.Path))
	if clean == "." || clean == ".." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return result, fmt.Errorf("paired result path escapes reference root")
	}
	path := filepath.Join(base, clean)
	info, err := os.Lstat(path)
	if err != nil {
		return result, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return result, fmt.Errorf("paired result must be a regular non-symlink file")
	}
	hash, err := fileSHA256(path)
	if err != nil {
		return result, err
	}
	if hash != reference.SHA256 {
		return result, fmt.Errorf("paired result hash mismatch")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

func armFromRunResult(result runResult) pairedArm {
	arm := pairedArm{
		Correct: result.FinalProof.Passed, LatencyMS: result.DurationMS,
		OracleIntact: result.ProofClosureHash != "", ProofPresent: result.FinalProof.ExitCode == 0,
		IsolationReady: result.FinalProof.SandboxImage != "",
	}
	arm.TelemetryComplete = len(result.Phases) > 0
	arm.RouteMatch = len(result.Phases) > 0
	for index, phase := range result.Phases {
		if !phase.Valid || !phase.AICAvailable || phase.Calls < 1 {
			arm.TelemetryComplete = false
		}
		if !phase.ModelMatch {
			arm.RouteMatch = false
		}
		next, err := addContextCost(arm.AIUNano, phase.AIUNano)
		if err != nil {
			arm.TelemetryComplete = false
		} else {
			arm.AIUNano = next
		}
		arm.Calls += int64(phase.Calls)
		if index > 0 {
			arm.FallbackUsed = true
			arm.RepairTailAIUNano += phase.AIUNano
		}
	}
	return arm
}

func gradePairedSamples(samples []pairedSample) (pairedGradeReport, error) {
	if len(samples) == 0 {
		return pairedGradeReport{}, fmt.Errorf("paired samples are required")
	}
	grouped := map[string][]pairedSample{}
	seen := map[string]bool{}
	for _, sample := range samples {
		key := fmt.Sprintf("%s/%d", sample.TaskID, sample.Seed)
		if sample.TaskID == "" || sample.Family == "" || seen[key] {
			return pairedGradeReport{}, fmt.Errorf("duplicate or incomplete paired sample %q", key)
		}
		seen[key] = true
		if err := validatePairedArm("direct", sample.Direct); err != nil {
			return pairedGradeReport{}, fmt.Errorf("%s: %w", key, err)
		}
		if err := validatePairedArm("amr", sample.AMR); err != nil {
			return pairedGradeReport{}, fmt.Errorf("%s: %w", key, err)
		}
		grouped[sample.Family] = append(grouped[sample.Family], sample)
	}
	report := pairedGradeReport{SchemaVersion: 1, ReleaseDecision: "not-promoted", Families: map[string]pairedFamilyGrade{}}
	for family, familySamples := range grouped {
		grade := pairedFamilyGrade{Family: family, Samples: len(familySamples), ZeroRightToWrong: true}
		reductions := make([]float64, 0, len(familySamples))
		directLatency := make([]int64, 0, len(familySamples))
		amrLatency := make([]int64, 0, len(familySamples))
		repairTails := make([]int64, 0, len(familySamples))
		for _, sample := range familySamples {
			if sample.Direct.Correct {
				grade.DirectCorrect++
			}
			if sample.AMR.Correct {
				grade.AMRCorrect++
			}
			if sample.Direct.Correct && !sample.AMR.Correct {
				grade.ZeroRightToWrong = false
			}
			var err error
			grade.DirectAIUNano, err = addContextCost(grade.DirectAIUNano, sample.Direct.AIUNano)
			if err != nil {
				return pairedGradeReport{}, err
			}
			grade.AMRAIUNano, err = addContextCost(grade.AMRAIUNano, sample.AMR.AIUNano)
			if err != nil {
				return pairedGradeReport{}, err
			}
			reductions = append(reductions, float64(sample.Direct.AIUNano-sample.AMR.AIUNano)/float64(sample.Direct.AIUNano))
			directLatency = append(directLatency, sample.Direct.LatencyMS)
			amrLatency = append(amrLatency, sample.AMR.LatencyMS)
			repairTails = append(repairTails, sample.AMR.RepairTailAIUNano)
			grade.DirectInterventions, err = addContextCost(grade.DirectInterventions, sample.Direct.Interventions)
			if err != nil {
				return pairedGradeReport{}, err
			}
			grade.AMRInterventions, err = addContextCost(grade.AMRInterventions, sample.AMR.Interventions)
			if err != nil {
				return pairedGradeReport{}, err
			}
			if sample.AMR.FallbackUsed {
				grade.FallbackCount++
			}
			report.PaidCalls, err = addContextCost(report.PaidCalls, sample.Direct.Calls+sample.AMR.Calls)
			if err != nil {
				return pairedGradeReport{}, err
			}
		}
		grade.PairedMedianReduction = contextMedian(reductions)
		grade.ConfidenceLower, grade.ConfidenceUpper = bootstrapMedianInterval(reductions, 0.95)
		grade.DirectLatencyP90MS = percentileInt64(directLatency, 0.90)
		grade.AMRLatencyP90MS = percentileInt64(amrLatency, 0.90)
		grade.RepairTailP90Nano = percentileInt64(repairTails, 0.90)
		grade.PromotionEligible = grade.Samples >= 20 &&
			grade.ZeroRightToWrong &&
			grade.AMRCorrect >= grade.DirectCorrect &&
			grade.AMRAIUNano < grade.DirectAIUNano &&
			grade.PairedMedianReduction >= 0.20 &&
			grade.ConfidenceLower > 0 &&
			grade.AMRInterventions <= grade.DirectInterventions
		report.Families[family] = grade
	}
	return report, nil
}

func validatePairedArm(name string, arm pairedArm) error {
	if arm.AIUNano <= 0 || arm.LatencyMS < 0 || arm.Interventions < 0 || arm.RepairTailAIUNano < 0 {
		return fmt.Errorf("%s arm has invalid numeric evidence", name)
	}
	if !arm.TelemetryComplete || !arm.RouteMatch || !arm.OracleIntact || !arm.IsolationReady || !arm.ProofPresent {
		return fmt.Errorf("%s arm is invalid: telemetry, route, oracle, isolation, and proof must all be complete", name)
	}
	return nil
}

func percentileInt64(values []int64, percentile float64) int64 {
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	index := int(float64(len(sorted)-1) * percentile)
	return sorted[index]
}
