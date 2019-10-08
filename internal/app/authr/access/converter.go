package access

import "bitbucket.org/teachingstrategies/go-svc-bootstrap/authorization"

// Converter represent methods for converting db rows to authorization.Access
type Converter interface {
	// converting method
	Convert([]*accessDataRow) *authorization.Access
}

// converts db struct into rest result struct
type accessConverter struct{}

// Convert does the conversion
func (conv *accessConverter) Convert(rows []*accessDataRow) *authorization.Access {
	res := new(authorization.Access)
	if len(rows) > 0 && rows[0].userTypeID.Valid {

		userTypeID := rows[0].userTypeID.Int64

		conv.convertSuperUser(rows, res)

		conv.convertAdmin(rows, res, userTypeID)

		conv.convertFSAdmin(rows, res, userTypeID)

		conv.convertTeacher(rows, res, userTypeID)

		conv.convertTeamMember(rows, res, userTypeID)
	}
	return res
}

func (*accessConverter) convertSuperUser(rows []*accessDataRow, res *authorization.Access) {
	res.SuperUser = rows[0].superUserTypeID.Int64 != int64(0)
}

func (*accessConverter) convertTeacher(rows []*accessDataRow, res *authorization.Access, userTypeID int64) {
	teachers := make(map[int64]map[int64]interface{})
	for _, row := range rows {
		teacherTypeID, class := row.teacherTypeID, row.classID
		if teacherTypeID.Valid && class.Valid {
			if foundTeacher, ok := teachers[teacherTypeID.Int64]; ok {
				foundTeacher[class.Int64] = struct{}{}
			} else {
				teachers[teacherTypeID.Int64] = map[int64]interface{}{class.Int64: struct{}{}}
			}
		}
	}

	if foundTeacher, ok := teachers[1]; (ok && len(foundTeacher) > 0) || userTypeID == 1 { // teacher
		res.Teacher = &authorization.TeacherType{Cls: keys(foundTeacher)}
	}
	if foundCoTeacher, ok := teachers[2]; ok && len(foundCoTeacher) > 0 && userTypeID == 1 { // co teacher
		res.CoTeacher = &authorization.TeacherType{Cls: keys(foundCoTeacher)}
	}
	if foundAssistant, ok := teachers[3]; ok && len(foundAssistant) > 0 && userTypeID == 1 { // assistant
		res.AssistantTeacher = &authorization.TeacherType{Cls: keys(foundAssistant)}
	}
}

func (*accessConverter) convertTeamMember(rows []*accessDataRow, res *authorization.Access, userTypeID int64) {
	teamMembers := make(map[int64]interface{})
	for _, row := range rows {
		if teamChildID := row.teamChildID; teamChildID.Valid {
			teamMembers[teamChildID.Int64] = struct{}{}
		}
	}
	if userTypeID == 5 || len(teamMembers) > 0 {
		res.TeamMember = &authorization.TeamMemberType{Kid: keys(teamMembers)}
	}
}

func (*accessConverter) convertFSAdmin(rows []*accessDataRow, res *authorization.Access, userTypeID int64) {
	if userTypeID != 7 {
		return
	}
	fsAdminItems := make(map[int64]interface{})
	fsItems := make(map[int64]interface{})
	for _, row := range rows {
		if fsAdminEntityID := row.fsAdminEntityID; fsAdminEntityID.Valid {
			fsAdminItems[fsAdminEntityID.Int64] = struct{}{}
		}
		if fsAdminEntityID := row.fundSourceID; fsAdminEntityID.Valid {
			fsItems[fsAdminEntityID.Int64] = struct{}{}
		}
	}
	fundSourceAdminTypeID := rows[0].fundSourceAdminTypeID
	if len(fsItems) > 0 && fundSourceAdminTypeID.Valid {
		admin := &authorization.FsAdminType{
			AdminType: authorization.AdminType{Ent: keys(fsAdminItems)},
			FundSrc:   keys(fsItems)}
		switch fundSourceAdminTypeID.Int64 {
		case 0:
			res.FSAdmin = admin
		case 1:
			res.FSVOAdmin = admin
		}
	}
}

func (*accessConverter) convertAdmin(rows []*accessDataRow, res *authorization.Access, userTypeID int64) {
	if userTypeID != 3 {
		return
	}
	entityIDs := make(map[int64]interface{})
	for _, row := range rows {
		if adminEntityID := row.adminEntityID; adminEntityID.Valid {
			entityIDs[adminEntityID.Int64] = struct{}{}
		}
	}
	adminTypeID := rows[0].adminTypeID
	if len(entityIDs) > 0 && adminTypeID.Valid {
		admin := &authorization.AdminType{Ent: keys(entityIDs)}
		switch adminTypeID.Int64 {
		case 0:
			res.Admin = admin
		case 1:
			res.VOAdmin = admin
		case 2:
			res.VONoChildAdmin = admin
		}
	}
}

// returns slice of map's keys of map[int64]interface{}
func keys(m map[int64]interface{}) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
