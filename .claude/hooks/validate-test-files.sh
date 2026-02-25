#!/bin/bash
# PostToolUse hook: reminds agents to create test files for new .go/.tsx files.
# Outputs additionalContext (non-blocking) when a test file is missing.

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Exit early if no file path
if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# --- Go files ---
if [[ "$FILE_PATH" == *.go ]] && [[ "$FILE_PATH" != *_test.go ]]; then
  # Skip files that don't need tests
  if [[ "$FILE_PATH" == */migrations/* ]] || \
     [[ "$FILE_PATH" == */models/* ]] || \
     [[ "$FILE_PATH" == */main.go ]] || \
     [[ "$FILE_PATH" == */settings.go ]] || \
     [[ "$FILE_PATH" == */router.go ]]; then
    exit 0
  fi

  TEST_FILE="${FILE_PATH%.go}_test.go"
  if [ ! -f "$TEST_FILE" ]; then
    cat <<EOF
{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"WARNING: Created ${FILE_PATH} but no test file exists at ${TEST_FILE}. Every new .go file needs a corresponding _test.go file."}}
EOF
  fi
  exit 0
fi

# --- React/TypeScript component files ---
if [[ "$FILE_PATH" == *.tsx ]] && [[ "$FILE_PATH" != *.test.tsx ]]; then
  # Skip API routes and pages — only components need snapshot tests
  if [[ "$FILE_PATH" == */pages/api/* ]] || \
     [[ "$FILE_PATH" == */pages/* && "$FILE_PATH" != */packages/* ]]; then
    exit 0
  fi

  DIR=$(dirname "$FILE_PATH")
  BASENAME=$(basename "$FILE_PATH" .tsx)
  TEST_FILE="${DIR}/__tests__/${BASENAME}.test.tsx"

  if [ ! -f "$TEST_FILE" ]; then
    cat <<EOF
{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"WARNING: Created ${FILE_PATH} but no snapshot test exists at ${TEST_FILE}. Every new component needs a snapshot test."}}
EOF
  fi
  exit 0
fi

exit 0
