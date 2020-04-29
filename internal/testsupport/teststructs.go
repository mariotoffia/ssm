package testsupport

// SingleStringStruct with single string
type SingleStringStruct struct {
	Name string `pms:"test, prefix=simple,tag1=nanna banna panna"`
}

// StructWithSubStruct with single sub-struct
type StructWithSubStruct struct {
	Name string `pms:"test, prefix=simple"`
	Sub  struct {
		Apa int    `pms:"ext"`
		Nu  string `pms:"myname"`
	}
}

// StructWithNonExistantVariable that has one var that is not
// backed in the PMS
type StructWithNonExistantVariable struct {
	Name string `pms:"test, prefix=simple"`
	Sub  struct {
		Apa     int    `pms:"ext"`
		Nu      string `pms:"myname"`
		Missing string `pms:"gonemissing"`
	}
}
