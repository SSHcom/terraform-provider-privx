#!/usr/bin/env sh
set -eu

suite_start=$(date +%s)
suite_end=$suite_start   # ensure defined even if we break early

slowest_test=""
slowest_time=0
timings_file="/tmp/acc-test-times.$$"
: > "$timings_file"

while IFS= read -r test; do
  echo

  # trim whitespace
  test_stripped=$(printf '%s' "$test" | tr -d '[:space:]')
  [ -z "$test_stripped" ] && continue

  # skip comments
  case "$test" in
    [[:space:]]\#* | \#* | [[:space:]]//* | //* )
      continue
      ;;
  esac

  # stop running if the line is exactly "exit"
  if [ "$test_stripped" = "exit" ]; then
    echo "Exit marker found. Stopping tests."
    break
  fi

  # sanitize for filename (avoid spaces, slashes, etc.)
  safe_test=$(printf '%s' "$test" | LC_ALL=C tr -cs 'A-Za-z0-9._-' '_')

  test_start=$(date +%s)

  ./scripts/godotenv.sh -f .env -- sh -c \
    "TF_ACC=1 TF_LOG=TRACE TF_LOG_PATH=/tmp/terraform-${safe_test}.log \
     go test -count=1 ./internal/provider -v -run '$test'"

  test_end=$(date +%s)
  test_elapsed=$((test_end - test_start))

  # track slowest
  if [ "$test_elapsed" -gt "$slowest_time" ]; then
    slowest_time="$test_elapsed"
    slowest_test="$test"
  fi

  # store for sorted summary
  echo "$test_elapsed|$test" >> "$timings_file"

  test_end=$(date +%s)
  test_elapsed=$((test_end - test_start))

  now=$(date +%s)
  suite_so_far=$((now - suite_start))

  printf '=== FINISHED %s in %02d:%02d | total time: %02d:%02d:%02d\n' \
    "$test" \
    $((test_elapsed / 60)) \
    $((test_elapsed % 60)) \
    $((suite_so_far / 3600)) \
    $((suite_so_far % 3600 / 60)) \
    $((suite_so_far % 60))

done < ./scripts/acceptance-tests.txt

suite_end=$(date +%s)
suite_elapsed=$((suite_end - suite_start))

printf '\n=============================\n'
printf 'Total suite time: %02d:%02d:%02d\n' \
  $((suite_elapsed / 3600)) \
  $((suite_elapsed % 3600 / 60)) \
  $((suite_elapsed % 60))

if [ -n "$slowest_test" ]; then
  printf 'Slowest test: %s (%02d:%02d)\n' \
    "$slowest_test" \
    $((slowest_time / 60)) \
    $((slowest_time % 60))
else
  printf 'Slowest test: (none)\n'
fi

printf '\n⏱  Test duration summary (slowest → fastest):\n'

sort -nr "$timings_file" | while IFS='|' read -r secs name; do
  printf '  %02d:%02d  %s\n' \
    $((secs / 60)) \
    $((secs % 60)) \
    "$name"
done

rm -f "$timings_file"
printf '=============================\n'
