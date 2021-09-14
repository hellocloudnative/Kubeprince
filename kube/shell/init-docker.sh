# Install docker
chmod a+x docker.sh
sh  docker.sh


# 修改kubelet
mkdir -p /etc/systemd/system/kubelet.service.d
chmod a+x init-kube-docker.sh
sh init-kube-docker.sh
