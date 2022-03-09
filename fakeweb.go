package fakeweb

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

var (
	Seed int64 = 1
	rng  *rand.Rand
)

var (
	HostNameMinLen   int = 4
	HostNameMaxLen   int = 8
	DirNameMinLen    int = 4
	DirNameMaxLen    int = 6
	MaxDirDepth      int = 3
	MaxSubdirsPerDir int = 3
	MaxFilesPerDir   int = 3
	MaxLinksPerFile  int = 4
)

var (
	Sites []*Site
	web   map[string]*Site
)

func word(min, max int) string {
	length := rng.Intn(max-min+1) + min
	str := ""
	for i := 0; i < length; i++ {
		str += string(rng.Intn('z'-'a') + 'a')
	}
	return str
}

/*
unusedWord will be as such when generics are out:
func unusedWord[V any](min, max int, uses *map[string]V) string {
	attempts := 10
	for i := 0; i < attempts; i++ {
		w := word(min, max)
		if _, used := (*uses)[w]; !used {
			return w
		}
	}
	panic(fmt.Errorf("could not generate a unique word after %v attempts; try increasing max length", attempts))
}
but for now there is this terrible version as a function:
*/
func unusedWord(min, max int, inUse func(string) bool) string {
	attempts := 10
	for i := 0; i < attempts; i++ {
		w := word(min, max)
		if !inUse(w) {
			return w
		}
	}
	panic(fmt.Errorf("could not generate a unique word after %v attempts; try increasing max length", attempts))
}

// As in 'can be in a file path'. TODO: better name please
type pathable interface {
	find(string) (*file, error)
}

type file struct {
	name    *string
	content string
	parent  *dir
}

func (f *file) find(path string) (*file, error) {
	if path != "" {
		return nil, fmt.Errorf("file not found")
	}
	return f, nil
}

func (f *file) getPath() string {
	path := *f.name
	parent := f.parent
	for parent != nil {
		path = *parent.name + "/" + path
		parent = parent.parent
	}
	return path
}

type dir struct {
	name    *string
	content map[string]pathable
	parent  *dir
}

func (d *dir) find(path string) (*file, error) {
	if path == "" {
		return nil, fmt.Errorf("path was to a directory")
	}
	path = strings.TrimPrefix(path, "/")
	sep := strings.IndexByte(path, '/')
	if sep == -1 {
		sep = len(path)
	}
	name := path[:sep]
	switch name {
	case "", ".":
		return d.find(path[sep:])
	case "..":
		if d.parent == nil {
			return nil, fmt.Errorf(".. in path with no parent")
		}
		return d.parent.find(path[sep:])
	}
	f, ok := d.content[name]
	if !ok {
		fmt.Println(path)
		return nil, fmt.Errorf("file not found")
	}
	return f.find(path[sep:])
}

func (d *dir) print(tabs int) {
	fmt.Printf("/%v\n", *d.name)
	for _, sub := range d.content {
		for j := 0; j < tabs; j++ {
			fmt.Print("\t")
		}
		switch v := sub.(type) {
		case *dir:
			v.print(tabs + 1)
		case *file:
			fmt.Printf("/%v (file) %v\n", *v.name, strings.ReplaceAll(v.content, "\n", " "))
		}
	}
}

func newDirWithChildren(name *string, parent *dir, depth int, registerTo *[]*file) *dir {
	d := &dir{name: name, content: make(map[string]pathable), parent: parent}
	unused := func(w string) bool { _, c := d.content[w]; return c }
	fileCount := rng.Intn(MaxFilesPerDir-1) + 1
	for i := 0; i < fileCount; i++ {
		newName := unusedWord(DirNameMinLen, DirNameMaxLen, unused)
		f := &file{name: &newName, parent: d}
		d.content[newName] = f
		*registerTo = append(*registerTo, f)
	}
	if depth > 0 {
		depth--
		subdirCount := rng.Intn(MaxSubdirsPerDir)
		for i := 0; i < subdirCount; i++ {
			newName := unusedWord(DirNameMinLen, DirNameMaxLen, unused)
			d.content[newName] = newDirWithChildren(&newName, d, depth, registerTo)
		}
	}
	return d
}

type Site struct {
	scheme, host string
	root         *dir
	files        []*file
}

// TODO: better linking (links that include '.', '..', ect.)
func (s *Site) populate() {
	for _, f := range s.files {
		totalLinks := rng.Intn(MaxLinksPerFile-1) + 1
		for i := 0; i < totalLinks; i++ {
			// My webcrawler looks for this, but it can be whatever
			f.content += "<a href=\"" + Sites[rng.Intn(len(Sites))].GetRandLink() + "\"><\\a>\n"
		}
	}
}

func (s *Site) GetRandLink() string {
	return s.scheme + "://" + s.host + s.files[rng.Intn(len(s.files))].getPath()
}

func (s *Site) Print() {
	fmt.Println(s.scheme + "://" + s.host)
	s.root.print(1)
}

func Init(size int) {
	rng = rand.New(rand.NewSource(Seed))
	web = make(map[string]*Site)
	Sites = make([]*Site, size)
	for i := 0; i < size; i++ {
		s := &Site{scheme: "https", host: word(HostNameMinLen, HostNameMaxLen) + ".com"}
		s.host = unusedWord(HostNameMinLen, HostNameMaxLen, func(w string) bool { _, c := web[s.scheme+"://"+w+".com"]; return c }) + ".com"
		noname := ""
		s.root = newDirWithChildren(&noname, nil, MaxDirDepth, &s.files)
		Sites[i] = s
		web[s.scheme+"://"+s.host] = s
	}
	for i := 0; i < size; i++ {
		Sites[i].populate()
	}
}

func Get(urlstr string) (*http.Response, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	s, ok := web[u.Scheme+"://"+u.Host]
	if !ok {
		return nil, fmt.Errorf("no such site")
	}
	f, err := s.root.find(u.Path)
	if err != nil {
		return nil, err
	}
	fakeResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "not a real one",
		Body:       io.NopCloser(strings.NewReader(f.content)),
	}
	return fakeResp, nil
}

