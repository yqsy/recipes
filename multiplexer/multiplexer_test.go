package main

import "testing"

func TestIdGenSimple(t *testing.T) {
	var idGen IdGen

	idGen.initWithMaxId(65535)

	freeId, err := idGen.getFreeId()

	if err != nil {
		t.Fatalf("err?%v", err)
	}

	if freeId != 0 {
		t.Fatalf("id?%v", freeId)
	}

	freeIdNum := idGen.getFreeIdNum()
	if freeIdNum != 65534 {
		t.Fatalf("free num?%v", freeIdNum)
	}

	freeId, err = idGen.getFreeId()
	if freeId != 1 {
		t.Fatalf("id?%v", freeId)
	}

	freeIdNum = idGen.getFreeIdNum()
	if freeIdNum != 65533 {
		t.Fatalf("free num?%v", freeIdNum)
	}

	idGen.releaseFreeId(0)
	idGen.releaseFreeId(1)

	freeIdNum = idGen.getFreeIdNum()
	if freeIdNum != 65535 {
		t.Fatalf("free num?%v", freeIdNum)
	}

	freeId, err = idGen.getFreeId()
	if freeId != 2 {
		t.Fatalf("id?%v", freeId)
	}

	freeIdNum = idGen.getFreeIdNum()
	if freeIdNum != 65534 {
		t.Fatalf("free num?%v", freeIdNum)
	}

	for i := 0; i < 65534; i++ {
		_, _ = idGen.getFreeId()
	}

	freeId, err = idGen.getFreeId()
	if err == nil {
		t.Fatalf("get freeid?%v", freeId)
	}
}
