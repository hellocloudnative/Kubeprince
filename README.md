# kubeprince

## 概要设计
![image-20210901152230166](https://cdn.jsdelivr.net/gh/hellocloudnative/PicGoimages@main/202109/image-20210901152230166.png)

## 节点部署流程图
![image-20210901150732606](https://cdn.jsdelivr.net/gh/hellocloudnative/PicGoimages@main/202109/image-20210901150732606.png)


## 安装前提条件

- 一台或多台运行着下列系统的机器:
  - Uos 20
  - Ubuntu 16.04+
  - Debian 9
  - CentOS 7
  - RHEL 7
- 每台机器 2 GB 或更多的 RAM (如果少于这个数字将会影响您应用的运行内存)
- 2 CPU 核心或更多


### Master 节点

| 规则 | 方向    | 端口范围  | 作用                    | 使用者               |
| :--- | :------ | :-------- | :---------------------- | :------------------- |
| TCP  | Inbound | 6443*     | Kubernetes API server   | All                  |
| TCP  | Inbound | 2379-2380 | etcd server client API  | kube-apiserver, etcd |
| TCP  | Inbound | 10250     | Kubelet API             | Self, Control plane  |
| TCP  | Inbound | 10251     | kube-scheduler          | Self                 |
| TCP  | Inbound | 10252     | kube-controller-manager | Self                 |

### Worker 节点

| 规则 | 方向    | 端口范围    | 作用                | 使用者              |
| :--- | :------ | :---------- | :------------------ | :------------------ |
| TCP  | Inbound | 10250       | Kubelet API         | Self, Control plane |
| TCP  | Inbound | 30000-32767 | NodePort Services** | All                 |

## kubeprince特性与优势：
- 一键安装集群，默认支持iSulad

- 支持离线安装，工具与资源包（二进制程序 配置文件 镜像 yaml文件等）分离

- 使用简单

- 支持自定义配置

- 内核负载均衡

- 支持国产全硬件，全国产CPU架构

| 架构 | 是否支持   |
| :--- | :------ | 
| Amd64  | ✅ | 
| Arm64  | ✅ | 
| Mips64  | ✅ | 
| Sw64  | ✅ | 



## kubeprince安装

- 官方同步服务器时间
- 主机名不可重复
- 安装并启动iSulad(没有会自动安装iSulad)
- 配置免密或者统所有节点密码一致



### 安装单节点集群

```
kubeprince init --master 192.168.0.2 \
    --node 192.168.0.4 \
    --node 192.168.0.5 \
    --user root \
    --password your-server-password \
    --version v1.18.5 \
    --pkg-url /root/ucc-kube1.18.5-amd64.tar.gz
```



### 安装HA集群

```shell
kubeprince init --master 192.168.0.2 \
    --master 192.168.0.3 \
    --master 192.168.0.4 \
    --node 192.168.0.5 \
    --user root \
    --password your-server-password \
    --version v1.18.5 \
    --pkg-url /root/ucc-kube1.18.5-amd64.tar.gz

```
参数含义：

```
--master   master服务器地址列表
--node     node服务器地址列表
--user     服务器ssh用户名
--password   服务器ssh用户密码
--pkg-url  离线包位置，可以放在本地目录，也可以放在一个http服务器上
--version  kubernetes版本
--pk       ssh私钥地址，配置免密钥默认就是/root/.ssh/id_rsa
--vip      virtual ip (default "10.103.97.2") 本地负载时虚拟ip，不推荐修改，集群外不可访问
```
tar包的目录结构对应kube文件夹，如下:
```
├── bin
│   ├── conntrack
│   ├── crictl
│   ├── kubeadm
│   ├── kubectl
│   ├── kubelet
│   └── kubelet-pre-start.sh
├── conf
│   ├── 10-kubeadm.conf.docker
│   ├── 10-kubeadm.conf.isulad
│   ├── calico.yaml
│   ├── daemon.json
│   ├── kubeadm.yaml
│   └── kubelet.service
├── docker
│   └── docker-deb-amd64.tar.gz
├── images
│   ├── calico-cni.tar
│   ├── calico-kube-controllers.tar
│   ├── calico-node.tar
│   ├── calico-pod2daemon-flexvol.tar
│   ├── coredns.tar
│   ├── etcd.tar
│   ├── kube-apiserver.tar
│   ├── kube-controller-manager.tar
│   ├── kube-proxy.tar
│   ├── kube-scheduler.tar
│   ├── lvsucc.tar
│   └── pause.tar
├── isulad
│   ├── cni-plugins-linux-amd64.tar.gz
│   └── isulad-deb-amd64.tar.gz
├── Metadata
├── README.md
└── shell
    ├── docker.sh
    ├── init-docker.sh
    ├── init-isulad.sh
    ├── init-kube-docker.sh
    ├── init-kube-isulad.sh
    ├── isulad.sh
    ├── killport.sh
    ├── master.sh
    └── update.sh

```

#### 添加node节点

```shell
kubeprince join
    --master 192.168.0.2 \
    --master 192.168.0.3 \
    --master 192.168.0.4 \
    --vip 10.103.97.2 \
    --node 192.168.0.5 \
    --user root \
    --password your-server-password \
    --pkg-url /root/ucc-kube1.18.5-amd64.tar.gz
```

### 清理

```shell
kubeprince clean --all
```


#### 后续更新

1.完善命令行功能，做到极简快速。

2.添加配置文件部署选项。

3.多种cni插件选择


