# socket_for_nats
本项目基于nats实现，用于兼容上一版socket+fastdds的总线方案。因为现阶段信息高铁总线系统已舍弃socket+方案，所以本项目仅做兼容旧版用，新项目请直接调用最新总线接口。
# 环境部署
1. 下载go（根据部署机器替换包）`wget https://go.dev/dl/go1.18.3.linux-arm64.tar.gz`
2. 解压并安装`sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.18.3.linux-amd64.tar.gz`
3. 添加路径，在$HOME/.profile中添加`export PATH=$PATH:/usr/local/go/bin`并`source ~/.profile`
4. 检查，运行`go version`
5. 下载本项目`git clone https://github.com/wutianze/socket_for_nats.git`
6. 进入本项目目录并运行`go run .`
# 参数指定
`--address="127.0.0.1:8000"`，socket的监听地址，默认为":8000"
`--nats="nats://39.101.140.145:4222"`，nats服务器地址，默认为我们部署在阿里云上的地址，不需要修改
`-num=3`，对于总控侧，该参数指定有多少个接入的部门，对于各部门，该参数指定自己的序号。例：总控接入高通量、接入网两个部门，总控侧num=2，高通量指定num=0，接入网指定num=1(具体需要需要提前商量好)。
`--name="server"`，指定自己是总控侧（name="server"）还是各部门（name="client"）
`--debug=true`，默认为false，该选项用于开发者调试，为true时本次运行会作为一个socket client尝试连接address指定的socket sever，会等待来自socket server的消息，并发送几条消息给socket server。
# 运行样例
## 总控
`go run . --name="server", --num=3`
## 某部门
`go run . --name="client", --num=0`
## 开发者debug
`go run . --debug=true --address="127.0.0.1:8000"`
# 连接逻辑解析
## 接入部门
接入部门作为socket client连接本项目，上报给总控的消息用socket发给本项目，本项目再转发给总线。总控下达的控制指令由总线发送到本项目，再由本项目通过socket发给接入部门的socket client。
## 总控
总控为每个接入的部门启动一个socket client，都连接本项目。不同部门上报的消息会由总线发送到本项目，本项目会将消息转发到对应的socket连接上。总控下达控制指令通过特定的socket连接发到本项目，本项目转发给总线，最终到达对应的部门。
## 其他
socket server在收到"exit"消息后会退出。其他时候直接ctrl-c信号退出。
