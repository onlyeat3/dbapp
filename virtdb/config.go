package virtdb

type VirtdbConfig struct {
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

	RedisAddresses []string
	RedisAddress   string
	RedisPoolSize  int
	RedisPassword  string
}
