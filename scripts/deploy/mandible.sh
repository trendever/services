#!/bin/sh

cat > "$1/onbuild.sh" << EOF
  update-ca-certificates
  cat > /etc/apk/repositories << APK
http://dl-cdn.alpinelinux.org/alpine/edge/main
http://dl-cdn.alpinelinux.org/alpine/edge/community
http://dl-cdn.alpinelinux.org/alpine/edge/testing
APK
  apk update
  apk upgrade
  apk add ca-certificates openssl
  ln -sv /usr/bin/gm /usr/bin/convert
  apk add graphicsmagick optipng libjpeg-turbo-utils exiftool
  rm -rf /var/cache/apk/*
EOF
