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
	MySQLConnPoolAddress  string

	RedisAddress  string
	RedisPoolSize int
}
