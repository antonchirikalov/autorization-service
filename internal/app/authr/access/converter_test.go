package access

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkConvert(b *testing.B) {
	conv := &accessConverter{}
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(testCases); j++ {
			p := testCaseToPayload(testCases[j].rows)
			_ = conv.Convert(p)
		}
	}
}

func TestConvert(t *testing.T) {
	conv := &accessConverter{}
	for j := 0; j < len(testCases); j++ {
		testCase := testCases[j]
		rows := testCaseToPayload(testCase.rows)
		res := conv.Convert(rows)

		sortNestedInt64Slices(res)

		bytes, err := json.Marshal(res)
		assert.Nil(t, err)
		if result := assert.JSONEq(t, string(bytes), testCase.result); result {
			t.Log("test ok:", testCase.name)
		} else {
			t.Error("test failed:", testCase.name)
		}
	}
}

// used to sort all nested fields which has type []int64, because `assert.JSONEq`
// consider that serialized json `[1,2]` is not equal `[2,1]`.
// test probes also should be defined as sorted array.
func sortNestedInt64Slices(x interface{}) {
	v := reflect.ValueOf(x)
	if reflect.ValueOf(x).Kind() == reflect.Ptr { // if pointer -> extact
		v = reflect.ValueOf(x).Elem()
	}
	if v.Kind() < 10 { // skip `primitives`, see reflect.Kind
		return
	}
	// check if []int64
	if reflect.TypeOf(v.Interface()).AssignableTo(reflect.TypeOf([]int64{})) {
		slice := v.Interface().([]int64) // sort
		sort.Slice(slice, func(i, j int) bool { return slice[i] < slice[j] })
		return
	}
	// try to sort fields
	for i := 0; i < v.NumField(); i++ {
		sortNestedInt64Slices(v.Field(i).Interface())
	}
}

func testCaseToPayload(testCaseAccessData []accessDataRowTest) []*accessDataRow {
	result := make([]*accessDataRow, len(testCaseAccessData))
	for i := 0; i < len(testCaseAccessData); i++ {
		result[i] = &accessDataRow{
			convertInt64toNullInt64(testCaseAccessData[i].userTypeID),
			convertInt64toNullInt64(testCaseAccessData[i].adminTypeID),
			convertInt64toNullInt64(testCaseAccessData[i].fundSourceAdminTypeID),
			convertInt64toNullInt64(testCaseAccessData[i].superUserTypeID),
			convertInt64toNullInt64(testCaseAccessData[i].fundSourceID),
			convertInt64toNullInt64(testCaseAccessData[i].adminEntityID),
			convertInt64toNullInt64(testCaseAccessData[i].fsAdminEntityID),
			convertInt64toNullInt64(testCaseAccessData[i].classID),
			convertInt64toNullInt64(testCaseAccessData[i].teacherTypeID),
			convertInt64toNullInt64(testCaseAccessData[i].teamChildID)}
	}
	return result
}

// creates sql.NullInt64 object from int, and set Valid = false if val = -1
func convertInt64toNullInt64(val int64) sql.NullInt64 {
	if val == -1 {
		return sql.NullInt64{Int64: 0, Valid: false}
	}
	return sql.NullInt64{Int64: val, Valid: true}
}

var testCases = []testCase{
	{ //1.SuperUser = rows[0].superUserTypeID != 0
		name:   "superuser test",
		result: `{"SuperUser":true}`,
		rows: []accessDataRowTest{
			{userTypeID: 6, adminTypeID: -1, fundSourceAdminTypeID: -1, superUserTypeID: 1, fundSourceID: -1,
				adminEntityID: -1, fsAdminEntityID: -1, classID: -1, teacherTypeID: -1, teamChildID: -1},
		},
	},
	{ //2.if rows[0].userTypeID == 3 && rows[0].adminTypeID == 0 => Admin=distinct(list(adminEntityID))
		name:   "admin test",
		result: `{"SuperUser":false,"Admin":{"ent":[111]}}`,
		rows: []accessDataRowTest{
			{userTypeID: 3, adminTypeID: 0, fundSourceAdminTypeID: -1, superUserTypeID: -1, fundSourceID: -1,
				adminEntityID: 111, fsAdminEntityID: -1, classID: -1, teacherTypeID: -1, teamChildID: -1},
		},
	},
	{ //3.if rows[0].userTypeID == 3 && rows[0].adminTypeID == 1 => VOAdmin=distinct(list(adminEntityID))
		name:   "voadmin test",
		result: `{"SuperUser":false,"VOAdmin":{"ent":[111]}}`,
		rows: []accessDataRowTest{
			{userTypeID: 3, adminTypeID: 1, fundSourceAdminTypeID: -1, superUserTypeID: -1, fundSourceID: -1,
				adminEntityID: 111, fsAdminEntityID: -1, classID: -1, teacherTypeID: -1, teamChildID: -1},
		},
	},
	{ //4.if rows[0].userTypeID == 3 && rows[0].adminTypeID == 2 => VONoChildAdmin=distinct(list(adminEntityID))
		name:   "vonochildadmin test",
		result: `{"SuperUser":false,"VONoChildAdmin":{"ent":[111]}}`,
		rows: []accessDataRowTest{
			{userTypeID: 3, adminTypeID: 2, fundSourceAdminTypeID: -1, superUserTypeID: -1, fundSourceID: -1,
				adminEntityID: 111, fsAdminEntityID: -1, classID: -1, teacherTypeID: -1, teamChildID: -1},
		},
	},
	{ //5.if rows[0].userTypeID == 7 && len(list(rows[0].FundSourceID)) && rows[0].fundSourceAdminTypeID == 0 => FSAdmin=distinct(list(fsAdminEntityID))
		name:   "fsadmin test",
		result: `{"SuperUser":false,"FSAdmin":{"ent":[222],"fundSrc":[111]}}`,
		rows: []accessDataRowTest{
			{userTypeID: 7, adminTypeID: -1, fundSourceAdminTypeID: 0, superUserTypeID: -1, fundSourceID: 111,
				adminEntityID: -1, fsAdminEntityID: 222, classID: -1, teacherTypeID: -1, teamChildID: -1},
		},
	},
	{ //6.if rows[0].userTypeID == 7 && len(list(rows[0].FundSourceID)) && rows[0].fundSourceAdminTypeID == 1 => FSVOAdmin=distinct(list(fsAdminEntityID))
		name:   "fsvoadmin test",
		result: `{"SuperUser":false,"FSVOAdmin":{"ent":[222],"fundSrc":[111]}}`,
		rows: []accessDataRowTest{
			{userTypeID: 7, adminTypeID: -1, fundSourceAdminTypeID: 1, superUserTypeID: -1, fundSourceID: 111,
				adminEntityID: -1, fsAdminEntityID: 222, classID: -1, teacherTypeID: -1, teamChildID: -1},
		},
	},
	{ //7.If rows[0].userTypeID == 1 || len(list (classID if teacherTypeID == 1)) >  => Teacher = distinct( list (classID if teacherTypeID == 1)  )
		name:   "teacher test",
		result: `{"SuperUser":false,"Teacher":{"cls":[222,333]}}`,
		rows: []accessDataRowTest{
			{userTypeID: 1, adminTypeID: -1, fundSourceAdminTypeID: -1, superUserTypeID: -1, fundSourceID: -1,
				adminEntityID: -1, fsAdminEntityID: -1, classID: 333, teacherTypeID: 1, teamChildID: -1},
			{userTypeID: 1, adminTypeID: -1, fundSourceAdminTypeID: -1, superUserTypeID: -1, fundSourceID: -1,
				adminEntityID: -1, fsAdminEntityID: -1, classID: 222, teacherTypeID: 1, teamChildID: -1},
		},
	},
	{ //8.If rows[0].userTypeID == 1 && len(list (classID if teacherTypeID == 2)) => CoTeacher = distinct( list (classID if teacherTypeID == 2)  )
		name:   "coteacher test",
		result: `{"SuperUser":false,"Teacher":{},"CoTeacher":{"cls":[333]}}`,
		rows: []accessDataRowTest{
			{userTypeID: 1, adminTypeID: -1, fundSourceAdminTypeID: -1, superUserTypeID: -1, fundSourceID: -1,
				adminEntityID: -1, fsAdminEntityID: -1, classID: 333, teacherTypeID: 2, teamChildID: -1},
		},
	},
	{ //9.If rows[0].userTypeID == 1&& len(list (classID if teacherTypeID == 3)) => AssistantTeacher = distinct( list (classID if teacherTypeID == 3)  )
		name:   "assistantteacher test",
		result: `{"SuperUser":false,"Teacher":{},"AssistantTeacher":{"cls":[333]}}`,
		rows: []accessDataRowTest{
			{userTypeID: 1, adminTypeID: -1, fundSourceAdminTypeID: -1, superUserTypeID: -1, fundSourceID: -1,
				adminEntityID: -1, fsAdminEntityID: -1, classID: 333, teacherTypeID: 3, teamChildID: -1},
		},
	},
	{ //10.If rows[0].UserTypeID == 5 || len(list(teamChildID)) > 0 =>TeamMember = distinct ( list(teamChildID) )
		name:   "team member test",
		result: `{"SuperUser":false,"TeamMember":{"kid":[444]}}`,
		rows: []accessDataRowTest{
			{userTypeID: 5, adminTypeID: -1, fundSourceAdminTypeID: -1, superUserTypeID: -1, fundSourceID: -1,
				adminEntityID: -1, fsAdminEntityID: -1, classID: -1, teacherTypeID: -1, teamChildID: 444},
		},
	},
}

type testCase struct {
	name   string
	result string
	rows   []accessDataRowTest
}

type accessDataRowTest struct {
	userTypeID            int64
	adminTypeID           int64
	fundSourceAdminTypeID int64
	superUserTypeID       int64
	fundSourceID          int64
	adminEntityID         int64
	fsAdminEntityID       int64
	classID               int64
	teacherTypeID         int64
	teamChildID           int64
}
