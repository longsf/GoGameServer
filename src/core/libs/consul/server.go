package consul

import (
	. "core/libs"
	"core/libs/common"
	"github.com/hashicorp/consul/api"
	"os"
	"os/signal"
	"strconv"
)

func InitServer(serviceName string, serviceId int, servicePort string) error {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}

	//服务器配置
	address := common.GetLocalIp()
	port, _ := strconv.Atoi(servicePort)
	id := address + ":" + servicePort + "_" + serviceName + ":" + NumToString(serviceId)
	name := serviceName

	//健康检查配置
	checkPath := address + ":" + servicePort

	//服务注册
	service := &api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: address,
		Port:    port,
		Tags:    []string{name},
		Check: &api.AgentServiceCheck{
			TCP:                            checkPath,
			Timeout:                        "1s",
			Interval:                       "3s",
			DeregisterCriticalServiceAfter: "10s", //check失败后10秒删除本服务
		},
	}
	err = client.Agent().ServiceRegister(service)
	if err != nil {
		return err
	}

	//关闭处理
	go WaitToUnRegistService(client, id)

	return nil
}

func WaitToUnRegistService(client *api.Client, serviceId string) {
	//监听系统退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	//取消监听
	signal.Stop(quit)
	close(quit)

	//从服务中移除
	err := client.Agent().ServiceDeregister(serviceId)
	if err != nil {
		ERR(err)
	}
}
