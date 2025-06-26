package data

import "testing"

func TestIDGen(t *testing.T) {
	InitMgo("127.0.0.1", "27017", "", "", "test")
	idgen := InitIDGen(FACEID_KEY)
	for i := 0; i < 400; i++ {
		id := idgen.GenID()
		t.Log(id)
	}
}
