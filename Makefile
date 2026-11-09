# Differential checks against optional reference tools.
# Missing tools are skipped (see scripts/diffharness.py).
.PHONY: test diff diff-round

test:
	go test ./...

diff:
	python3 scripts/diffharness.py --bin /tmp/goword-text testdata/word

# Compare two goword binaries (before and after) and write a JSON report.
diff-round:
	python3 scripts/diffharness.py --bin /tmp/goword-text --before-bin /tmp/goword-before-text --report testdata/diff-report.json testdata/word
