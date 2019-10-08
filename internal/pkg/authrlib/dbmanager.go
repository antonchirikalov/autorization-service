package authrlib

import (
	"database/sql"
	"fmt"
)

// DbManager manages a connection with database
type DbManager interface {
	// Provides a valid *sql.DB object
	Db() *sql.DB
	// Closes connection with db
	Release()
}

// impl of DbManager
type dbManager struct {
	ctx  *AppContext
	conn *sql.DB
}

// NewDbManager instantiates new mssql database manager
func NewDbManager(ctx *AppContext) (DbManager, error) {
	connConfig := ctx.ConfigService.Config().MsSQL.Connection

	mssconn, err := sql.Open("mssql",
		fmt.Sprintf("server=%s; port=%v; database=%s; user id=%s; password=%s;",
			connConfig.Host,
			connConfig.Port,
			connConfig.Database,
			connConfig.User,
			connConfig.Password))
	if err != nil {
		return nil, err
	}

	manager := &dbManager{ctx, mssconn}

	ctx.Healthchecks.AddHealthCheck("CCNet", func() (bool, error) {
		err := manager.conn.Ping()
		return err == nil, err
	})

	return manager, nil
}

func (m *dbManager) Release() {
	if err := m.conn.Close(); err != nil {
		m.ctx.Logger.Error().Err(err).Msg("unable to release db connection")
	}
}

func (m *dbManager) Db() *sql.DB {
	return m.conn
}
