package types

type (
	// 游戏服 登录配置
	LoginConfig struct {
		Server        string      `json:"server,omitempty"` // 服务器地址
		Region        int32       `json:"region,omitempty"` // 区服
		Device        interface{} `json:"device_id,omitempty"` // 设备号
		Account       interface{} `json:"account,omitempty"` // 压测时忽略
		ClientVersion string      `json:"client_version,omitempty"` // 客户端版本号
	}

	// 命令行参数
	CmdArgs struct {
		Tree     string `arg:",01,压测用例"`
		Start    int32  `arg:"start,1,起始序号"`
		Count    int32  `arg:"count,1,压测数量"`
		Timeout  int32  `arg:"timeout,0,压测时间(秒)"`
		Region   int32  `arg:"region,1,区服"`
		Server   string `arg:"server,127.0.0.1:10002,服务器地址"`
		StopWait int32  `arg:"wait,0,压测停止后等待时间(秒)"`
	}

	AppConfig struct {
		LogLevel string `json:"log_level,omitempty"`
		LogPath  string `json:"log_path,omitempty"`
		TickMs   int64  `json:"tick_ms,omitempty"` // 行为树每x毫秒跑一次
	}
)

var (
	Args    CmdArgs
	AppConf *AppConfig
)
