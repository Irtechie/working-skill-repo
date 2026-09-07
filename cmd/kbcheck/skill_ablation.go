package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

// These are trusted imports, not authenticated capture attestations. No command
// from an input is executed. Hashes establish byte integrity and identity only.
type ablationIdentity struct {
	RunID       string `json:"run_id"`
	Case        string `json:"case"`
	Repetition  int    `json:"repetition"`
	Host        string `json:"host"`
	HostVersion string `json:"host_version"`
	Model       string `json:"model"`
	Config      string `json:"config_sha256"`
	Project     string `json:"project_sha256"`
	Prompt      string `json:"prompt_sha256"`
	Condition   string `json:"condition"`
}
type ablationRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type ablationRecord struct {
	Schema     int              `json:"schema_version"`
	Kind       string           `json:"evidence_kind"`
	Identity   ablationIdentity `json:"identity"`
	Transcript ablationRef      `json:"transcript"`
	Inventory  ablationRef      `json:"inventory"`
	Proof      ablationRef      `json:"proof"`
}
type ablationTranscript struct {
	Schema   int              `json:"schema_version"`
	Identity ablationIdentity `json:"identity"`
	Source   string           `json:"source"`
	Profile  string           `json:"profile_sha256"`
	Content  string           `json:"content"`
}
type ablationInstruction struct {
	Name    string      `json:"name"`
	Scope   string      `json:"scope"`
	KB      *bool       `json:"kb"`
	Content ablationRef `json:"content"`
}
type ablationInventory struct {
	Schema       int                   `json:"schema_version"`
	Identity     ablationIdentity      `json:"identity"`
	Source       string                `json:"source"`
	Coverage     map[string]bool       `json:"discovery_coverage"`
	Skills       []ablationInstruction `json:"skills"`
	Instructions []ablationInstruction `json:"instructions"`
}
type ablationMetric struct {
	Value  *float64 `json:"value"`
	Source string   `json:"source"`
}
type ablationMetrics struct {
	Tokens        ablationMetric `json:"tokens"`
	Cost          ablationMetric `json:"cost_usd"`
	Interventions ablationMetric `json:"interventions"`
	Artifacts     ablationMetric `json:"artifact_count"`
	Elapsed       ablationMetric `json:"elapsed_ms"`
}
type ablationProof struct {
	Schema     int              `json:"schema_version"`
	Identity   ablationIdentity `json:"identity"`
	Source     string           `json:"source"`
	Profile    string           `json:"profile_sha256"`
	Transcript string           `json:"transcript_sha256"`
	Command    []string         `json:"command"`
	Exit       *int             `json:"exit_code"`
	Started    string           `json:"started_at"`
	Ended      string           `json:"ended_at"`
	Contract   ablationRef      `json:"check_contract"`
	Outputs    []ablationRef    `json:"outputs"`
	Metrics    ablationMetrics  `json:"metrics"`
}
type ablationContract struct {
	Schema  int         `json:"schema_version"`
	Case    string      `json:"case"`
	Command []string    `json:"command"`
	Oracle  ablationRef `json:"oracle"`
}
type ablationOutcome struct {
	RunID   string              `json:"run_id"`
	Success bool                `json:"success"`
	Exit    int                 `json:"exit_code"`
	Metrics map[string]*float64 `json:"metrics"`
}
type ablationGroup struct {
	Identity    ablationIdentity               `json:"identity"`
	Arms        map[string]ablationOutcome     `json:"arms"`
	Differences map[string]map[string]*float64 `json:"differences"`
	Regressions []string                       `json:"correctness_regressions"`
}
type ablationReport struct {
	Schema     int                 `json:"schema_version"`
	Evidence   string              `json:"evidence_class"`
	Limitation string              `json:"integrity_limit"`
	Eligible   int                 `json:"eligible"`
	Excluded   int                 `json:"excluded"`
	Conditions map[string]int      `json:"conditions"`
	Matched    []ablationGroup     `json:"matched_groups"`
	Incomplete []map[string]any    `json:"incomplete_groups"`
	Issues     []map[string]string `json:"issues"`
}
type ablationCandidate struct {
	row        ablationRecord
	file, key  string
	outcome    ablationOutcome
	comparison string
	err        error
}
type ablationReader struct {
	root string
	used map[string]bool
}

func (r *ablationReader) read(ref ablationRef, value any) error {
	if !validModelTierHash(ref.SHA256) {
		return fmt.Errorf("missing or invalid evidence SHA256")
	}
	path, err := resolveModelTierFile(r.root, ref.Path)
	if err != nil {
		return err
	}
	r.used[strings.ToLower(path)] = true
	raw, err := readSafeBoundedModelTierFile(path)
	if err != nil {
		return err
	}
	if hashString(string(raw)) != ref.SHA256 {
		return fmt.Errorf("evidence hash mismatch: %s", ref.Path)
	}
	if value != nil {
		return decodeAblationJSON(path, raw, value)
	}
	return nil
}
func decodeAblationJSON(path string, raw []byte, value any) error {
	// JSON duplicate keys otherwise permit a later success to hide an earlier failure.
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var walk func() error
	walk = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		if delimiter, ok := token.(json.Delim); ok {
			switch delimiter {
			case '{':
				seen := map[string]bool{}
				for decoder.More() {
					key, err := decoder.Token()
					if err != nil {
						return err
					}
					name := strings.ToLower(key.(string))
					if seen[name] {
						return fmt.Errorf("duplicate JSON key: %s", name)
					}
					seen[name] = true
					if err = walk(); err != nil {
						return err
					}
				}
			case '[':
				for decoder.More() {
					if err := walk(); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unexpected JSON delimiter")
			}
			_, err = decoder.Token()
			return err
		}
		return nil
	}
	if err := walk(); err != nil {
		return err
	}
	return decodeStrictModelTierJSON(path, raw, value)
}
func ablationIdentityValid(id ablationIdentity) bool {
	return id.RunID != "" && id.Case != "" && id.Repetition > 0 && id.Host != "" && id.HostVersion != "" && id.Model != "" && validModelTierHash(id.Config) && validModelTierHash(id.Project) && validModelTierHash(id.Prompt) && (id.Condition == "full" || id.Condition == "reduced" || id.Condition == "none")
}
func ablationKey(id ablationIdentity) string {
	id.RunID = ""
	id.Condition = ""
	b, _ := json.Marshal(id)
	return string(b)
}
func ablationMetricValue(m ablationMetric, integer bool) *float64 {
	if m.Value == nil || m.Source != "external-capture" || *m.Value < 0 {
		return nil
	}
	if integer && *m.Value != float64(int64(*m.Value)) {
		return nil
	}
	return m.Value
}
func validateAblation(r *ablationReader, row ablationRecord) (ablationOutcome, string, error) {
	outcome := ablationOutcome{RunID: row.Identity.RunID, Metrics: map[string]*float64{}}
	var inventory ablationInventory
	if err := r.read(row.Inventory, &inventory); err != nil {
		return outcome, "", err
	}
	if inventory.Schema != 1 || inventory.Identity != row.Identity || inventory.Source != "external-capture" {
		return outcome, "", fmt.Errorf("inventory capture identity mismatch")
	}
	for _, scope := range []string{"global", "project", "skills", "native"} {
		if !inventory.Coverage[scope] {
			return outcome, "", fmt.Errorf("effective instruction discovery incomplete: %s", scope)
		}
	}
	baseline := []string{}
	seen := map[string]bool{}
	kbCount := 0
	nativeCount := 0
	for _, entry := range append(append([]ablationInstruction{}, inventory.Skills...), inventory.Instructions...) {
		if entry.Name == "" || entry.KB == nil || entry.Scope == "" {
			return outcome, "", fmt.Errorf("instruction inventory entry incomplete")
		}
		key := entry.Scope + "|" + entry.Name
		if seen[key] {
			return outcome, "", fmt.Errorf("duplicate inventory entry")
		}
		seen[key] = true
		if err := r.read(entry.Content, nil); err != nil {
			return outcome, "", err
		}
		isKB := *entry.KB || strings.HasPrefix(strings.ToLower(entry.Name), "kb-") || strings.Contains(strings.ToLower(entry.Name), "/kb-")
		if isKB {
			kbCount++
		} else {
			baseline = append(baseline, key+"|"+entry.Content.SHA256)
			if entry.Scope == "native" {
				nativeCount++
			}
		}
	}
	if row.Identity.Condition == "none" && kbCount != 0 {
		return outcome, "", fmt.Errorf("none condition contaminated by KB instructions")
	}
	if row.Identity.Condition != "none" && kbCount == 0 {
		return outcome, "", fmt.Errorf("KB profile has no effective KB instructions")
	}
	if nativeCount == 0 {
		return outcome, "", fmt.Errorf("host-native instruction baseline missing")
	}
	sort.Strings(baseline)
	var transcript ablationTranscript
	if err := r.read(row.Transcript, &transcript); err != nil {
		return outcome, "", err
	}
	if transcript.Schema != 1 || transcript.Identity != row.Identity || transcript.Source != "external-capture" || transcript.Profile != row.Inventory.SHA256 || strings.TrimSpace(transcript.Content) == "" {
		return outcome, "", fmt.Errorf("transcript capture identity/profile mismatch")
	}
	var proof ablationProof
	if err := r.read(row.Proof, &proof); err != nil {
		return outcome, "", err
	}
	if proof.Schema != 1 || proof.Identity != row.Identity || proof.Source != "external-capture" || proof.Profile != row.Inventory.SHA256 || proof.Transcript != row.Transcript.SHA256 || proof.Exit == nil || len(proof.Command) == 0 || len(proof.Outputs) == 0 {
		return outcome, "", fmt.Errorf("command receipt capture identity/result missing or mismatched")
	}
	started, e1 := time.Parse(time.RFC3339Nano, proof.Started)
	ended, e2 := time.Parse(time.RFC3339Nano, proof.Ended)
	if e1 != nil || e2 != nil || ended.Before(started) {
		return outcome, "", fmt.Errorf("command receipt timestamps invalid")
	}
	var contract ablationContract
	if err := r.read(proof.Contract, &contract); err != nil {
		return outcome, "", err
	}
	if contract.Schema != 1 || contract.Case != row.Identity.Case || !reflect.DeepEqual(contract.Command, proof.Command) || strings.TrimSpace(proof.Command[0]) == "" {
		return outcome, "", fmt.Errorf("captured command differs from frozen check contract")
	}
	if err := r.read(contract.Oracle, nil); err != nil {
		return outcome, "", err
	}
	for _, artifact := range proof.Outputs {
		if err := r.read(artifact, nil); err != nil {
			return outcome, "", err
		}
	}
	outcome.Exit = *proof.Exit
	outcome.Success = *proof.Exit == 0
	outcome.Metrics = map[string]*float64{"tokens": ablationMetricValue(proof.Metrics.Tokens, true), "cost_usd": ablationMetricValue(proof.Metrics.Cost, false), "interventions": ablationMetricValue(proof.Metrics.Interventions, true), "artifact_count": ablationMetricValue(proof.Metrics.Artifacts, true), "elapsed_ms": ablationMetricValue(proof.Metrics.Elapsed, false)}
	return outcome, proof.Contract.SHA256 + "|" + contract.Oracle.SHA256 + "|" + hashString(strings.Join(baseline, "\n")), nil
}

// Output and enumeration roots must remain contained, including existing ancestors.
func ablationOutputPath(root, input string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	path := input
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("ablation path must remain under root")
	}
	probe := root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		probe = filepath.Join(probe, part)
		info, err := os.Lstat(probe)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("ablation symlink path forbidden")
		}
	}
	return path, nil
}
func runSkillAblationCommand(root string, opts options, stdout, stderr io.Writer) int {
	if opts.resultRoot == "" || opts.output == "" {
		fmt.Fprintln(stderr, "skill-eval-ablation requires --result-root and --output")
		return 2
	}
	if _, err := ablationOutputPath(root, filepath.Join(opts.resultRoot, ".enumeration-boundary")); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	files, err := evalFiles(root, opts.resultRoot, "")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(files) > 512 {
		fmt.Fprintln(stderr, "ablation record limit is 512")
		return 1
	}
	reader := &ablationReader{root: root, used: map[string]bool{}}
	report := ablationReport{Schema: 1, Evidence: "imported-capture", Limitation: "Hashes verify integrity and matching identities, not capture authorship. The operator-selected import root is the trust boundary; this reducer executes no imported command and proves no live skill benefit.", Conditions: map[string]int{}, Matched: []ablationGroup{}, Incomplete: []map[string]any{}, Issues: []map[string]string{}}
	candidates := []ablationCandidate{}
	armCounts := map[string]int{}
	runCounts := map[string]int{}
	for _, file := range files {
		c := ablationCandidate{file: file}
		path, e := resolveModelTierFile(root, file)
		if e == nil {
			reader.used[strings.ToLower(path)] = true
			var raw []byte
			raw, e = readSafeBoundedModelTierFile(path)
			if e == nil {
				e = decodeAblationJSON(path, raw, &c.row)
			}
		}
		if e == nil && (c.row.Schema != 1 || c.row.Kind != "live" || !ablationIdentityValid(c.row.Identity)) {
			e = fmt.Errorf("not an admissible live capture record")
		}
		if e == nil {
			c.key = ablationKey(c.row.Identity)
			armCounts[c.key+"|"+c.row.Identity.Condition]++
			runCounts[c.row.Identity.RunID]++
		}
		c.err = e
		candidates = append(candidates, c)
	}
	groups := map[string][]ablationCandidate{}
	for _, c := range candidates {
		if c.err == nil && (armCounts[c.key+"|"+c.row.Identity.Condition] > 1 || runCounts[c.row.Identity.RunID] > 1) {
			c.err = fmt.Errorf("duplicate run or arm invalidates every affected record")
		}
		if c.err == nil {
			c.outcome, c.comparison, c.err = validateAblation(reader, c.row)
		}
		if c.err != nil {
			report.Excluded++
			report.Issues = append(report.Issues, map[string]string{"record": filepath.Base(c.file), "reason": c.err.Error()})
			continue
		}
		report.Eligible++
		report.Conditions[c.row.Identity.Condition]++
		groups[c.key] = append(groups[c.key], c)
	}
	keys := []string{}
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		rows := groups[key]
		arms := map[string]ablationOutcome{}
		comparison := rows[0].comparison
		compatible := true
		for _, c := range rows {
			arms[c.row.Identity.Condition] = c.outcome
			if c.comparison != comparison {
				compatible = false
			}
		}
		if len(arms) != 3 || !compatible {
			missing := []string{}
			for _, condition := range []string{"full", "reduced", "none"} {
				if _, ok := arms[condition]; !ok {
					missing = append(missing, condition)
				}
			}
			report.Incomplete = append(report.Incomplete, map[string]any{"group": key, "missing": missing, "matching_check_oracle_native_instructions": compatible})
			continue
		}
		id := rows[0].row.Identity
		id.RunID = ""
		id.Condition = ""
		group := ablationGroup{Identity: id, Arms: arms, Differences: map[string]map[string]*float64{}, Regressions: []string{}}
		for _, condition := range []string{"reduced", "none"} {
			full, other := arms["full"], arms[condition]
			delta := 0.0
			if other.Success {
				delta++
			}
			if full.Success {
				delta--
			}
			differences := map[string]*float64{"success": &delta}
			for name, a := range full.Metrics {
				b := other.Metrics[name]
				differences[name] = nil
				if a != nil && b != nil {
					v := *b - *a
					differences[name] = &v
				}
			}
			group.Differences[condition+"_minus_full"] = differences
			if full.Success && !other.Success {
				group.Regressions = append(group.Regressions, condition)
			}
		}
		report.Matched = append(report.Matched, group)
	}
	output, err := ablationOutputPath(root, opts.output)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if reader.used[strings.ToLower(output)] {
		fmt.Fprintln(stderr, "report would overwrite an input or evidence artifact")
		return 1
	}
	// A failed/duplicate record may reference evidence we never reached. Never
	// overwrite an existing non-report merely because it was not visited.
	if _, statErr := os.Lstat(output); statErr == nil {
		previous, readErr := readSafeBoundedModelTierFile(output)
		var prior ablationReport
		if readErr != nil || decodeAblationJSON(output, previous, &prior) != nil || prior.Schema != 1 || prior.Evidence != "imported-capture" || prior.Limitation == "" {
			fmt.Fprintln(stderr, "existing output is not an ablation report")
			return 1
		}
	} else if !os.IsNotExist(statErr) {
		fmt.Fprintln(stderr, statErr)
		return 1
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err == nil {
		err = os.MkdirAll(filepath.Dir(output), 0755)
	}
	if err == nil {
		err = os.WriteFile(output, append(raw, '\n'), 0644)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, report)
	} else {
		fmt.Fprintf(stdout, "Skill ablation: eligible=%d excluded=%d matched=%d\n", report.Eligible, report.Excluded, len(report.Matched))
	}
	return 0
}
