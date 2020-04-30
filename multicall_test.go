package xmlrpc

import (
	"io/ioutil"
	"testing"
)

func Test_splitMulticall(t *testing.T) {
	b, err := ioutil.ReadFile("fixtures/multicall_error.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = splitMulticall(b)
	mf, ok := err.(MulticallFault)
	if !ok {
		t.Errorf("expected multicall fault, got %s", err)
	}
	if mf.Index != 1 {
		t.Errorf("wrong position for fault %d", mf.Index)
	}

	b, err = ioutil.ReadFile("fixtures/multicall_ok.xml")
	if err != nil {
		t.Fatal(err)
	}
	out, err := splitMulticall(b)
	if err != nil {
		t.Error(err)
	}
	if L := len(out); L != 2 {
		t.Errorf("expected 2 answers, got %d", L)
	}
}

func TestUnmarshal(t *testing.T) {
	b, err := ioutil.ReadFile("fixtures/multicall_ok.xml")
	if err != nil {
		t.Fatal(err)
	}
	type data struct {
		NbFiles int `xmlrpc:"nbfiles"`
	}
	var d1, d2 data
	out := []interface{}{&d1, &d2}
	err = ResponseMulticall(b).Unmarshal(out)
	if err != nil {
		t.Error(err)
	}
	if nb1 := d1.NbFiles; nb1 != 4 {
		t.Errorf("expected 4, got %d", nb1)
	}
	if nb2 := d2.NbFiles; nb2 != 1 {
		t.Errorf("expected 4, got %d", nb2)
	}

	outArray := [2]interface{}{&d1, &d2}
	err = ResponseMulticall(b).Unmarshal(outArray)
	if err != nil {
		t.Error(err)
	}
	if nb1 := d1.NbFiles; nb1 != 4 {
		t.Errorf("expected 4, got %d", nb1)
	}
	if nb2 := d2.NbFiles; nb2 != 1 {
		t.Errorf("expected 4, got %d", nb2)
	}

	var outWrong string
	err = ResponseMulticall(b).Unmarshal(&outWrong)
	if err == nil {
		t.Error("expected error")
	}

}
