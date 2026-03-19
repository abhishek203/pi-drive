# Autoresearch: PiDrive Codebase Optimization

## Objective
Improve the PiDrive codebase through code quality improvements, performance optimizations, and best practices.

## Primary Metric
- **issue_count**: Number of static analysis issues found (lower is better)
  - Measured via `staticcheck ./...` and `go vet ./...`

## Secondary Metrics
- **binary_kb**: CLI binary size in KB (lower is better)
- **build_ms**: Build time in milliseconds (lower is better)

## Benchmark Command
```bash
./autoresearch.benchmark.sh
```

## Rules
1. **No cheating**: Don't disable linters, suppress warnings, or add `//nolint` comments just to reduce issue count
2. **Fix root causes**: Address the actual code issues, don't work around them
3. **Maintain functionality**: All changes must preserve existing behavior
4. **Don't overfit**: Focus on genuine improvements, not gaming the metrics
5. **Test changes**: Ensure code still compiles and builds correctly

## Areas to Optimize
- Fix static analysis warnings
- Remove dead code
- Improve error handling
- Optimize imports
- Reduce binary size
- Add proper documentation
- Improve code structure

## Constraints
- Code must compile successfully
- Existing API contracts must be preserved
- No removal of functional code
