package database

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type CassandraConfig struct {
	Hosts    []string
	Keyspace string
	Username string
	Password string
}

func NewCassandra(cfg *viper.Viper, log *zap.Logger) (*gocql.Session, error) {
	config := CassandraConfig{
		Hosts:    strings.Split(cfg.GetString("CASSANDRA_HOSTS"), ","),
		Keyspace: cfg.GetString("CASSANDRA_KEYSPACE"),
		Username: cfg.GetString("CASSANDRA_USERNAME"),
		Password: cfg.GetString("CASSANDRA_PASSWORD"),
	}

	// Create cluster configuration
	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Keyspace = config.Keyspace
	cluster.Consistency = gocql.Quorum

	if config.Username != "" && config.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: config.Username,
			Password: config.Password,
		}
	}

	// Create session
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cassandra: %w", err)
	}

	log.Info("Cassandra connected successfully",
		zap.Strings("hosts", config.Hosts),
		zap.String("keyspace", config.Keyspace),
	)

	return session, nil
}