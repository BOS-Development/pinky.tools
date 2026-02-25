#!/bin/bash
# SubagentStop hook: checks agent output for signs of missing conventions.
# Injects additionalContext when the agent hit issues that suggest
# its instructions need updating.

INPUT=$(cat)
AGENT_TYPE=$(echo "$INPUT" | jq -r '.agent_type // empty')
LAST_MSG=$(echo "$INPUT" | jq -r '.last_assistant_message // empty')

# Only review our dev agents
if [[ "$AGENT_TYPE" != "backend-dev" ]] && [[ "$AGENT_TYPE" != "frontend-dev" ]]; then
  exit 0
fi

ISSUES=""

# Check for signs the agent was missing context
if echo "$LAST_MSG" | grep -qi "couldn't find\|not sure\|unable to locate\|I don't have"; then
  ISSUES="${ISSUES}Agent reported uncertainty or missing information. "
fi

# Check for signs the agent deviated from patterns
if echo "$LAST_MSG" | grep -qi "workaround\|hack\|temporary\|TODO\|FIXME"; then
  ISSUES="${ISSUES}Agent used workarounds or left TODOs. "
fi

# Check if agent reported test failures
if echo "$LAST_MSG" | grep -qi "test.*fail\|FAIL\|failed"; then
  ISSUES="${ISSUES}Agent reported test failures. "
fi

# Check if agent mentioned missing conventions
if echo "$LAST_MSG" | grep -qi "no existing pattern\|no convention\|wasn't clear"; then
  ISSUES="${ISSUES}Agent noted missing conventions. "
fi

if [ -n "$ISSUES" ]; then
  cat <<EOF
{"hookSpecificOutput":{"hookEventName":"SubagentStop","additionalContext":"AGENT REVIEW (${AGENT_TYPE}): ${ISSUES}Consider updating .claude/agents/${AGENT_TYPE}.md with clearer instructions to prevent this in future tasks."}}
EOF
fi

exit 0
