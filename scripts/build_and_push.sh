#!/bin/zsh

set -e

if [ $# -ne 1 ]; then
  echo "Usage: $0 <tag>"
  exit 1
fi

TAG=$1
FRONTEND_IMAGE_NAME="sakuheinonen/todo-frontend"
IMAGE_SERVICE_NAME="sakuheinonen/image-service"
TODO_SERVICE_NAME="sakuheinonen/todo-service"

docker build -t ${FRONTEND_IMAGE_NAME}:${TAG} ./frontend && docker push ${FRONTEND_IMAGE_NAME}:${TAG}
docker build -t ${IMAGE_SERVICE_NAME}:${TAG} ./backend/image-todo-service/ && docker push ${IMAGE_SERVICE_NAME}:${TAG}
docker build -t ${TODO_SERVICE_NAME}:${TAG} ./backend/todo-service/ && docker push ${TODO_SERVICE_NAME}:${TAG}

