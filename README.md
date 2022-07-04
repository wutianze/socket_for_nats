# socket_for_nats
本项目基于nats实现，用于兼容上一版socket+fastdds的总线方案。因为现阶段信息高铁总线系统已舍弃socket+方案，所以本项目仅做兼容旧版用，新项目请直接调用最新总线接口。
# 环境部署
1. 下载go（根据部署机器替换包）`wget https://go.dev/dl/go1.18.3.linux-arm64.tar.gz`，如果是x86机器，换成go1.18.3.linux-amd64.tar.gz
2. 解压并安装`sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.18.3.linux-arm64.tar.gz`
3. 添加路径，在$HOME/.profile中添加`export PATH=$PATH:/usr/local/go/bin`并`source ~/.profile`
4. 检查，运行`go version`
5. 下载本项目`git clone https://github.com/wutianze/socket_for_nats.git`
6. 进入本项目目录并运行`go run .` （如果国内网络限制下载失败，可以使用go代理，`go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn`）
## 容器
除了上面的直接部署，也可选择容器部署，amd64镜像：sauronwu/socket_nats_amd64。容器运行示例`sudo docker run -p 8000:8000 -dw /home/socket_for_nats/ sauronwu/socket_nats_amd64:v0.1 git pull && go run . --name="server" --num=1`（一般情况下可以不用git pull）
# 参数指定
`--address="127.0.0.1:8000"`，socket的监听地址，默认为":8000"  
`--nats="nats://39.101.140.145:4222"`，nats服务器地址，默认为我们部署在阿里云上的地址，如果只能访问南京内网，则使用192.168.103.4:4222。  
`-num=3`，对于综合管控侧，该参数指定有多少个接入的部门，对于各部门（接入网、智能网、高通量），该参数指定自己的序号。例：总控接入高通量、接入网两个部门，总控侧num=2，高通量指定num=0，接入网指定num=1(具体需要需要提前商量好)。  
`--name="server"`，指定自己是综合管控侧（name="server"）还是各部门（name="client"）  
`--debug=true`，默认为false，该选项用于开发者调试，为true时本次运行会作为一个socket client尝试连接address指定的socket sever，会等待来自socket server的消息，并发送几条消息给socket server。  
# 运行样例
## 综合管控
`go run . --name="server" --num=3`
## 接入网、智能网、高通量
`go run . --name="client" --num=0`
## 开发者debug
`go run . --debug=true --address="127.0.0.1:8000"`
# 连接逻辑解析
## 接入网、智能网、高通量
接入部门作为socket client连接本项目，上报给总控的消息用socket发给本项目，本项目再转发给总线。总控下达的控制指令由总线发送到本项目，再由本项目通过socket发给接入部门的socket client。
## 综合管控
综合管控为每个接入的部门启动一个socket client，都连接本项目。不同部门上报的消息会由总线发送到本项目，本项目会将消息转发到对应的socket连接上。总控下达控制指令通过特定的socket连接发到本项目，本项目转发给总线，最终到达对应的部门。
## 其他
socket server在收到"exit"消息后会退出。其他时候直接ctrl-c信号退出。
# 直接使用总线接口
见总线接口说明文档，安装：`go get github.com/wutianze/nats.go@main`。直接使用总线接口暂不支持使用request-respond接口和其他使用socket的用户进行交互。
