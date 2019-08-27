package main

import (
	"fmt"
	"io/ioutil"
	"os"

	redis "github.com/chremoas/services-common/redis"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

func setupRedis(addr, password, namespace, description string, serverAdmins []string) {
	service := fmt.Sprintf("%s.%s.%s", namespace, "srv", "perms")
	redisClient := redis.Init(addr, password, 0, service)
	permName := redisClient.KeyName("members:server_admins")
	permDesc := redisClient.KeyName("description:server_admins")

	_, err := redisClient.Client.Ping().Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Setting Server Admins description: %s\n", description)

	_, err = redisClient.Client.Set(permDesc, description, 0).Result()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Removing all server admins")
	redisClient.Client.Del(permName)
	for i := range serverAdmins {
		fmt.Printf("Adding: %s\n", serverAdmins[i])
		_, err = redisClient.Client.SAdd(permName, serverAdmins[i]).Result()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func setupConsul(addr, configNamespace, configFile string, consulCredentials map[string]string) {
	dat, err := ioutil.ReadFile(configFile)
	config := consulApi.DefaultConfig()
	config.Address = addr
	if consulCredentials != nil {
		config.HttpAuth = &consulApi.HttpBasicAuth{Username: consulCredentials["username"], Password: consulCredentials["password"]}
	}
	consul, err := consulApi.NewClient(config)
	if err != nil {
		fmt.Println(err)
	}

	kv := consul.KV()
	configData := &consulApi.KVPair{
		Key:   fmt.Sprintf("%s/config", configNamespace),
		Value: dat,
	}
	_, err = kv.Put(configData, nil)
	if err != nil {
		fmt.Println(err)
	}

	configType := &consulApi.KVPair{
		Key:   fmt.Sprintf("%s/configType", configNamespace),
		Value: []byte("yaml"),
	}
	_, err = kv.Put(configType, nil)
	if err != nil {
		fmt.Println(err)
	}

}

func main() {
	viper.SetConfigName("chremoas")
	viper.AddConfigPath(".")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.description", "Server Admins")
	viper.SetDefault("redis.namespace", "com.aba-eve")

	viper.SetDefault("consul.host", "localhost")
	viper.SetDefault("consul.port", 8500)
	viper.SetDefault("consul.config.file", "chremoas-config.yaml")
	viper.SetDefault("consul.config.namespace", "chremoas-default")
	viper.SetDefault("consul.credentials.username", "")
	viper.SetDefault("consul.credentials.password", "")

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// admin list is required
	if !viper.IsSet("redis.admins") {
		fmt.Println("Admin list is required!")
		os.Exit(1)
	}

	setupRedis(
		fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		viper.GetString("redis.password"),
		viper.GetString("redis.namespace"),
		viper.GetString("redis.description"),
		viper.GetStringSlice("redis.admins"),
	)

	consulCredentials := map[string]string{
		"username": viper.GetString("consul.credentials.username"),
		"password": viper.GetString("consul.credentials.password"),
	}

	if (consulCredentials["username"] == "") || (consulCredentials["password"] == "") {
		consulCredentials = nil
	}

	setupConsul(
		fmt.Sprintf("%s:%d", viper.GetString("consul.host"), viper.GetInt("consul.port")),
		viper.GetString("consul.config.namespace"),
		viper.GetString("consul.config.file"),
		consulCredentials,
	)
}
