#!/bin/bash
if ! [ -x /usr/bin/docker  ]; then
  tar  -xvzf ../docker/docker-deb-amd64.tar.gz     -C /    &&  dpkg -i  /deb-amd64/*.deb
  systemctl enable  docker.service
  systemctl restart docker.service
fi
mkdir -p /opt/cni/bin   &&  tar -xvzf ../isulad/cni-plugins-linux-amd64.tar.gz  -C /opt/cni/bin
cat > /etc/docker/daemon.json <<EOF
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ]
}
EOF
for image_name in $(ls ../images/)

do

   docker  load   -i  ../images/${image_name} || true

done

cat > /etc/crictl.yaml  << eof
runtime-endpoint: unix:///var/run/docker.sock
eof
systemctl daemon-reload

systemctl restart docker
