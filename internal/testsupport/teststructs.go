package testsupport

// SingleStringPmsStruct with single string
type SingleStringPmsStruct struct {
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

// StructPmsWithNonExistantVariable that has one var that is not
// backed in the PMS
type StructPmsWithNonExistantVariable struct {
	Name string `pms:"test, prefix=simple"`
	Sub  struct {
		Apa     int    `pms:"ext"`
		Nu      string `pms:"myname"`
		Missing string `pms:"gonemissing"`
	}
	AsmSub struct {
		Apa2     int    `asm:"ext"`
		Nu2      string `asm:"myname"`
		Missing2 string `asm:"gonemissing"`
	}
}

// MyDbServiceConfig is a fake test struct
type MyDbServiceConfig struct {
	Name       string `pms:"test, prefix=simple,tag1=nanna banna panna"`
	Connection struct {
		User     string `json:"user"`
		Password string `json:"password,omitempty"`
		Timeout  int    `json:"timeout"`
	} `pms:"bubbibobbo"`
}

// MyDbServiceConfigAsm is a fake test struct
type MyDbServiceConfigAsm struct {
	Name       string
	Connection struct {
		User     string `json:"user"`
		Password string `json:"password,omitempty"`
		Timeout  int    `json:"timeout"`
	} `asm:"bubbibobbo, strkey=password"`
}
