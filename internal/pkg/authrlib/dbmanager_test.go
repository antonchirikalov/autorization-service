package authrlib

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"bitbucket.org/teachingstrategies/go-svc-bootstrap/health"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type fakeMSSQLDriver struct{ mock.Mock }

func (d *fakeMSSQLDriver) Open(name string) (driver.Conn, error) { return nil, nil }

func TestNewDbManager(t *testing.T) {
	// error case
	healthcheck := health.NewHealthCheckCollection()
	ctx := &AppContext{ConfigService: &configService{config: &Config{
		MsSQL: MsSQLConfig{
			Connection: Connection{},
		},
	}}, Healthchecks: healthcheck}
	_, err := NewDbManager(ctx)
	assert.Error(t, err)

	sql.Register("mssql", &fakeMSSQLDriver{})
	_, err = NewDbManager(ctx)
	assert.NoError(t, err)

	b, err := healthcheck.IsHealthy()

	assert.True(t, b)
	assert.NoError(t, err)
}

func TestDB(t *testing.T) {
	mockDB := &sql.DB{}
	dbmanager := &dbManager{conn: mockDB}
	assert.Equal(t, mockDB, dbmanager.Db())
}

func TestRelease(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	mock.ExpectClose()
	dbmanager := &dbManager{conn: db}
	dbmanager.Release()

	// fail close
	db, mock, err = sqlmock.New()
	assert.NoError(t, err)
	var buf bytes.Buffer
	logger := &AppLogger{Logger: zerolog.New(&buf)}
	dbmanager = &dbManager{conn: db, ctx: &AppContext{Logger: logger}}
	mock.ExpectClose().WillReturnError(fmt.Errorf("fake close db"))
	dbmanager.Release()
	assert.Equal(t, `{"level":"error","error":"fake close db","message":"unable to release db connection"}`, strings.TrimSpace(buf.String()))
}
