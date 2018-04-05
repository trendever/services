FROM debian:stable

RUN apt-get update
RUN apt-get -y install imagemagick optipng libjpeg-progs exiv2 libimage-exiftool-perl && \
      ln -sfv /usr/bin/convert-im6 /usr/bin/convert
