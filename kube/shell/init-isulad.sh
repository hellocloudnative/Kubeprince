# Install isulad
chmod a+x isulad.sh
sh  isulad.sh


# 修改kubelet
mkdir -p /etc/systemd/system/kubelet.service.d
cat > /etc/systemd/system/kubelet.service.d/isulad.conf << eof
[Service]
Environment="KUBELET_EXTRA_ARGS=--container-runtime=remote  --runtime-request-timeout=15m --container-runtime-endpoint=unix:///var/run/isulad.sock --image-service-endpoint=unix:///var/run/isulad.sock"
eof

chmod a+x init-kube-isulad.sh
sh init-kube-isulad.sh
