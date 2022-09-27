package structs

type Options struct {
	RedisDB           int  `required:"true"`
	RedisIsCluster    bool `required:"true"`
	RedisHost         string
	RedisPort         string
	RedisPass         string
	RedisClusterHosts []string
}

var OptionsLoaded = &Options{}
