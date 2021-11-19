#!/bin/bash

JSON='{}'
for FILE in *.yaml; do
  [[ "$FILE" =~ common.* ]] && continue
  echo $FILE
  JSON="$(jq -s '.[0] * .[1]' <(echo "$JSON") <(yaml2json < "$FILE"))"
done
jq '.info.title="Horizon API" | .info.description="Horizon API" | del(.servers)'<<<"$JSON" > swagger.json