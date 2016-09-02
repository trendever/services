#!/bin/sh

cat > "$1/onbuild.sh" << EOF
  cat > /etc/apk/repositories << APK
http://dl-cdn.alpinelinux.org/alpine/edge/main
http://dl-cdn.alpinelinux.org/alpine/edge/community
http://dl-cdn.alpinelinux.org/alpine/edge/testing
APK
  apk update
  apk upgrade
  apk add graphicsmagick optipng
  rm -rf /var/cache/apk/*
EOF
