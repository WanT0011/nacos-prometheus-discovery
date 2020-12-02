package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"nacos-prometheus-discovery/httputil"
	"nacos-prometheus-discovery/model"
	"os"
	"strconv"
	"strings"
)

func FetchPrometheusConfig(config model.Config) {
	targetFilePath := config.TargetFilePath
	nacosHost := config.NacosHost
	namespaceId := config.NamespaceId
	group := config.Group
	//cluster := config.Cluster
	tenant := config.NamespaceId
	dataId := config.DataId

	log.Println("dataId:", dataId)

	configString := GetConfig(nacosHost, tenant, namespaceId, dataId, group)
	log.Println("configString:", configString)

	wfErr := ioutil.WriteFile(targetFilePath, []byte(configString), os.ModePerm)
	if wfErr != nil {
		log.Println("generate target file failed", wfErr)
	}
}

func GeneratePrometheusTarget(config model.Config) {
	//targetFilePath := config.TargetFilePath
	targetFilePath := os.Getenv("TARGET_FILE_PATH")
	nacosHost := config.NacosHost
	namespaceId := config.NamespaceId
	group := config.Group
	cluster := config.Cluster
	serviceNames := GetServiceNames(nacosHost, namespaceId, group)

	// generate target json
	promJsonTargets := []model.PromTarget{}

	for i, serviceName := range serviceNames {
		log.Println()
		log.Println("service["+strconv.Itoa(i)+"] : ", serviceName)
		instance := GetInstance(nacosHost, serviceName, namespaceId, cluster)
		targets := []string{}
		lables := make(map[string]string)

		hosts := instance.Hosts
		for j, host := range hosts {
			// 用元数据信息中的actuator管理端口替换应用服务端口
			var hostAddress string
			metadata := host.Metadata
			if _, ok := metadata["management.port"]; ok {
				hostAddress = host.Ip + ":" + metadata["management.port"]
			} else {
				hostAddress = host.Ip + ":" + strconv.Itoa(host.Port)
				fmt.Println(host.ServiceName, "服务中元信息没有management.port属性,请检查")
			}
			//hostAddress = host.Ip + ":" + strconv.Itoa(host.Port)
			log.Println("host["+strconv.Itoa(j)+"] :", hostAddress)
			log.Println(">> metadata :")
			targets = append(targets, hostAddress)
			for key, value := range metadata {
				log.Println("["+key+"] = ", value)
				validKey := ReplaceInvalidChar(key)
				lables[validKey] = value
			}
			// 将服务名添加到label中
			lables["application_name"] = ReplaceInvalidChar(host.ServiceName)
			// 将服务应用端口添加到label中
			lables["application_port"] = strconv.Itoa(host.Port)
		}
		pt := model.PromTarget{Labels: &lables, Targets: &targets}
		promJsonTargets = append(promJsonTargets, pt)
	}

	targetJson, jsonErr := json.MarshalIndent(promJsonTargets, "", "  ")
	if jsonErr != nil {
		log.Println("marshal json failed", jsonErr)
	}
	jsonString := string(targetJson)
	log.Println("targetJson:", jsonString)

	wfErr := ioutil.WriteFile(targetFilePath, targetJson, os.ModePerm)
	if wfErr != nil {
		log.Println("generate target file failed", wfErr)
	}
}

func GetInstance(nacosHost string, serviceName string, namespaceId string, cluster string) model.Instance {
	instancesUrl := fmt.Sprintf("%s/v1/ns/instance/list?serviceName=%s&namespaceId=%s&cluster=%s", nacosHost, serviceName, namespaceId, cluster)
	//log.Println("=== instancesUrl:", instancesUrl)

	instanceJson, ierr := httputil.Get(instancesUrl)
	if ierr != nil {
		log.Println("get instanceJson failed", ierr)
	}
	instance := model.Instance{}
	json.Unmarshal([]byte(instanceJson), &instance)
	return instance
}

func ReplaceInvalidChar(key string) string {
	validKey := key
	validKey = strings.ReplaceAll(validKey, ".", "_")
	validKey = strings.ReplaceAll(validKey, "-", "_")
	return validKey
}

func GetServiceNums(nacosHost, namespaceId, group string) string {
	serviceUrl := fmt.Sprintf("%s/v1/ns/service/list?pageNo=1&pageSize=10&namespaceId=%s&groupName=%s", nacosHost, namespaceId, group)
	services, err := httputil.Get(serviceUrl)
	//fmt.Println("services:", services)
	if err != nil {
		log.Println("get service num failed", err)
	}
	service := model.ServiceNum{}
	_ = json.Unmarshal([]byte(services), &service)
	fmt.Println("service:", service)
	return strconv.Itoa(service.Count)
}

func GetServiceNames(nacosHost string, namespaceId string, group string) []string {
	pageSize := GetServiceNums(nacosHost, namespaceId, group)
	fmt.Println("pageSize:", pageSize)
	serviceUrl := fmt.Sprintf("%s/v1/ns/service/list?pageNo=1&pageSize=%s&namespaceId=%s&groupName=%s", nacosHost, pageSize, namespaceId, group)
	log.Println("=== serviceUrl:", serviceUrl)

	services, serr := httputil.Get(serviceUrl)
	if serr != nil {
		log.Println("get service failed", serr)
	}

	service := model.Service{}
	_ = json.Unmarshal([]byte(services), &service)
	return service.Doms
}

func GetConfig(nacosHost, tenant, namespaceId, dataId, group string) string {
	configUrl := fmt.Sprintf("%s/v1/cs/configs?tenant=%s&namespaceId=%s&dataId=%s&group=%s", nacosHost, tenant, namespaceId, dataId, group)
	log.Println("=== configUrl:", configUrl)

	config, serr := httputil.Get(configUrl)
	if serr != nil {
		log.Println("get config failed", serr)
	}
	return config
}
