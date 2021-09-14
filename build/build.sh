#!/bin/bash

# Environment Variables:
# - REPOSITORY_PREFIX

export REPOSITORY_PREFIX="${REPOSITORY_PREFIX:-"harbor.mockserver.org/staffyun163music/cloudnative/library/"}"

build(){
  local DOCKERFILE_DIR="$1" && shift
  local DOCKERFILE_PATH="$DOCKERFILE_DIR/Dockerfile"
  local COMPONENT="${DOCKERFILE_DIR##*/}" && [[ "$COMPONENT" =~ .*horizon.* ]] || COMPONENT="horizon-$COMPONENT"
  local CONTEXT NO_PUSH
  while ARG="$1" && shift; do
    case "$ARG" in
    "--context")
      CONTEXT="$1" && shift || break
      ;;
    "--no-push")
      NO_PUSH="y"
      ;;
    *)
      shift
      ;;
    esac
  done
  [[ ! -z "$CONTEXT" ]] || CONTEXT="$PWD"
  local IMAGE="${REPOSITORY_PREFIX%/}/$COMPONENT:$APP_REVISION"
  docker build --network=host --rm -t "$IMAGE" -f "$DOCKERFILE_PATH" "$CONTEXT" || exit 1
  [[ "$NO_PUSH" == "y" ]] || {
    docker push "$IMAGE" || exit 1
    docker rmi "$IMAGE" -f
  }
}

main(){
    local SCRIPT="${BASH_SOURCE[0]}" && [[ -L "$SCRIPT" ]] && SCRIPT="$(readlink -f "$SCRIPT")"
    local SCRIPT_DIR="$(cd "$(dirname $SCRIPT)"; pwd)" SUB_PATH
    for SUB in $(ls "$SCRIPT_DIR"); do
      SUB_PATH="$SCRIPT_DIR/$SUB"
      [[ -f "$SUB_PATH" ]] && continue
      # gateway镜像固定，不需要每次都打镜像
      [[ "$SUB" == "gateway" ]] && continue
      build "$SUB_PATH" "$@"
    done
}

main "$@"