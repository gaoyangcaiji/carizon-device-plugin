FROM hub.hobot.cc/library/centos:7.6-extra

COPY horizon-device-plugin /usr/bin/horizon-device-plugin

ENTRYPOINT ["horizon-device-plugin"]
