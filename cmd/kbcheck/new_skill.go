package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Adding a skill to this bundle is not one file. Three source-of-truth surfaces
// must agree before two generated artifacts can be rebuilt, and until this
// command existed each one failed separately at the end of a multi-minute
// aggregate run:
//
//	.github/skills/<name>/SKILL.md   the skill itself
//	config/skill-guidance-audit.json exactly one audit row per skill
//	config/skill-categories.json     exactly one display category per skill
//
// then, generated from those:
//
//	packages/universal-ui-skills-contribution/src/catalog.generated.js
//	packages/universal-ui-skills-contribution/release/*.tgz and its lock
//
// This command owns that whole set. `--action check` reports every gap at once
// instead of one per run; `--action apply` registers a new skill across all of
// them in a single step.

const (
	skillCategoriesRelPath = "config/skill-categories.json"
	catalogPackageRelPath  = "packages/universal-ui-skills-contribution"
	generatedCatalogRel    = catalogPackageRelPath + "/src/catalog.generated.js"
	uncategorizedCategory  = "Other"
)

var skillNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type skillCategoryFile struct {
	SchemaVersion int                  `json:"schema_version"`
	Comment       string               `json:"comment,omitempty"`
	Categories    []skillCategoryEntry `json:"categories"`
}

type skillCategoryEntry struct {
	Name   string   `json:"name"`
	Skills []string `json:"skills"`
}

type newSkillRegistration struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	ArgumentHint      string   `json:"argument_hint"`
	Category          string   `json:"category"`
	Purpose           string   `json:"purpose"`
	Callers           []string `json:"callers"`
	RoutingEvidence   string   `json:"routing_evidence"`
	RetainedCapabilit string   `json:"retained_capability"`
	Proof             string   `json:"proof"`
}

type newSkillSurface struct {
	Surface string `json:"surface"`
	Path    string `json:"path"`
	State   string `json:"state"`
	Detail  string `json:"detail,omitempty"`
}

type newSkillResult struct {
	OK          bool              `json:"ok"`
	Action      string            `json:"action"`
	Skill       string            `json:"skill,omitempty"`
	Categories  []string          `json:"available_categories,omitempty"`
	Surfaces    []newSkillSurface `json:"surfaces"`
	Issues      []string          `json:"issues"`
	Regenerated []string          `json:"regenerated,omitempty"`
	NextActions []string          `json:"next_actions,omitempty"`
}

func runNewSkillCommand(root string, opts options, stdout, stderr io.Writer) int {
	action := opts.sliceLeaseAction
	if action == "" {
		action = "check"
	}

	var result newSkillResult
	var err error
	switch action {
	case "check":
		result, err = checkSkillRegistration(root)
	case "apply":
		result, err = applyNewSkill(root, opts)
	default:
		fmt.Fprintln(stderr, "new-skill action must be check or apply")
		return 1
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if opts.json {
		writeJSON(stdout, result)
	} else {
		writeNewSkillText(stdout, result)
	}
	if !result.OK {
		return 1
	}
	return 0
}

func writeNewSkillText(stdout io.Writer, result newSkillResult) {
	fmt.Fprintf(stdout, "Skill registration (%s): %d issues\n", result.Action, len(result.Issues))
	for _, surface := range result.Surfaces {
		detail := ""
		if surface.Detail != "" {
			detail = " :: " + surface.Detail
		}
		fmt.Fprintf(stdout, "%-8s %-18s %s%s\n", surface.State, surface.Surface, surface.Path, detail)
	}
	for _, issue := range result.Issues {
		fmt.Fprintln(stdout, "ERROR "+issue)
	}
	for _, regenerated := range result.Regenerated {
		fmt.Fprintln(stdout, "regenerated "+regenerated)
	}
	for _, next := range result.NextActions {
		fmt.Fprintln(stdout, "next: "+next)
	}
}

func loadSkillCategories(root string) (skillCategoryFile, error) {
	var file skillCategoryFile
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(skillCategoriesRelPath)))
	if err != nil {
		return file, fmt.Errorf("read %s: %w", skillCategoriesRelPath, err)
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		return file, fmt.Errorf("parse %s: %w", skillCategoriesRelPath, err)
	}
	if len(file.Categories) == 0 {
		return file, fmt.Errorf("%s defines no categories", skillCategoriesRelPath)
	}
	return file, nil
}

func categoryNames(file skillCategoryFile) []string {
	names := make([]string, 0, len(file.Categories))
	for _, entry := range file.Categories {
		names = append(names, entry.Name)
	}
	return names
}

func installedSkillNames(root string) ([]string, error) {
	skillRoot := filepath.Join(root, ".github", "skills")
	entries, err := os.ReadDir(skillRoot)
	if err != nil {
		return nil, fmt.Errorf("read skill root: %w", err)
	}
	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// checkSkillRegistration reports every registration gap for every skill in one
// pass, so a missing audit row and an unmapped category surface together rather
// than one per aggregate run.
func checkSkillRegistration(root string) (newSkillResult, error) {
	result := newSkillResult{Action: "check", Surfaces: []newSkillSurface{}, Issues: []string{}}

	names, err := installedSkillNames(root)
	if err != nil {
		return result, err
	}
	categories, err := loadSkillCategories(root)
	if err != nil {
		return result, err
	}
	result.Categories = categoryNames(categories)

	auditRows, err := loadSkillGuidanceAudit(filepath.Join(root, filepath.FromSlash(auditRelPathForNewSkill(root))))
	if err != nil {
		return result, err
	}
	auditCount := map[string]int{}
	for _, row := range auditRows {
		auditCount[row.Name]++
	}

	categoryOf := map[string][]string{}
	for _, entry := range categories.Categories {
		for _, skill := range entry.Skills {
			categoryOf[skill] = append(categoryOf[skill], entry.Name)
		}
	}

	generated, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(generatedCatalogRel)))
	if err != nil {
		return result, fmt.Errorf("read %s: %w", generatedCatalogRel, err)
	}
	generatedText := string(generated)

	for _, name := range names {
		if count := auditCount[name]; count != 1 {
			state := "missing"
			if count > 1 {
				state = "duplicate"
			}
			result.Issues = append(result.Issues,
				fmt.Sprintf("audit-row-%s: %s :: add exactly one row to %s", state, name, skillGuidanceAuditRelPath))
		}
		switch mapped := categoryOf[name]; len(mapped) {
		case 1:
		case 0:
			result.Issues = append(result.Issues,
				fmt.Sprintf("category-missing: %s :: add it to one category in %s, or the catalog silently ships it as %q",
					name, skillCategoriesRelPath, uncategorizedCategory))
		default:
			result.Issues = append(result.Issues,
				fmt.Sprintf("category-duplicate: %s :: mapped to %s in %s", name, strings.Join(mapped, ", "), skillCategoriesRelPath))
		}
		if !strings.Contains(generatedText, `"id": "`+name+`"`) {
			result.Issues = append(result.Issues,
				fmt.Sprintf("catalog-stale: %s is absent from %s :: run npm run catalog:build", name, generatedCatalogRel))
		}
	}

	// A skill listed in a category but no longer on disk would ship a dead row.
	installed := map[string]bool{}
	for _, name := range names {
		installed[name] = true
	}
	for skill := range categoryOf {
		if !installed[skill] {
			result.Issues = append(result.Issues,
				fmt.Sprintf("category-orphan: %s is categorised in %s but has no skill directory", skill, skillCategoriesRelPath))
		}
	}
	if strings.Contains(generatedText, `"category": "`+uncategorizedCategory+`"`) {
		result.Issues = append(result.Issues,
			fmt.Sprintf("catalog-uncategorised: %s contains a %q category :: every skill needs a real category",
				generatedCatalogRel, uncategorizedCategory))
	}

	result.Surfaces = append(result.Surfaces,
		newSkillSurface{"skills", ".github/skills", "checked", fmt.Sprintf("%d skills", len(names))},
		newSkillSurface{"audit", skillGuidanceAuditRelPath, "checked", fmt.Sprintf("%d rows", len(auditRows))},
		newSkillSurface{"categories", skillCategoriesRelPath, "checked", fmt.Sprintf("%d categories", len(categories.Categories))},
		newSkillSurface{"catalog", generatedCatalogRel, "checked", ""},
	)
	sort.Strings(result.Issues)
	result.OK = len(result.Issues) == 0
	return result, nil
}

const skillGuidanceAuditRelPath = "config/skill-guidance-audit.json"

func auditRelPathForNewSkill(root string) string {
	config, err := loadSkillGuidanceConfig(filepath.Join(root, "config", "skill-guidance.json"))
	if err != nil || config.AuditPath == "" {
		return skillGuidanceAuditRelPath
	}
	return config.AuditPath
}

func applyNewSkill(root string, opts options) (newSkillResult, error) {
	result := newSkillResult{Action: "apply", Surfaces: []newSkillSurface{}, Issues: []string{}}

	registration := newSkillRegistration{
		Name:              strings.TrimSpace(opts.newSkillName),
		Description:       strings.TrimSpace(opts.newSkillDescription),
		ArgumentHint:      strings.TrimSpace(opts.newSkillArgumentHint),
		Category:          strings.TrimSpace(opts.newSkillCategory),
		Purpose:           strings.TrimSpace(opts.newSkillPurpose),
		RoutingEvidence:   strings.TrimSpace(opts.newSkillRoutingEvidence),
		RetainedCapabilit: strings.TrimSpace(opts.newSkillRetainedCapability),
		Proof:             strings.TrimSpace(opts.newSkillProof),
	}
	for _, caller := range strings.Split(opts.newSkillCallers, ",") {
		if trimmed := strings.TrimSpace(caller); trimmed != "" {
			registration.Callers = append(registration.Callers, trimmed)
		}
	}
	if len(registration.Callers) == 0 {
		registration.Callers = []string{"user"}
	}

	categories, err := loadSkillCategories(root)
	if err != nil {
		return result, err
	}
	result.Categories = categoryNames(categories)
	result.Skill = registration.Name

	if err := validateNewSkillRegistration(root, registration, categories); err != nil {
		return result, err
	}

	skillDir := filepath.Join(root, ".github", "skills", registration.Name)
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return result, fmt.Errorf("create skill directory: %w", err)
	}
	if err := os.WriteFile(skillFile, []byte(renderSkillTemplate(registration)), 0o644); err != nil {
		return result, fmt.Errorf("write SKILL.md: %w", err)
	}
	result.Surfaces = append(result.Surfaces,
		newSkillSurface{"skills", ".github/skills/" + registration.Name + "/SKILL.md", "created", ""})

	auditPath := filepath.Join(root, filepath.FromSlash(auditRelPathForNewSkill(root)))
	if err := insertSkillAuditRow(auditPath, registration); err != nil {
		return result, err
	}
	result.Surfaces = append(result.Surfaces,
		newSkillSurface{"audit", skillGuidanceAuditRelPath, "updated", "one row added"})

	if err := insertSkillCategory(root, categories, registration); err != nil {
		return result, err
	}
	result.Surfaces = append(result.Surfaces,
		newSkillSurface{"categories", skillCategoriesRelPath, "updated", registration.Category})

	if opts.newSkillSkipGenerated {
		result.NextActions = append(result.NextActions,
			"npm --prefix "+catalogPackageRelPath+" run catalog:build",
			"npm --prefix "+catalogPackageRelPath+" run pack:release")
	} else {
		regenerated, err := regenerateCatalogArtifacts(root)
		if err != nil {
			return result, err
		}
		result.Regenerated = regenerated
		result.Surfaces = append(result.Surfaces,
			newSkillSurface{"catalog", generatedCatalogRel, "regenerated", ""},
			newSkillSurface{"release", catalogPackageRelPath + "/release", "regenerated", ""})
	}

	result.NextActions = append(result.NextActions,
		"write the real body of .github/skills/"+registration.Name+"/SKILL.md",
		"go run ./cmd/kbcheck new-skill --action check")
	result.OK = true
	return result, nil
}

func validateNewSkillRegistration(root string, registration newSkillRegistration, categories skillCategoryFile) error {
	if registration.Name == "" {
		return fmt.Errorf("new-skill apply requires --name")
	}
	if !skillNamePattern.MatchString(registration.Name) {
		return fmt.Errorf("skill name %q must be lowercase kebab-case", registration.Name)
	}
	missing := []string{}
	for field, value := range map[string]string{
		"--description":          registration.Description,
		"--purpose":              registration.Purpose,
		"--routing-evidence":     registration.RoutingEvidence,
		"--retained-capability":  registration.RetainedCapabilit,
		"--proof":                registration.Proof,
		"--category":             registration.Category,
	} {
		if value == "" {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("new-skill apply requires %s; each one is a required registration field, not decoration",
			strings.Join(missing, ", "))
	}
	known := false
	for _, entry := range categories.Categories {
		if entry.Name == registration.Category {
			known = true
		}
		for _, skill := range entry.Skills {
			if skill == registration.Name {
				return fmt.Errorf("skill %q is already categorised under %q", registration.Name, entry.Name)
			}
		}
	}
	if !known {
		return fmt.Errorf("category %q is not defined in %s; known categories: %s",
			registration.Category, skillCategoriesRelPath, strings.Join(categoryNames(categories), ", "))
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "skills", registration.Name)); err == nil {
		return fmt.Errorf("skill directory .github/skills/%s already exists; this command creates a new skill and never overwrites one", registration.Name)
	}
	return nil
}

func renderSkillTemplate(registration newSkillRegistration) string {
	hint := registration.ArgumentHint
	if hint == "" {
		hint = "[request]"
	}
	title := strings.ReplaceAll(registration.Name, "-", " ")
	body := strings.Builder{}
	body.WriteString("---\n")
	body.WriteString("name: " + registration.Name + "\n")
	body.WriteString("description: " + registration.Description + "\n")
	body.WriteString("argument-hint: " + jsonQuote(hint) + "\n")
	body.WriteString("---\n\n")
	body.WriteString("# " + strings.ToUpper(title[:1]) + title[1:] + "\n\n")
	body.WriteString(registration.Purpose + ".\n\n")
	body.WriteString("## Input\n\n<input> #$ARGUMENTS </input>\n\n")
	body.WriteString("## When to use\n\n")
	body.WriteString("Use this when " + strings.ToLower(registration.RoutingEvidence) + ".\n\n")
	body.WriteString("Do not use it when another lane already owns the work.\n\n")
	body.WriteString("## Sequence\n\n")
	body.WriteString("1. Replace this sequence with the real steps.\n")
	body.WriteString("2. Name the exact commands, paths, and skills this lane invokes.\n")
	body.WriteString("3. State what this lane refuses to do and who owns that instead.\n\n")
	body.WriteString("## Proof\n\n")
	body.WriteString(registration.Proof + ".\n")
	return body.String()
}

func jsonQuote(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `"` + value + `"`
	}
	return string(encoded)
}

// insertSkillAuditRow rewrites the audit file in place, keeping its one-row-per-line
// shape so a future diff stays readable.
func insertSkillAuditRow(path string, registration newSkillRegistration) error {
	rows, err := loadSkillGuidanceAudit(path)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Name == registration.Name {
			return fmt.Errorf("audit row for %q already exists", registration.Name)
		}
	}
	rows = append(rows, skillGuidanceAuditRow{
		Name:               registration.Name,
		Purpose:            registration.Purpose,
		Callers:            registration.Callers,
		RoutingEvidence:    registration.RoutingEvidence,
		RetainedCapability: registration.RetainedCapabilit,
		Disposition:        "keep",
		Proof:              registration.Proof,
	})
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		encoded, err := json.Marshal(row)
		if err != nil {
			return err
		}
		lines = append(lines, "  "+string(encoded))
	}
	content := "[\n" + strings.Join(lines, ",\n") + "\n]\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

func insertSkillCategory(root string, categories skillCategoryFile, registration newSkillRegistration) error {
	for index := range categories.Categories {
		if categories.Categories[index].Name != registration.Category {
			continue
		}
		categories.Categories[index].Skills = append(categories.Categories[index].Skills, registration.Name)
		encoded, err := json.MarshalIndent(categories, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(root, filepath.FromSlash(skillCategoriesRelPath)), append(encoded, '\n'), 0o644)
	}
	return fmt.Errorf("category %q disappeared while writing", registration.Category)
}

// regenerateCatalogArtifacts rebuilds the two generated surfaces. Leaving them
// stale is the exact failure this command exists to remove, so a missing Node
// runtime is an error and never a silent skip.
func regenerateCatalogArtifacts(root string) ([]string, error) {
	packageDir := filepath.Join(root, filepath.FromSlash(catalogPackageRelPath))
	regenerated := []string{}
	for _, script := range []string{"build-catalog.mjs", "pack-release.mjs"} {
		command := exec.Command("node", filepath.Join("scripts", script))
		command.Dir = packageDir
		output, err := command.CombinedOutput()
		if err != nil {
			return regenerated, fmt.Errorf("regenerate %s failed: %w :: %s\nrerun with --skip-generated to register the source surfaces only",
				script, err, strings.TrimSpace(string(output)))
		}
		regenerated = append(regenerated, script+": "+strings.TrimSpace(string(output)))
	}
	return regenerated, nil
}
