#!/usr/bin/env bash
set -euo pipefail # Strict

{
  echo "/*"

  prefix=""
  while IFS= read -r line || [ -n "$line" ]; do
    if [[ $line == \`\`\`* ]]; then
      if [[ $prefix == "" ]]; then
        prefix="	"
      else
        prefix=""
      fi
    else
      echo "$prefix$line"
    fi
  done <"README.md"

  echo "*/"

  echo "package twigots"
} >"doc.go"
