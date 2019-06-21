package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []node     `xml:",any"`
}

// UnmarshalXML used for decoding XML
func (n *node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type tnode node

	return d.DecodeElement((*tnode)(n), &start)
}

func walk(nodes []node, f func(node) (bool, error)) error {
	for _, n := range nodes {
		b, err := f(n)
		if err != nil {
			return err
		}
		if b {
			if err = walk(n.Nodes, f); err != nil {
				return err
			}
		}
	}
	return nil
}

func matchAttr(a []xml.Attr, attrreg []AttrReg) bool {
	for _, x := range attrreg {
		found := false
		for _, y := range a {

			if x.Key.MatchString(y.Name.Local) {
				if x.Value.MatchString(y.Value) {
					found = true
					break
				}
			}
		}

		if !found {
			return false
		}

	}

	return true
}

func highLight(input string, reg *regexp.Regexp) string {

	m := reg.FindStringSubmatch(input)
	l := len(m)
	c := 31
	for index, i := range m {
		if l < 1 && index == 0 {
			continue
		}
		cpart := fmt.Sprintf("\x1b[%d;1m", c)
		input = strings.ReplaceAll(input, i, cpart+i+"\x1b[0m")
		c++
	}
	return input
}

func matchXML(filename string, nreg, treg, creg *regexp.Regexp, attrreg []AttrReg, colors bool) error {

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("unable to parse: %v", err)
	}

	buf := bytes.NewBuffer(b)
	dec := xml.NewDecoder(buf)

	var n node
	if err = dec.Decode(&n); err != nil {
		return fmt.Errorf("unable to decode XML: %v", err)
	}

	if err = walk([]node{n}, func(n node) (bool, error) {
		if nreg.MatchString(n.XMLName.Space) {

			if colors {
				n.XMLName.Space = highLight(n.XMLName.Space, nreg)
			}

			if treg.MatchString(n.XMLName.Local) {

				if colors {
					n.XMLName.Local = highLight(n.XMLName.Local, treg)
				}

				if matchAttr(n.Attrs, attrreg) {

					c := string(n.Content)

					if creg.MatchString(c) {
						if colors {
							c = highLight(html.UnescapeString(c), creg)
						}

						fmt.Println(
							fmt.Sprintf("\nFilename:%s\nNamespace: %s\nTag: %s\nAttributes: %v\nContent:\n%s\n",
								filename,
								n.XMLName.Space,
								n.XMLName.Local,
								n.Attrs,
								c,
							),
						)

					}
				}
			}
		}

		return true, nil
	}); err != nil {
		return fmt.Errorf("unable to walk: %v", err)
	}

	return nil
}

type flagList []string

func (i *flagList) String() string {
	return "my string representation"
}

func (i *flagList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type AttrReg struct {
	Key   *regexp.Regexp
	Value *regexp.Regexp
}

func compileParams(namespace, tag, content string) (*regexp.Regexp, *regexp.Regexp, *regexp.Regexp, error) {

	nreg, err := regexp.Compile(namespace)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("namespace regex: %v", err)
	}

	treg, err := regexp.Compile(tag)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("tag regex: %v", err)
	}

	creg, err := regexp.Compile(content)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("content regex: %v", err)
	}

	return nreg, treg, creg, nil
}

func compileAttr(attrlist flagList) ([]AttrReg, error) {

	attrreg := make([]AttrReg, 0, len(attrlist))
	for _, x := range attrlist {
		t := strings.Split(x, "=")
		if len(t) != 2 {
			log.Fatalf("Unable to split attr %s", x)
		}

		var a AttrReg
		var err error

		a.Key, err = regexp.Compile(t[0])
		if err != nil {
			return nil, fmt.Errorf("unable to compile key in %s: %v", t[0], err)
		}
		a.Value, err = regexp.Compile(t[1])
		if err != nil {
			return nil, fmt.Errorf("unable to compile value in %s: %v", t[0], err)
		}

		attrreg = append(attrreg, a)
	}

	return attrreg, nil
}

func work(files []string, namespace, tag, content string, attrlist flagList, cores int, colors bool) {
	attrreg, err := compileAttr(attrlist)
	if err != nil {
		log.Fatalf("Unable to compile attribute list: %v", err)
	}

	nreg, treg, creg, err := compileParams(namespace, tag, content)
	if err != nil {
		log.Fatalf("Unable to compile: %v", err)
	}

	sem := make(chan bool, cores)
	var wg sync.WaitGroup

	for _, filename := range files {
		wg.Add(1)
		sem <- true

		go func(filename string, nreg, treg, creg *regexp.Regexp, attrreg []AttrReg, colors bool) {
			defer wg.Done()
			defer func() {
				<-sem
			}()
			err = matchXML(filename, nreg, treg, creg, attrreg, colors)
			if err != nil {
				log.Printf("Unable to parse %s : %v", filename, err)
			}
		}(filename, nreg, treg, creg, attrreg, colors)
	}

	wg.Wait()
}

func main() {

	var (
		tag       string
		namespace string
		content   string
		attrlist  flagList
		cores     int
		colors    bool
	)

	flag.StringVar(&tag, "tag", "", "Regex to match a tag")
	flag.StringVar(&namespace, "namespace", "", "Regex to match a namespace")
	flag.StringVar(&content, "content", "", "Regex to match content")
	flag.IntVar(&cores, "cores", runtime.NumCPU(), "Number of cores to run on")
	flag.BoolVar(&colors, "color", false, "Enable color highlight")
	flag.Var(&attrlist, "attr", "Attr key/value")

	flag.Parse()

	work(flag.Args(), namespace, tag, content, attrlist, cores, colors)

}
