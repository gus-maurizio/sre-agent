#!/usr/bin/env bash
DDD="$(date +%Y.%j.%s)"
for i in plug* sre-agent
do
  echo Updating GIT $i
  cd $i
  git add -A
  git commit -m "$(date)"
  git push 
  cd ..
done
