.PHONY: install install-orc install-dev-shim dev build test lint lint-fix schema-check check-test-presence check-coverage check-skills init install-hooks clean help deploy-glue schema-diff schema-apply schema-inspect setup-workbench schema-diff-workbench schema-apply-workbench bootstrap bootstrap-dev bootstrap-test bootstrap-shell uninstall

# Go binary location (handles empty GOBIN)
GOBIN := $(shell go env GOPATH)/bin

# Version info (injected at build time)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS := -X 'github.com/example/orc/internal/version.Commit=$(COMMIT)' \
           -X 'github.com/example/orc/internal/version.BuildTime=$(BUILD_TIME)'

# Default target
.DEFAULT_GOAL := help

#---------------------------------------------------------------------------
# Directory guard for dangerous targets
#---------------------------------------------------------------------------

# Validates we're running from ~/src/orc or ~/wb/*
# Usage: $(call check-dir)
define check-dir
	@if [ "$$(pwd)" != "$$HOME/src/orc" ] && ! echo "$$(pwd)" | grep -qE "^$$HOME/wb/[^/]+$$"; then \
		echo "Error: This command must be run from ~/src/orc or ~/wb/<workbench>"; \
		echo ""; \
		echo "Current directory: $$(pwd)"; \
		echo ""; \
		echo "Expected locations:"; \
		echo "  ~/src/orc     - ORC development"; \
		echo "  ~/wb/<name>   - Workbench implementation"; \
		exit 1; \
	fi
endef

#---------------------------------------------------------------------------
# Installation (global binary + dev shim)
#---------------------------------------------------------------------------

# Full install: binary + dev shim
install: install-orc install-dev-shim
	$(call check-dir)
	@# Register binaries in host manifest
	@mkdir -p ~/.orc
	@MANIFEST="$(HOST_MANIFEST)"; \
	if [ ! -f "$$MANIFEST" ]; then \
		echo '{}' > "$$MANIFEST"; \
	fi; \
	jq --arg orc "$(GOBIN)/orc" --arg dev "$(GOBIN)/orc-dev" \
		'.binaries = [$$orc, $$dev]' "$$MANIFEST" > /tmp/host-manifest.json && \
		mv /tmp/host-manifest.json "$$MANIFEST"
	@echo ""
	@echo "Installed:"
	@echo "  orc      - global binary (production DB)"
	@echo "  orc-dev  - workbench DB shim"

# Install the orc binary
install-orc:
	@echo "Building orc..."
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/orc ./cmd/orc

# Install orc-dev shim for development (requires workbench DB)
install-dev-shim:
	@echo "Installing orc-dev shim..."
	@cp glue/bin/orc-dev $(GOBIN)/orc-dev
	@chmod +x $(GOBIN)/orc-dev
	@echo "orc-dev installed"

#---------------------------------------------------------------------------
# Development (local binary)
#---------------------------------------------------------------------------

# Build local binary for development (preferred command)
dev:
	@echo "Building local ./orc..."
	@go build -ldflags "$(LDFLAGS)" -o orc ./cmd/orc
	@echo "✓ Built ./orc (local development binary)"

# Alias for backwards compatibility
build: dev

#---------------------------------------------------------------------------
# Testing & Maintenance
#---------------------------------------------------------------------------

# Run all tests
test:
	go test ./...

#---------------------------------------------------------------------------
# Linting
#---------------------------------------------------------------------------

# Run all linters (golangci-lint + architecture + schema-check + test checks + skills)
lint: schema-check check-test-presence check-coverage check-skills
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	@golangci-lint run ./...
	@echo "Running architecture lint..."
	@command -v go-arch-lint >/dev/null 2>&1 || { echo "go-arch-lint not installed. Run: go install github.com/fe3dback/go-arch-lint@latest"; exit 1; }
	@go-arch-lint check
	@echo "✓ All linters passed"

# Validate test schemas use the authoritative schema.go
# This prevents schema drift where tests pass but production queries fail
# Protection layers:
#   1. schema-check: Blocks hardcoded CREATE TABLE in tests
#   2. Tests use db.GetSchemaSQL(): SQLite fails if queries reference missing columns
#   3. CI runs both lint (includes schema-check) and test
schema-check:
	@echo "Checking for hardcoded test schemas..."
	@if grep -r "CREATE TABLE IF NOT EXISTS" internal/adapters/sqlite/*_test.go 2>/dev/null | grep -v "^Binary"; then \
		echo "ERROR: Found hardcoded CREATE TABLE in test files"; \
		echo "Tests should use db.GetSchemaSQL() instead"; \
		exit 1; \
	fi
	@echo "✓ No hardcoded test schemas found"
	@echo "Checking testutil uses authoritative schema..."
	@if ! grep -q 'db.GetSchemaSQL()' internal/adapters/sqlite/testutil_test.go; then \
		echo "ERROR: testutil_test.go must use db.GetSchemaSQL()"; \
		exit 1; \
	fi
	@echo "✓ Test setup uses authoritative schema"

# Check that all source files have corresponding test files
check-test-presence:
	@./scripts/check-test-presence.sh

# Check coverage thresholds per package
check-coverage:
	@./scripts/check-coverage.sh

# Check skills have valid frontmatter and are documented
check-skills:
	@./scripts/check-skills.sh

# Run golangci-lint with auto-fix
lint-fix:
	@echo "Running golangci-lint with --fix..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	@golangci-lint run --fix ./...
	@echo "✓ Lint fixes applied"

#---------------------------------------------------------------------------
# Schema Management (Atlas)
#---------------------------------------------------------------------------

# Preview schema changes (diff current DB vs schema.sql)
schema-diff:
	@echo "Comparing current database to schema.sql..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema apply --env local --dry-run

# Apply schema changes from schema.sql to database
schema-apply:
	$(call check-dir)
	@echo "Applying schema.sql to database..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema apply --env local --auto-approve

# Dump current database schema
schema-inspect:
	@echo "Inspecting current database schema..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema inspect --env local

# Schema management for workbench-local database
schema-diff-workbench:
	@if [ ! -f ".orc/workbench.db" ]; then \
		echo "No workbench DB found. Run: make setup-workbench"; \
		exit 1; \
	fi
	@echo "Comparing workbench DB to schema.sql..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema apply --env workbench --dry-run

schema-apply-workbench:
	@if [ ! -f ".orc/workbench.db" ]; then \
		echo "No workbench DB found. Run: make setup-workbench"; \
		exit 1; \
	fi
	@echo "Applying schema.sql to workbench DB..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema apply --env workbench --auto-approve

#---------------------------------------------------------------------------
# Development Environment Setup
#---------------------------------------------------------------------------

# Install git hooks (handles both regular repos and worktrees)
install-hooks:
	@HOOKS_DIR=$$(git rev-parse --git-common-dir)/hooks; \
	mkdir -p "$$HOOKS_DIR"; \
	cp scripts/hooks/pre-commit "$$HOOKS_DIR/pre-commit"; \
	chmod +x "$$HOOKS_DIR/pre-commit"; \
	cp scripts/hooks/post-merge "$$HOOKS_DIR/post-merge"; \
	chmod +x "$$HOOKS_DIR/post-merge"; \
	cp scripts/hooks/post-checkout "$$HOOKS_DIR/post-checkout"; \
	chmod +x "$$HOOKS_DIR/post-checkout"; \
	echo "✓ Git hooks installed to $$HOOKS_DIR"

# Initialize development environment
init: install-hooks
	@echo "✓ ORC development environment initialized"

# First-time setup for new users
bootstrap:
	$(call check-dir)
	@if ! command -v brew >/dev/null 2>&1; then \
		echo "Error: Homebrew is required but not installed."; \
		echo ""; \
		echo "Install Homebrew with:"; \
		echo '  /bin/bash -c "$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'; \
		echo ""; \
		echo "Then run 'make bootstrap' again."; \
		exit 1; \
	fi
	@if [ -d "$$HOME/.orc" ] && [ -f "$$HOME/.orc/orc.db" ]; then \
		echo "Already bootstrapped. Run 'make init' to refresh hooks."; \
	else \
		echo "Bootstrapping ORC..."; \
		echo ""; \
		echo "Installing dependencies via Homebrew..."; \
		brew bundle; \
		echo ""; \
		$(MAKE) init; \
		$(MAKE) install; \
		$(MAKE) deploy-glue; \
		echo ""; \
		echo "Creating ORC directories..."; \
		mkdir -p ~/.orc/ws ~/wb; \
		echo ""; \
		echo "Configuring PATH..."; \
		GOBIN="$$(go env GOPATH)/bin"; \
		PATH_EXPORT="export PATH=\"\$$PATH:\$$(go env GOPATH)/bin\""; \
		PROFILE_ENTRIES=""; \
		if ! grep -q 'GOPATH.*bin' ~/.zprofile 2>/dev/null; then \
			echo "$$PATH_EXPORT" >> ~/.zprofile; \
			echo "  Added to ~/.zprofile (login shells)"; \
			PROFILE_ENTRIES="$$PROFILE_ENTRIES $$HOME/.zprofile"; \
		else \
			echo "  Already in ~/.zprofile"; \
		fi; \
		if ! grep -q 'GOPATH.*bin' ~/.zshrc 2>/dev/null; then \
			echo "$$PATH_EXPORT" >> ~/.zshrc; \
			echo "  Added to ~/.zshrc (interactive shells)"; \
			PROFILE_ENTRIES="$$PROFILE_ENTRIES $$HOME/.zshrc"; \
		else \
			echo "  Already in ~/.zshrc"; \
		fi; \
		export PATH="$$PATH:$$GOBIN"; \
		MANIFEST="$(HOST_MANIFEST)"; \
		if [ -f "$$MANIFEST" ] && [ -n "$$PROFILE_ENTRIES" ]; then \
			PROFILES_JSON=$$(echo "$$PROFILE_ENTRIES" | tr ' ' '\n' | sed '/^$$/d' | jq -R '{file: ., line: "export PATH=...GOPATH/bin"}' | jq -s .); \
			jq --argjson profiles "$$PROFILES_JSON" '.shell_profiles = $$profiles' "$$MANIFEST" > /tmp/host-manifest.json && \
				mv /tmp/host-manifest.json "$$MANIFEST"; \
		fi; \
		echo ""; \
		echo "Creating default factory..."; \
		orc factory create default; \
		echo ""; \
		echo "Registering ORC repository..."; \
		orc repo create orc --path ~/src/orc --default-branch main; \
		echo ""; \
		echo "Running health check..."; \
		orc doctor || true; \
		echo ""; \
		echo "✓ ORC bootstrapped successfully!"; \
		echo ""; \
		echo "Next step: Run 'orc hello' to start the first-run experience"; \
	fi

# Test bootstrap in a fresh macOS VM (requires tart)
bootstrap-test:
	@./scripts/bootstrap-test.sh

# Bootstrap and drop into VM shell for exploration
bootstrap-shell:
	@./scripts/bootstrap-test.sh --shell

# Install development dependencies (VM testing, schema migrations)
bootstrap-dev: bootstrap
	@echo "Installing development dependencies..."
	@brew bundle --file=Brewfile.dev
	@echo ""
	@echo "✓ Development dependencies installed"
	@echo ""
	@echo "Available tools:"
	@echo "  tart     - VM management for bootstrap testing"
	@echo "  sshpass  - SSH password automation"
	@echo "  atlas    - Database schema migrations"

# Setup workbench-local development database
setup-workbench:
	@echo "Creating workbench-local database..."
	@mkdir -p .orc
	@rm -f .orc/workbench.db
	@ORC_DB_PATH=.orc/workbench.db go run ./cmd/orc dev reset --force
	@echo ""
	@echo "✓ Workbench DB created: .orc/workbench.db"
	@echo ""
	@echo "Usage:"
	@echo "  orc-dev ...    → uses this local DB (when present)"
	@echo "  orc ...        → uses production DB"

# Clean build artifacts
clean:
	rm -f orc
	go clean
	@echo "✓ Cleaned local build artifacts"

#---------------------------------------------------------------------------
# Claude Code Integration (Glue)
#---------------------------------------------------------------------------

# Manifest path for tracking all ORC host artifacts
HOST_MANIFEST := $(HOME)/.orc/host-manifest.json

# Deploy skills and hooks to Claude Code
deploy-glue:
	$(call check-dir)
	@mkdir -p ~/.claude/skills ~/.claude/hooks ~/.orc
	@# --- Read old manifest (supports both legacy and new format) ---
	@OLD_SKILLS=""; \
	OLD_HOOKS=""; \
	MANIFEST="$(HOST_MANIFEST)"; \
	LEGACY_MANIFEST="$$HOME/.orc/glue-manifest.json"; \
	if [ -f "$$MANIFEST" ]; then \
		OLD_SKILLS=$$(jq -r '.skills // [] | .[]' "$$MANIFEST" 2>/dev/null); \
		OLD_HOOKS=$$(jq -r '.hooks // [] | .[]' "$$MANIFEST" 2>/dev/null); \
	elif [ -f "$$LEGACY_MANIFEST" ]; then \
		OLD_SKILLS=$$(jq -r '.skills // [] | .[]' "$$LEGACY_MANIFEST" 2>/dev/null); \
		OLD_HOOKS=$$(jq -r '.hooks // [] | .[]' "$$LEGACY_MANIFEST" 2>/dev/null); \
	fi; \
	\
	echo "Deploying Claude Code skills..."; \
	CURRENT_SKILLS=""; \
	SKILLS_PATHS=""; \
	for dir in glue/skills/*/; do \
		name=$$(basename "$$dir"); \
		echo "  → $$name"; \
		rm -rf ~/.claude/skills/$$name; \
		cp -r "$$dir" ~/.claude/skills/$$name; \
		CURRENT_SKILLS="$$CURRENT_SKILLS $$name"; \
		SKILLS_PATHS="$$SKILLS_PATHS $$HOME/.claude/skills/$$name"; \
	done; \
	echo "✓ Skills deployed to ~/.claude/skills/"; \
	\
	echo "Checking for orphan skills..."; \
	for old_skill in $$OLD_SKILLS; do \
		is_current=false; \
		for cur in $$CURRENT_SKILLS; do \
			if [ "$$old_skill" = "$$cur" ]; then \
				is_current=true; \
				break; \
			fi; \
		done; \
		if [ "$$is_current" = "false" ] && [ -d "$$HOME/.claude/skills/$$old_skill" ]; then \
			echo "  ✗ Removing orphan: $$old_skill"; \
			rm -rf "$$HOME/.claude/skills/$$old_skill"; \
		fi; \
	done; \
	\
	if [ -d "glue/hooks" ] && [ "$$(ls -A glue/hooks/*.sh 2>/dev/null)" ]; then \
		echo "Deploying Claude Code hook scripts..."; \
		for hook in glue/hooks/*.sh; do \
			[ -f "$$hook" ] || continue; \
			name=$$(basename "$$hook"); \
			echo "  → $$name"; \
			cp "$$hook" $$HOME/.claude/hooks/$$name; \
			chmod +x $$HOME/.claude/hooks/$$name; \
		done; \
		echo "✓ Hook scripts deployed to ~/.claude/hooks/"; \
	fi; \
	\
	CURRENT_HOOKS=""; \
	if [ -f "glue/hooks.json" ]; then \
		echo "Configuring hooks in settings.json..."; \
		CURRENT_HOOKS=$$(jq -r 'keys[]' glue/hooks.json 2>/dev/null); \
		for old_hook in $$OLD_HOOKS; do \
			is_current=false; \
			for cur in $$CURRENT_HOOKS; do \
				if [ "$$old_hook" = "$$cur" ]; then \
					is_current=true; \
					break; \
				fi; \
			done; \
			if [ "$$is_current" = "false" ]; then \
				echo "  ✗ Removing orphan hook: $$old_hook"; \
				jq "del(.hooks[\"$$old_hook\"])" \
					$$HOME/.claude/settings.json > /tmp/settings.json && \
					mv /tmp/settings.json $$HOME/.claude/settings.json; \
			fi; \
		done; \
		jq -s '.[0].hooks = (.[0].hooks // {}) * .[1] | .[0]' \
			$$HOME/.claude/settings.json glue/hooks.json > /tmp/settings.json && \
			mv /tmp/settings.json $$HOME/.claude/settings.json; \
		echo "✓ Hooks configured in settings.json"; \
	fi; \
	\
	TMUX_SCRIPTS=""; \
	if [ -d "glue/tmux" ] && [ "$$(ls -A glue/tmux 2>/dev/null)" ]; then \
		echo "Deploying tmux scripts..."; \
		mkdir -p $$HOME/.orc/tmux; \
		for script in glue/tmux/*.sh; do \
			[ -f "$$script" ] || continue; \
			name=$$(basename "$$script"); \
			echo "  → $$name"; \
			cp "$$script" $$HOME/.orc/tmux/$$name; \
			chmod +x $$HOME/.orc/tmux/$$name; \
			TMUX_SCRIPTS="$$TMUX_SCRIPTS $$HOME/.orc/tmux/$$name"; \
		done; \
		echo "✓ TMux scripts deployed to ~/.orc/tmux/"; \
	fi; \
	\
	echo "Writing host manifest..."; \
	SKILLS_JSON=$$(echo "$$CURRENT_SKILLS" | tr ' ' '\n' | sed '/^$$/d' | jq -R . | jq -s .); \
	HOOKS_JSON=$$(echo "$$CURRENT_HOOKS" | tr ' ' '\n' | sed '/^$$/d' | jq -R . | jq -s .); \
	SKILLS_PATHS_JSON=$$(echo "$$SKILLS_PATHS" | tr ' ' '\n' | sed '/^$$/d' | jq -R . | jq -s .); \
	TMUX_JSON=$$(echo "$$TMUX_SCRIPTS" | tr ' ' '\n' | sed '/^$$/d' | jq -R . | jq -s .); \
	EXISTING="{}"; \
	if [ -f "$$MANIFEST" ]; then \
		EXISTING=$$(cat "$$MANIFEST"); \
	fi; \
	echo "$$EXISTING" | jq \
		--argjson skills "$$SKILLS_JSON" \
		--argjson hooks "$$HOOKS_JSON" \
		--argjson skill_paths "$$SKILLS_PATHS_JSON" \
		--argjson tmux_scripts "$$TMUX_JSON" \
		'. + {skills: $$skills, hooks: $$hooks, skill_paths: $$skill_paths, tmux_scripts: $$tmux_scripts, directories: (((.directories // []) + ["'"$$HOME"'/.orc", "'"$$HOME"'/.orc/tmux"]) | unique)}' \
		> /tmp/host-manifest.json && \
		mv /tmp/host-manifest.json "$$MANIFEST"; \
	echo "✓ Manifest written to $$MANIFEST"; \
	\
	if [ -f "$$LEGACY_MANIFEST" ]; then \
		rm -f "$$LEGACY_MANIFEST"; \
		echo "✓ Removed legacy glue-manifest.json"; \
	fi

#---------------------------------------------------------------------------
# Uninstall
#---------------------------------------------------------------------------

# Remove all ORC host artifacts tracked in the manifest
uninstall:
	@MANIFEST="$(HOST_MANIFEST)"; \
	if [ ! -f "$$MANIFEST" ]; then \
		echo "No host manifest found at $$MANIFEST"; \
		echo "Nothing to uninstall."; \
		exit 0; \
	fi; \
	echo "Uninstalling ORC..."; \
	echo ""; \
	\
	echo "Removing binaries..."; \
	for bin in $$(jq -r '.binaries // [] | .[]' "$$MANIFEST" 2>/dev/null); do \
		if [ -f "$$bin" ]; then \
			rm -f "$$bin"; \
			echo "  ✗ $$bin"; \
		fi; \
	done; \
	\
	echo "Removing skill directories..."; \
	for skill_path in $$(jq -r '.skill_paths // [] | .[]' "$$MANIFEST" 2>/dev/null); do \
		if [ -d "$$skill_path" ]; then \
			rm -rf "$$skill_path"; \
			echo "  ✗ $$skill_path"; \
		fi; \
	done; \
	\
	echo "Removing tmux scripts..."; \
	for script in $$(jq -r '.tmux_scripts // [] | .[]' "$$MANIFEST" 2>/dev/null); do \
		if [ -f "$$script" ]; then \
			rm -f "$$script"; \
			echo "  ✗ $$script"; \
		fi; \
	done; \
	\
	echo "Cleaning ORC hooks from settings.json..."; \
	SETTINGS="$$HOME/.claude/settings.json"; \
	if [ -f "$$SETTINGS" ]; then \
		for hook_key in $$(jq -r '.hooks // [] | .[]' "$$MANIFEST" 2>/dev/null); do \
			if jq -e ".hooks[\"$$hook_key\"]" "$$SETTINGS" >/dev/null 2>&1; then \
				jq "(.hooks[\"$$hook_key\"] // []) |= [.[] | select((.hooks // []) | all(.command | test(\"orc hook\") | not))] | if .hooks[\"$$hook_key\"] == [] then del(.hooks[\"$$hook_key\"]) else . end" \
					"$$SETTINGS" > /tmp/settings.json && \
					mv /tmp/settings.json "$$SETTINGS"; \
				echo "  ✗ Filtered orc hooks from $$hook_key"; \
			fi; \
		done; \
	fi; \
	\
	echo "Cleaning shell profile..."; \
	for profile_entry in $$(jq -r '.shell_profiles // [] | .[].file' "$$MANIFEST" 2>/dev/null); do \
		if [ -f "$$profile_entry" ]; then \
			sed -i '' '/GOPATH.*bin/d' "$$profile_entry" 2>/dev/null || true; \
			echo "  ✗ Removed PATH line from $$profile_entry"; \
		fi; \
	done; \
	\
	echo ""; \
	if [ -f "$$HOME/.orc/orc.db" ]; then \
		echo "WARNING: ~/.orc/orc.db contains your data and was NOT deleted."; \
		echo "  Remove it manually if you want a full cleanup:"; \
		echo "    rm ~/.orc/orc.db"; \
		echo ""; \
	fi; \
	\
	rm -f "$$MANIFEST"; \
	rm -f "$$HOME/.orc/glue-manifest.json"; \
	echo "  ✗ Removed host manifest"; \
	\
	if [ -d "$$HOME/.orc" ] && [ -z "$$(ls -A "$$HOME/.orc" 2>/dev/null | grep -v '^orc\.db$$' | grep -v '^\.')" ]; then \
		if [ ! -f "$$HOME/.orc/orc.db" ]; then \
			rmdir "$$HOME/.orc" 2>/dev/null && echo "  ✗ Removed empty ~/.orc/" || true; \
		fi; \
	fi; \
	\
	echo ""; \
	echo "✓ ORC uninstalled"

#---------------------------------------------------------------------------
# Help
#---------------------------------------------------------------------------

help:
	@echo "ORC Makefile Commands:"
	@echo ""
	@echo "Getting Started:"
	@echo "  make bootstrap       First-time setup (new users start here)"
	@echo "  make bootstrap-test  Test bootstrap in fresh macOS VM (requires tart)"
	@echo ""
	@echo "Development:"
	@echo "  make dev           Build local ./orc for development"
	@echo "  make test          Run all tests"
	@echo "  make lint          Run golangci-lint + architecture + schema-check"
	@echo "  make lint-fix      Run golangci-lint with auto-fix"
	@echo "  make schema-check  Verify test files use authoritative schema"
	@echo "  make clean         Remove local build artifacts"
	@echo ""
	@echo "Schema Management (Atlas):"
	@echo "  make schema-diff            Preview schema changes (production DB vs schema.sql)"
	@echo "  make schema-apply           Apply schema.sql to production database"
	@echo "  make schema-inspect         Dump current production database schema"
	@echo "  make schema-diff-workbench  Preview schema changes (workbench DB vs schema.sql)"
	@echo "  make schema-apply-workbench Apply schema.sql to workbench database"
	@echo "  make setup-workbench        Create/reset workbench-local database"
	@echo ""
	@echo "Installation:"
	@echo "  make install       Install orc binary and orc-dev shim"
	@echo "  make install-orc   Install only the orc binary"
	@echo "  make uninstall     Remove all ORC host artifacts"
	@echo "  make init          Refresh git hooks (after pull)"
	@echo ""
	@echo "Claude Code Integration:"
	@echo "  make deploy-glue   Deploy skills to ~/.claude/skills/"
