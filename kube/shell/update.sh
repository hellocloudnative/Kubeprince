#!/bin/sh
set -x
if ! [ -x /usr/local/bin/isula  ]; then
  tar  -xvzf ../isulad/isulad-deb-amd64.tar.gz     -C /    &&  dpkg -i  /deb-amd64/*.deb
  systemctl enable  isulad.service
  systemctl restart isulad.service
fi
#配置
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
# 修改kubelet
mkdir -p /etc/systemd/system/kubelet.service.d
cat > /etc/systemd/system/kubelet.service.d/isulad.conf << eof
[Service]
Environment="KUBELET_EXTRA_ARGS=--container-runtime=remote  --runtime-request-timeout=15m --container-runtime-endpoint=unix:///var/run/isulad.sock --image-service-endpoint=unix:///var/run/isulad.sock"
eof

[ -d /etc/systemd/system/kubelet.service.d ] || mkdir /etc/systemd/system/kubelet.service.d
cp ../conf/10-kubeadm.conf.isulad  /etc/systemd/system/kubelet.service.d/10-kubeadm.conf

systemctl stop docker 

systemctl  restart isulad
systemctl daemon-reload
systemctl  restart kubelet 

bash  killport.sh  10249
bash  killport.sh  6443
bash  killport.sh  2379

systemctl  restart kubelet 
sleep 3
kubectl  delete pod --all -n kube-system
