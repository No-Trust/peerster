# Run gofmt on all go files in directory and subdirectory and alert on non-formatted files.
LIST="$(gofmt -l -s .)"
if [ -n "$LIST" ]; then
  echo "Go code is not formatted, run 'gofmt -s' on:" >&2
  echo "$LIST" >&2
  false
fi
