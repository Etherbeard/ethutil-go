package ethutil

// Config struct isn't exposed
type config struct {
  Db Database
}

var Config *config

func SetConfig(db Database) *config {
  if Config == nil {
    Config = &config{
      Db: db,
    }
  }

  return Config
}
