#!/bin/bash

# params

chunkers=(
  "size-262144" # default
  "rabin-512-1024-2048"
  "rabin-512-1024-65536"
)

layouts=(
  "" # default
  "--trickle"
)

datastores=(
  "flatfs"
  "badgerds"
)

# code

USAGE=$(basename $0)" <path>"

usage() {
  echo "$USAGE"
  echo "benchmark ipfs with directory"
  echo ""
  echo "datastores: ${datastores[@]}"
  echo "chunkers: ${chunkers[@]}"
  echo "layouts: ${layouts[@]}"
  exit 0
}

die() {
  echo >&2 "error: $@"
  exit 1
}

log() {
  if [ $verbose ]; then
    echo >&2 "$@"
  fi
}

# main
testdir="$1"

if [ "$testdir" = "" ]; then
  usage
fi

if [[ ! -e "$testdir" ]]; then
  die "error: $testdir does not exist"
fi

techo() {
  echo $@
}

trace() {
  techo '```'
  echo "> $@"
  $@ 2>&1
  techo '```'
}

silent() {
  $@ >/dev/null 2>/dev/null
}

ipfs="ipfs"

techo "---"
techo "##" $(date -u +"%Y-%m-%d %H:%M:%SZ")
techo "benchmarking ipfs with directory: \`$testdir\`"

trace du -sh "$testdir"
# trace "sh -c 'find $testdir | wc'"

for d in ${datastores[@]}; do
  for c in ${chunkers[@]}; do
    for l in ${layouts[@]}; do

      h=$(echo "$d $c $l" | multihash -l 80)
      repo="ipfs-repo-$h"

      techo "### {$d, $c, $l}"
      techo "benchmark options:"
      techo "- datastore: $d"
      techo "- chunker: $c"
      techo "- layout: $l"
      techo "- repo: $repo"

      # init
      silent export IPFS_PATH="$repo"

      if [ "$d" == "flatfs" ]; then
        profile=""
      else
        profile="--profile=$d"
      fi

      silent $ipfs init "$profile" # dont trace

      # add
      tstart=`date +%s`
      trace $ipfs add -Q "--chunker=$c" "$l" -r "$testdir"
      tend=`date +%s`

      # stats
      techo "runtime:" $((tend-tstart))
      trace du -sh "$repo"
      trace $ipfs repo stat

      rm -rf "$repo"
    done
  done
done
