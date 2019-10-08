package access

import (
	"database/sql"
)

// Dao describes operations which can be done on the CCNET db
type Dao interface {
	// Qeries access data for user with given ID
	QueryAccessData(int) ([]*accessDataRow, error)
}

// DAO object which does logic related to quering db
// keeps dbmanager to get db connection
type accessRepo struct {
	db *sql.DB
}

func (r *accessRepo) QueryAccessData(userID int) ([]*accessDataRow, error) {
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	accessData := make([]*accessDataRow, 0)
	for rows.Next() {
		var row = new(accessDataRow)
		err = rows.Scan(&row.userTypeID,
			&row.adminTypeID,
			&row.fundSourceAdminTypeID,
			&row.superUserTypeID,
			&row.fundSourceID,
			&row.adminEntityID,
			&row.fsAdminEntityID,
			&row.classID,
			&row.teacherTypeID,
			&row.teamChildID)
		if err != nil {
			return nil, err
		}
		accessData = append(accessData, row)
	}
	return accessData, nil
}

type accessDataRow struct {
	userTypeID            sql.NullInt64
	adminTypeID           sql.NullInt64
	fundSourceAdminTypeID sql.NullInt64
	superUserTypeID       sql.NullInt64
	fundSourceID          sql.NullInt64
	adminEntityID         sql.NullInt64
	fsAdminEntityID       sql.NullInt64
	classID               sql.NullInt64
	teacherTypeID         sql.NullInt64
	teamChildID           sql.NullInt64
}

const query = `
	SELECT u.UserTypeID,
		u.AdminTypeID,
		u.FundSourceAdminTypeID,
		u.SuperUserTypeID,
		fsa.FundSourceID,
		COALESCE(es.EntityID, ep.EntityID, eo.EntityID) AS AdminEntityID,
		COALESCE(efs.EntityID, efp.EntityID, efo.EntityID) AS FSAdminEntityID,
		ct.ClassID,
		ct.TeacherTypeID,
		tci.ChildID AS TeamChildID
	FROM dbo.CC_Users u WITH (NOLOCK)
	 LEFT JOIN dbo.CC_UserAssoc ua WITH (NOLOCK) ON ua.UserID = u.UserID
	 LEFT JOIN dbo.CC_AdminFundSources fsa WITH (NOLOCK) ON fsa.UserID = u.UserID
	 LEFT JOIN dbo.CC_FSUserAssoc fsua WITH (NOLOCK) ON fsua.UserID = u.UserID
	 LEFT JOIN dbo.CC_TC_Invitations tci WITH (NOLOCK) ON tci.UserID = u.UserID AND tci.InvitationStatusID = 4
	 LEFT JOIN dbo.G2_EntityLink es WITH (NOLOCK) ON ua.SiteID = es.SiteID
	 LEFT JOIN dbo.G2_EntityLink ep WITH (NOLOCK) ON ua.ProgramID = ep.ProgramID
	 LEFT JOIN dbo.G2_EntityLink eo WITH (NOLOCK) ON ua.OrganizationID = eo.OrganizationID
	 LEFT JOIN dbo.G2_EntityLink efs WITH (NOLOCK) ON fsua.SiteID = efs.SiteID
	 LEFT JOIN dbo.G2_EntityLink efp WITH (NOLOCK) ON fsua.ProgramID = efp.ProgramID
	 LEFT JOIN dbo.G2_EntityLink efo WITH (NOLOCK) ON fsua.OrganizationID = efo.OrganizationID
	 LEFT JOIN dbo.CC_ClassesTeachers ct WITH (NOLOCK) ON ct.TeacherID = u.UserID
WHERE u.UserID = ?`
