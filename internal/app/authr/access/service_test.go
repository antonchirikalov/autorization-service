package access

import (
	"database/sql"
	"errors"
	"testing"

	"bitbucket.org/teachingstrategies/go-svc-bootstrap/authorization"
	"github.com/stretchr/testify/mock"
)

type accessServiceDepsMock struct{ mock.Mock }

func (m *accessServiceDepsMock) Convert(rows []*accessDataRow) *authorization.Access { //mock converter's method
	args := m.Called(rows)
	return args.Get(0).(*authorization.Access)
}

func (m *accessServiceDepsMock) QueryAccessData(userID int) ([]*accessDataRow, error) { // mock dao method
	args := m.Called(userID)
	data := args.Get(0)
	if data == nil {
		return nil, args.Error(1)
	}
	return data.([]*accessDataRow), args.Error(1)
}

func TestAccess(t *testing.T) {
	testCases := []struct {
		name      string
		err       string
		rows      []*accessDataRow
		accessErr error
	}{
		{
			name:      "db error",
			err:       "my query db error",
			rows:      nil,
			accessErr: errors.New("my query db error"),
		},
		{
			name:      "not found (empty db result)",
			err:       errNotFound.Error(),
			rows:      nil,
			accessErr: nil,
		},
		{
			name:      "not found (invalid user type id)",
			err:       errNotFound.Error(),
			rows:      []*accessDataRow{{userTypeID: sql.NullInt64{Int64: 0, Valid: false}}},
			accessErr: nil,
		},
		{
			name:      "not allowed",
			err:       errNotAllowed.Error(),
			rows:      []*accessDataRow{{userTypeID: sql.NullInt64{Int64: 6, Valid: true}}},
			accessErr: nil,
		},
		{
			name:      "success",
			err:       "",
			rows:      []*accessDataRow{{userTypeID: sql.NullInt64{Int64: 1, Valid: true}}},
			accessErr: nil,
		},
	}
	for _, testCase := range testCases {
		mock := &accessServiceDepsMock{}
		service := &accessService{conv: mock, repo: mock}
		testErr := errors.New(testCase.err)
		convReply := &authorization.Access{}

		if testCase.err == "" {
			mock.On("Convert", testCase.rows).Return(convReply).Once()
		}

		mock.On("QueryAccessData", 42).Return(testCase.rows, testCase.accessErr).Once()
		if reply, err := service.Access(42); testCase.err != "" && (err == nil || err.Error() != testCase.err) {
			t.Fatalf("test case failed: '%s' [expect '%v'; got '%v']", testCase.name, testErr, err)
			continue
		} else if testCase.err == "" && reply != convReply {
			t.Fatalf("test case failed (reply is not valid): '%s' [expected: '%v'; got '%v']", testCase.name, convReply, reply)
			continue
		}
		if result := mock.AssertExpectations(t); !result {
			t.Log("test case failed (mock: assert expectations)", testCase.name)
			continue
		}
		t.Log("test case ok", testCase.name)

	}
}
