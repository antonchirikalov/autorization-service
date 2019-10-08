package access

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestQueryAccessData(t *testing.T) {
	var columns = []string{"UserTypeID", "AdminTypeID", "FundSourceAdminTypeID",
		"SuperUserTypeID", "FundSourceID", "AdminEntityID", "FSAdminEntityID",
		"ClassID", "TeacherTypeID", "TeamChildID"}

	sqlQuery := regexp.QuoteMeta(query)
	var repo Dao
	db, mock, err := sqlmock.New()
	defer db.Close()

	if err != nil {
		t.Errorf("unable to create db mock: %v", err)
	}
	repo = &accessRepo{db}

	testCases := []struct {
		name        string
		err         string
		prepareMock func(string)
	}{
		{
			name: "sql prepare query error",
			err:  "prepare error", prepareMock: func(err string) {
				mock.ExpectPrepare(sqlQuery).WillReturnError(errors.New(err))
			}}, // prepare statement error
		{
			name: "sql query error",
			err:  "query error",
			prepareMock: func(err string) {
				mock.ExpectPrepare(sqlQuery).WillBeClosed().ExpectQuery().WithArgs(333).WillReturnError(errors.New(err))
			}}, // query error
		{
			name: "sql rows scan error",
			err:  "sql: Scan error on column index 9, name \"TeamChildID\": converting driver.Value type string (\"\") to a int64: invalid syntax",
			prepareMock: func(string) {
				rows := sqlmock.NewRows(columns).AddRow(10, 9, 8, 7, 6, 5, 4, 3, 2, "")
				mock.ExpectPrepare(sqlQuery).WillBeClosed().ExpectQuery().WithArgs(333).WillReturnRows(rows)
			}}, // scan error
		{
			name: "success case",
			err:  "",
			prepareMock: func(string) {
				rows := sqlmock.NewRows(columns).AddRow(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).AddRow(10, 9, 8, 7, 6, 5, 4, 3, 2, 1)
				mock.ExpectPrepare(sqlQuery).WillBeClosed().ExpectQuery().WithArgs(333).WillReturnRows(rows)
			}}, // success

	}

	for _, testCase := range testCases {
		testErr := errors.New(testCase.err)
		testCase.prepareMock(testCase.err)
		if data, err := repo.QueryAccessData(333); testCase.err != "" && (err == nil || err.Error() != testCase.err) {
			t.Fatalf("test case failed: '%s' [expect %v; got %v]", testCase.name, testErr, err)
		} else if testCase.err == "" {
			if result := assert.Equal(t, len(data), 2); !result {
				t.Logf("test case failed: %s", testCase.name)
				continue
			}
		}

		t.Log("test case ok:", testCase.name)
	}
}
