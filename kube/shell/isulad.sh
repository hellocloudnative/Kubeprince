#!/bin/sh
set -x
if ! [ -x /usr/local/bin/isula  ]; then
  tar  -xvzf ../isulad/isulad-deb-amd64.tar.gz     -C /    &&  dpkg -i  /deb-amd64/*.deb
  systemctl enable  isulad.service
  systemctl restart isulad.service
fi
  mkdir -p /opt/cni/bin   &&  tar -xvzf ../isulad/cni-plugins-linux-amd64.tar.gz  -C /opt/cni/bin 
# 已经安装了isulad并且运行了, 就不去重启.
isula  version || systemctl restart isulad.service
cat  > /etc/isulad/daemon.json  << eof
{
    "group": "isula",
    "default-runtime": "lcr",
    "graph": "/var/lib/isulad",
    "state": "/var/run/isulad",
    "engine": "lcr",
    "log-level": "ERROR",
    "pidfile": "/var/run/isulad.pid",
    "log-opts": {
        "log-file-mode": "0600",
        "log-path": "/var/lib/isulad",
        "max-file": "1",
        "max-size": "30KB"
    },
    "log-driver": "stdout",
    "container-log": {
        "driver": "json-file"
    },
    "hook-spec": "/etc/default/isulad/hooks/default.json",
    "start-timeout": "2m",
    "storage-driver": "overlay2",
    "storage-opts": [
        "overlay2.override_kernel_check=true"
    ],
    "registry-mirrors": [
	    "docker.io"
    ],
    "insecure-registries": [
            "harbor.sh.deepin.com"
    ],
    "pod-sandbox-image": "harbor.sh.deepin.com/amd64/pause:uos",
    "image-server-sock-addr": "unix:///var/run/isulad.sock",
    "native.umask": "secure",
    "network-plugin": "cni",
    "cni-bin-dir": "/opt/cni/bin",
    "cni-conf-dir": "/etc/cni/net.d",
    "image-layer-check": false,
    "use-decrypted-key": true,
    "insecure-skip-verify-enforce": false
}

eof

for image_name in $(ls ../images/)

do

  isula  load   -i  ../images/${image_name} || true

done

cat > /etc/crictl.yaml  << eof
runtime-endpoint: unix:///var/run/isulad.sock
eof
