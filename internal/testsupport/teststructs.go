package testsupport

// SingleStringStruct with single string
type SingleStringStruct struct {
	Name string `pms:"test, prefix=simple,tag1=nanna banna panna"`
}

// SingleStringAsmStruct with single string
type SingleStringAsmStruct struct {
	Name string `asm:"test, prefix=simple"`
}

// StructWithSubStruct with single sub-struct
type StructWithSubStruct struct {
	Name string `pms:"test, prefix=simple"`
	Sub  struct {
		Apa int    `pms:"ext"`
		Nu  string `pms:"myname"`
	}
	AsmSub struct {
		Apa2 int    `asm:"ext"`
		Nu2  string `asm:"myname"`
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
