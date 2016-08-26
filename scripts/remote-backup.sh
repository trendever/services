#!/bin/bash

if [ "$#" -lt 2 ]; then
  echo "Usage: $0 host --pgdump --params"
  exit
fi

host="$1"; shift

ssh $host "pg_dump $@ | gzip"  

