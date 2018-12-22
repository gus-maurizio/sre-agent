#!/usr/bin/env bash
DDD="$(date +%Y.%j)"
for i in plug* sre-agent
do
  echo Updating GIT $i
  cd $i
  git add -A
  git commit -m "$(date)"
  git tag "${1:-0.9.$DDD}"
  git push --follow-tags
  cd ..
done
