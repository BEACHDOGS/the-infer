package cores

import (
	"infer-microservices/cores/service_config"
)

type ServiceConfigDirector struct {
}

func (s *ServiceConfigDirector) ServiceConfigUpdateContainIndexDirector(domain string, dataId string,
	redisConfStr string, modelConfStr string, indexConfStr string) service_config.ServiceConfig {
	//load redis,faiss,model
	serviceConfigBuilder := service_config.ServiceConfigBuilder{}
	builder := serviceConfigBuilder.RedisClientBuilder(domain, dataId, redisConfStr).FaissClientBuilder(indexConfStr).ModelClientBuilder(modelConfStr)

	return builder.GetServiceConfig()
}

func (s *ServiceConfigDirector) ServiceConfigUpdaterNotContainIndexDirector(domain string, dataId string,
	redisConfStr string, modelConfStr string) service_config.ServiceConfig {
	//load redis,model
	serviceConfigBuilder := service_config.ServiceConfigBuilder{}
	builder := serviceConfigBuilder.RedisClientBuilder(domain, dataId, redisConfStr).ModelClientBuilder(modelConfStr)

	return builder.GetServiceConfig()
}
