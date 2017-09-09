package processfile

import 
(
	"encoding/xml"
//	"bytes"
	"fmt"
	"io/ioutil"
//	"os"
	"github.com/frericksm/pride/utils"
)

type Process struct {
	XMLName xml.Name `xml:"process"`
	Id  string    `xml:"id,attr"`
	Name  string    `xml:"name,attr"`
	Description Description  `xml:"description"`
	FormalParameters []FormalParameter `xml:"formal-parameters>formal-parameter"`
	Variables        []Variable `xml:"variables>variable"`
	Properties       []Property `xml:"properties>property"`
	Activities       []Activity `xml:"activities>activity"`
}

type Description struct {
	Value []byte `xml:",innerxml"`
}

type FormalParameter struct {
	Id  string    `xml:"id,attr"`
	Name  string    `xml:"name,attr"`
	Description Description  `xml:"description"`
	Direction string `xml:"direction,attr"`
	Hidden bool      `xml:"hidden,attr"`
	Required bool      `xml:"required,attr"`
}

type Variable struct {
	Id  string    `xml:"id,attr"`
	Name  string    `xml:"name,attr"`
	Hidden bool      `xml:"hidden,attr"`
}

type Property struct {
	Id  string    `xml:"id,attr"`
	Name  string    `xml:"name,attr"`
	Value string  `xml:"value,attr"`
	Description Description  `xml:"description"`
}

type Activity struct {
	Id  string    `xml:"id,attr"`
	Name  string    `xml:"name,attr"`
	Body Body  `xml:"body"`
	Transitions []Transition  `xml:"transitions>transition"`
}

type Body struct {
	ActivityType  string    `xml:"activity-type,attr"`
	EventType    string    `xml:"event-type,attr"`
	ImplementationRefId  string    `xml:"implementation-ref-id,attr"`
	ImplementationType  string    `xml:"implementation-type,attr"`
	DataMappings []DataMapping  `xml:"data-mappings>data-mapping"`
	NodeGraphicsInfo NodeGraphicsInfo `xml:"node-graphics-info"`
}

type DataMapping  struct {
	FormalParameter string  `xml:"formal-parameter,attr"`
	ActualParameter ActualParameter  `xml:"actual-parameter"`
}

type ActualParameter struct {
	Value []byte `xml:",innerxml"`
}

type NodeGraphicsInfo struct {
	CoordinateX string  `xml:"coordinate-x,attr"`
	CoordinateY string  `xml:"coordinate-y,attr"`
	With        string  `xml:"width,attr"`
	Height      string  `xml:"height,attr"`
}

type Transition struct {
	Id  string    `xml:"id,attr"`
	To  string    `xml:"to,attr"`
	Condition Condition  `xml:"condition"`
}

type Condition struct {
	Value string  `xml:",chardata"`
}

func (p Process) String() string {
	return fmt.Sprintf("%s - %s", p.Name, p.Description)
}

func FileContent (filepath string) []byte {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		panic(err)
	}
	return content
}

func FromBytes(content []byte) *Process {
	var p Process
	error :=xml.Unmarshal(content, &p)
	utils.Check(error)
	return &p
}

func ToBytes(p *Process) []byte  {
	content, error := xml.MarshalIndent(*p, "", "  ")
	utils.Check(error)
	content = append([]byte(xml.Header), content...)
	return content	
}


