package dbapp

type DBAppConfig struct {
	//应用端口
	ServerPort            int
	ServerUser            string
	ServerPassword        string
	ServerDBName          string
	MySQLConnPoolMinALive int
	MySQLConnPoolMaxAlive int
	MySQLConnPoolMaxIdle  int
	MySQLAddress          string
	MySQLUser             string
	MySQLPassword         string

	RedisAddress  string
	RedisPoolSize int
}
