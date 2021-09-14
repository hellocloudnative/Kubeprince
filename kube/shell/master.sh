kubeadm init --config ../conf/kubeadm.yaml
mkdir ~/.kube
cp /etc/kubernetes/admin.conf ~/.kube/config
kubectl taint nodes --all node-role.kubernetes.io/master-
kubectl apply -f ../conf/calico.yaml
sleep 5
kubectl apply -f ../conf/calico.yaml
