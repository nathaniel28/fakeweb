# fakeweb

## Overview

Package fakeweb provides generation of and communication with fake servers. These 'servers' are simply structs on your computer.

They are generated with Init.

They can be found using a function similar to http.Get, also called Get:

```
// Init must be called prior to the following.
randURL := fakeweb.RandSite().RandLink()
resp, err := fakeweb.Get(randURL)
if err != nil {
	// handle error (in this case however, Get will always succeed since
	// the string passed to it was generated from one of the existing sites)
}
// use resp...
```

## Variables

```
var (
	// Seed is the number used to seed random during Init.
	Seed int64 = 1
)
```

```
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
```

```
var (
	// Sites is a slice of pointers to all websites created by Init.
	Sites []*Site
)
```

## Functions

### func Init
    func Init(size int)
Init creates size number of Sites that can be found with Get. After init is called once, it may be unsafe to call again.

### func RandSite
    func RandSite() *Site
Calls to RandSite before Init is called will cause a panic. RandSite returns a random \*Site from Sites.

### func Get
    func Get(urlstr string) (resp *http.Response, err error)
Calls to Get before Init is called will cause a panic. The only fields of resp containing usefull information are resp.StatusCode and resp.Body. Links found in resp.Body are strings such as \<a href="example.com/foo"><\\a>, though resp.Body is not valid html.

Get returns an error if it cannot parse urlstr or find the desired host or find the desired path. If err is nil, resp is non-nil. Currently, if resp is non-nil, resp.StatusCode == 200.

This function is meant to mimic that of http.Get.

## Types

### type Site
```
type Site struct {
	// contains filtered or unexported fields
}
```

### func (\*Site) RandLink
    func (s *Site) RandLink() string
RandLink returns a link to a random file in Site s. This link can be passed to Get.

### func (\*Site) Print
    func (s *Site) Print()

## Example

```
package main

import (
	"fmt"
	"io"
	"github.com/nathaniel28/fakeweb"
)

func main() {
	fakeweb.Init(10)
	
	resp, err := fakeweb.Get(fakeweb.RandSite().RandLink())
	if err != nil {
		// This is unnecessary, as in this case Get will always
		// succeed since the link passed to it is from RandLink
		fmt.Println(err)
		return
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	fmt.Printf("%s\n", body)
}
```
