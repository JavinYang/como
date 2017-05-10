package como

// 配置实例
var Config config

// 初始化配置实例
func init() {
	Config = config{make(map[string]string)}
	Config.init()
}

type config struct {
	config map[string]string
}

func (this *config) init() {}

func (this *config) Get(key string) (val string) {
	return ""
}
