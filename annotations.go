package action

import (
	"log/slog"
	"strings"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"

	"action/go-gha-wrapper/sdk"
)

// printNoticeWarning will find noteworthy packages and print a notice or warning for the end user.
//   - filepath should be a file likely updated by the PR Usually the lock file.
//     (For PR only, doesn't matter for push event)
//   - line should be a line in the provided filepath likely updated by the PR
//     (For PR only, doesn't matter for push event).
func printNoticeWarning(diffMap contract.DiffMap, filepath string, line int) {
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

	if len(noticePkgs) > 0 {
		body := buildAnnotationBody(
			"Following packages are unchanged but abandoned ("+summary.AbandonedSymbol+") "+
				"and/or not using a semver version ("+summary.NonSemverSymbol+"):",
			noticePkgs,
		)
		sdk.NoticeAnnotation(body, filepath, "Noteworthy unchanged packages", line)
	}

	if len(warningPkgs) > 0 {
		body := buildAnnotationBody(
			"Following packages have been updated and are abandoned ("+summary.AbandonedSymbol+") "+
				"and/or not using a semver version ("+summary.NonSemverSymbol+"):",
			warningPkgs,
		)
		sdk.WarningAnnotation(body, filepath, "Noteworthy changed packages", line)
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
