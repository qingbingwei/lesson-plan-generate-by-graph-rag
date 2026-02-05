package database

import (
	"context"
	"fmt"

	"lesson-plan/backend/internal/config"
	"lesson-plan/backend/pkg/logger"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var neo4jDriver neo4j.DriverWithContext

// InitNeo4j 初始化Neo4j连接
func InitNeo4j(cfg *config.Neo4jConfig) (neo4j.DriverWithContext, error) {
	var err error
	neo4jDriver, err = neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.User, cfg.Password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}

	ctx := context.Background()
	if err := neo4jDriver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify neo4j connectivity: %w", err)
	}

	logger.Info("Neo4j connected successfully",
		logger.String("uri", cfg.URI),
	)

	return neo4jDriver, nil
}

// GetNeo4jDriver 获取Neo4j驱动
func GetNeo4jDriver() neo4j.DriverWithContext {
	return neo4jDriver
}

// CloseNeo4j 关闭Neo4j连接
func CloseNeo4j(ctx context.Context) error {
	if neo4jDriver != nil {
		return neo4jDriver.Close(ctx)
	}
	return nil
}

// Neo4jSession 创建Neo4j会话
func Neo4jSession(ctx context.Context, database string) neo4j.SessionWithContext {
	return neo4jDriver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
	})
}

// ExecuteQuery 执行Neo4j查询
func ExecuteQuery(ctx context.Context, database, query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	session := neo4jDriver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var results []map[string]interface{}
		for records.Next(ctx) {
			record := records.Record()
			row := make(map[string]interface{})
			for _, key := range record.Keys {
				val, _ := record.Get(key)
				row[key] = val
			}
			results = append(results, row)
		}
		return results, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]map[string]interface{}), nil
}

// ExecuteWrite 执行Neo4j写入
func ExecuteWrite(ctx context.Context, database, query string, params map[string]interface{}) error {
	session := neo4jDriver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	return err
}
