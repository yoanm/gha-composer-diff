package ghaction

import (
	"log/slog"
	"strings"

	"ghaction/go-gha-wrapper/ghasdk"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"
)

func (gha *Action) printNoticeWarning(diffMap contract.DiffMap) {
	slog.Info("Managing annotations...")
	// Notice for unchanged abandoned packages or unchanged package with non-semver version
	// Warning for added/updated abandoned packages and added/updated packages with non-semver version
	var (
		noticePkgs  []*contract.PackageChange
		warningPkgs []*contract.PackageChange
	)

	for _, chg := range diffMap {
		if chg.Package.GetVersion().Semver == nil || chg.Package.IsAbandoned() {
			if chg.Operation.Name == contract.NoChangeOperation {
				noticePkgs = append(noticePkgs, chg)
			} else if chg.Operation.Name != contract.RemovalOperation {
				warningPkgs = append(warningPkgs, chg)
			}
		}
	}

	// Composer specific: 7th line is usually the "content-hash" line in the composer lock.
	// Most time it will be updated so annotations will show up on the diff. It should at least be around the
	// content-hash line if not exactly on it.
	// (Annotations work also if attached to unchanged line, but are less noticeable on the diff UI)
	filepath, line := gha.paths.lock, 7

	if len(noticePkgs) > 0 {
		body := buildAnnotationBody(
			"Following packages are unchanged but abandoned ("+summary.AbandonedSymbol+") "+
				"and/or not using a semver version ("+summary.NonSemverSymbol+"):",
			noticePkgs,
		)
		ghasdk.NoticeAnnotation(body, filepath, "Noteworthy unchanged packages", line)
	}

	if len(warningPkgs) > 0 {
		body := buildAnnotationBody(
			"Following packages have been updated and are abandoned ("+summary.AbandonedSymbol+") "+
				"and/or not using a semver version ("+summary.NonSemverSymbol+"):",
			warningPkgs,
		)
		ghasdk.WarningAnnotation(body, filepath, "Noteworthy changed packages", line)
	}
}

func buildAnnotationBody(header string, changeList []*contract.PackageChange) string {
	builder := strings.Builder{}
	builder.WriteString(header + "\n")

	for _, chg := range changeList {
		abandonedSymbol := ""
		if chg.Package.IsAbandoned() {
			abandonedSymbol = summary.AbandonedSymbol
		}

		builder.WriteString(
			" - " + summary.GetPackageSymbol(chg.Package) + chg.Package.GetName() + abandonedSymbol + " " +
				summary.BuildVersionLabel(chg.Package.GetVersion()),
		)
	}

	return builder.String()
}
