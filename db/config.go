package db

type DatabaseConf struct {
	// Ip 数据库连接地址
	Ip string `ini:"ip" toml:"ip"`

	// Port 数据库连接端口
	Port int `ini:"port" toml:"port"`

	// Password 数据库用户密码
	Password string `ini:"password" toml:"password"`

	// DB 数据库名称
	DB string `ini:"db" toml:"db"`

	// User 数据库用户名
	User string `ini:"user" toml:"user"`

	// Timeout 数据库连接超时时间
	Timeout int `ini:"timeout" toml:"timeout"`

	// MaxConnection 设置打开数据库连接的最大数量
	MaxConnection int `ini:"max_connection" toml:"max_connection"`

	// MaxIdleConnection 设置空闲连接池中连接的最大数量
	MaxIdleConnection int `ini:"max_idle_connection" toml:"max_idle_connection"`

	// MaxLifetime 设置了连接可复用的最大时间
	MaxLifetime int `ini:"max_life_time" toml:"max_lifetime"`
}
