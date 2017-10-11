#!/bin/sh

cat > "$1/onbuild.sh" << EOF
  update-ca-certificates
  cat > /etc/apk/repositories << APK
http://dl-cdn.alpinelinux.org/alpine/latest-stable/main
http://dl-cdn.alpinelinux.org/alpine/latest-stable/community
http://dl-cdn.alpinelinux.org/alpine/latest-stable/releases
APK
  apk update
  apk upgrade
  apk add ca-certificates openssl
  ln -sv /usr/bin/gm /usr/bin/convert
  apk add graphicsmagick optipng libjpeg-turbo-utils exiftool
  rm -rf /var/cache/apk/*
EOF
