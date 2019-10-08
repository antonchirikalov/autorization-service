package access

import (
	"errors"

	"bitbucket.org/teachingstrategies/go-svc-bootstrap/authorization"
)

// Service represents access service functionality
type Service interface {
	// fetch access data and convert it to *authorization.Access object
	Access(int) (*authorization.Access, error)
}

// holds objects required to manage flow
type accessService struct {
	conv Converter // dbobject to rest object coverter
	repo Dao       // dao
}

var (
	errNotFound   = errors.New("user not found")
	errNotAllowed = errors.New("user not allowed")
	// a map which contains user type IDs allowed to request permissions
	allowedUserTypeIDs = map[int64]struct{}{1: {}, 3: {}, 4:{}, 5: {}, 7: {}}
)

// Access operates flow
func (serv *accessService) Access(userID int) (*authorization.Access, error) {
	relationalAccess, err := serv.repo.QueryAccessData(userID)
	if err != nil {
		return nil, err
	}
	if len(relationalAccess) == 0 || !relationalAccess[0].userTypeID.Valid {
		return nil, errNotFound
	}
	if _, ok := allowedUserTypeIDs[relationalAccess[0].userTypeID.Int64]; !ok {
		return nil, errNotAllowed
	}
	return serv.conv.Convert(relationalAccess), nil
}
