package processfile_test

import (
	"testing"
	"bytes"
	"github.com/frericksm/pride/utils"
	"github.com/frericksm/pride/processfile"
	"io/ioutil"
)


func TestFromFile(t *testing.T) {

	content := processfile.FileContent("testdata/A1.process")

	p := processfile.FromBytes(content)
	if id := p.Id; id != "de.michael.A1" {
		t.Errorf("Expected Id de.michael.A1,  but was %s:", id)
	}

	if l := len(p.FormalParameters); l != 3 {
		t.Errorf("Expected 3 formal parameters, but ist was %d:", l)
	}

	if l := len(p.Activities); l != 5 {
		t.Errorf("Expected 5 activities, but ist was %d:", l)
	}

	//fmt.Println("%s" , p)
}

func TestToString(t *testing.T) {
	content := processfile.FileContent("testdata/A1.process")
	p := processfile.FromBytes(content)
	content2 := processfile.ToBytes(p)

	err := ioutil.WriteFile("/home/michael/data/isp/michael-1.0.0/de/michael/A1.process_", content2, 0644)
	utils.Check(err)

	if !bytes.Equal(content,content2) {
//		t.Errorf("Roundtrip failed")
	}

/*
	for _, a := range p.Activities {
		t.Logf("Type %s, ImplementationRefId: %s" , a.Body.ActivityType, a.Body.ImplementationRefId)
	}
*/

	//t.Log(content)
	//t.Log(content2)

}
