## 特性：一键安装集群，默认支持iSulad

## 安装前提条件

- 一台或多台运行着下列系统的机器:
  - Ubuntu 16.04+
  - Debian 9
  - CentOS 7
  - RHEL 7
  - Fedora 25/26
  - HypriotOS v1.0.1+
  - Container Linux (针对1800.6.0 版本测试)
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

- 支持离线安装，工具与资源包（二进制程序 配置文件 镜像 yaml文件等）分离,这样不同版本替换不同离线包即可

- 证书延期

- 使用简单

- 支持自定义配置

- 内核负载均衡



## 为什么不使用ansilbe

二进制文件工具，没有任何依赖，文件分发与远程命令都通过调用sdk实现所以不依赖其它任何东西



## 定制kubeadm

kubeadm把证书时间写死了，所以需要定制把它改成100年

做本地负载时修改kubeadm代码是最方便的，因为在join时我们需要做两个事，第一join之前先创建好ipvs规则，第二创建static pod，如果这块不去定制kubeadm就把报静态pod目录已存在的错误，忽略这个错误很不优雅。 而且kubeadm中已经提供了一些很好用的sdk供我们去实现这个功能。



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
    --pkg-url /root/kube1.18.5.tar.gz
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
    --pkg-url /root/kube1.1.18.5.tar.gz
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
    --pkg-url /root/kube1.18.5.tar.gz
```

### 清理

```shell
kubeprince clean \
    --master 192.168.0.2 \
    --master 192.168.0.3 \
    --master 192.168.0.4 \
    --node 192.168.0.5 \
    --user root \
    --password your-server-password
```



#### 后续更新

1.完善命令行功能，做到极简快速。

2.增加云原生应用部署

3.完善日志功能

4.添加配置文件部署选项。


