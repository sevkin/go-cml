package xml

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"os"
)

// Read reads and unmarshal from io.Reader
func Read(r io.Reader) (*КоммерческаяИнформация, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	x := new(КоммерческаяИнформация)
	err = xml.Unmarshal(buf, x)
	if err != nil {
		return nil, err
	}
	return x, nil
}

// ReadFile reads and unmarshal from file
func ReadFile(fname string) (*КоммерческаяИнформация, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Read(f)
}

//ReadMust reads and unmarshal from file or panic
func ReadMust(fname string) *КоммерческаяИнформация {
	x, err := ReadFile(fname)
	if err != nil {
		panic(err)
	}
	return x
}

// Write writes x to w. Adds preceding xml declaration
func Write(x *КоммерческаяИнформация, w io.Writer) error {
	buf, err := xml.MarshalIndent(x, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>\n"))
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	return nil
}
