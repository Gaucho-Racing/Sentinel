package database

import (
	"context"
	"log"
	"sentinel/config"
	"sentinel/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestInitializeDB(t *testing.T) {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	utils.SugarLogger = logger.Sugar()

	// Start MySQL container
	ctx := context.Background()

	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithDatabase(config.DatabaseName),
		mysql.WithUsername(config.DatabaseUser),
		mysql.WithPassword(config.DatabasePassword),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	port, err := mysqlContainer.MappedPort(ctx, "3306/tcp")
	if err != nil {
		log.Fatalf("failed to get container port: %s", err)
	}

	host, err := mysqlContainer.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get container host: %s", err)
	}

	config.DatabaseHost = host
	config.DatabasePort = port.Port()

	// Clean up the container
	defer func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// Test successful connection
	t.Run("Successful Connection", func(t *testing.T) {
		InitializeDB()
		assert.NotNil(t, DB, "DB should not be nil after successful connection")
		assert.IsType(t, &gorm.DB{}, DB, "DB should be of type *gorm.DB")
	})

	// Test retry mechanism
	t.Run("Retry Mechanism", func(t *testing.T) {
		config.DatabaseHost = "non-existent-host"

		done := make(chan bool)
		go func() {
			InitializeDB()
			done <- true
		}()

		select {
		case <-done:
			t.Error("InitializeDB should not have succeeded with non-existent host")
		case <-time.After(10 * time.Second):
			// Test passes if it times out after 10 seconds
		}
	})
}
