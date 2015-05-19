package main

import (
	"fmt"
	"github.com/elazarl/go-bindata-assetfs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

func bindata_read(data, name string) ([]byte, error) {
	var empty [0]byte
	sx := (*reflect.StringHeader)(unsafe.Pointer(&data))
	b := empty[:]
	bx := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bx.Data = sx.Data
	bx.Len = len(data)
	bx.Cap = bx.Len
	return b, nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindata_file_info struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindata_file_info) Name() string {
	return fi.name
}
func (fi bindata_file_info) Size() int64 {
	return fi.size
}
func (fi bindata_file_info) Mode() os.FileMode {
	return fi.mode
}
func (fi bindata_file_info) ModTime() time.Time {
	return fi.modTime
}
func (fi bindata_file_info) IsDir() bool {
	return false
}
func (fi bindata_file_info) Sys() interface{} {
	return nil
}

var _default_pac = "\x66\x75\x6e\x63\x74\x69\x6f\x6e\x20\x46\x69\x6e\x64\x50\x72\x6f\x78\x79\x46\x6f\x72\x55\x52\x4c\x28\x75\x72\x6c\x2c\x20\x68\x6f\x73\x74\x29\x0a\x7b\x0a\x20\x20\x72\x65\x74\x75\x72\x6e\x20\x22\x44\x49\x52\x45\x43\x54\x22\x3b\x0a\x7d\x0a"

func default_pac_bytes() ([]byte, error) {
	return bindata_read(
		_default_pac,
		"default.pac",
	)
}

func default_pac() (*asset, error) {
	bytes, err := default_pac_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "default.pac", size: 59, mode: os.FileMode(420), modTime: time.Unix(1432025197, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _htdocs_index_html = "\x3c\x68\x74\x6d\x6c\x3e\x0a\x3c\x68\x65\x61\x64\x3e\x3c\x2f\x68\x65\x61\x64\x3e\x0a\x3c\x62\x6f\x64\x79\x3e\x0a\x20\x20\x3c\x68\x31\x3e\x48\x65\x6c\x6c\x6f\x3c\x2f\x68\x31\x3e\x0a\x3c\x2f\x62\x6f\x64\x79\x3e\x0a\x3c\x2f\x68\x74\x6d\x6c\x3e\x0a"

func htdocs_index_html_bytes() ([]byte, error) {
	return bindata_read(
		_htdocs_index_html,
		"htdocs/index.html",
	)
}

func htdocs_index_html() (*asset, error) {
	bytes, err := htdocs_index_html_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "htdocs/index.html", size: 61, mode: os.FileMode(420), modTime: time.Unix(1431512444, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _tmpl_html_home_tmpl = "\x3c\x21\x64\x6f\x63\x74\x79\x70\x65\x20\x68\x74\x6d\x6c\x3e\x0a\x3c\x68\x74\x6d\x6c\x20\x6c\x61\x6e\x67\x3d\x22\x65\x6e\x22\x3e\x0a\x3c\x68\x65\x61\x64\x3e\x0a\x20\x20\x3c\x6d\x65\x74\x61\x20\x63\x68\x61\x72\x73\x65\x74\x3d\x22\x75\x74\x66\x2d\x38\x22\x3e\x0a\x20\x20\x3c\x74\x69\x74\x6c\x65\x3e\x7b\x7b\x20\x2e\x4e\x61\x6d\x65\x20\x7d\x7d\x20\x76\x7b\x7b\x20\x2e\x56\x65\x72\x73\x69\x6f\x6e\x20\x7d\x7d\x3c\x2f\x74\x69\x74\x6c\x65\x3e\x0a\x3c\x2f\x68\x65\x61\x64\x3e\x0a\x3c\x62\x6f\x64\x79\x3e\x0a\x20\x20\x3c\x68\x31\x3e\x7b\x7b\x20\x2e\x4e\x61\x6d\x65\x20\x7d\x7d\x20\x3c\x65\x6d\x3e\x76\x7b\x7b\x20\x2e\x56\x65\x72\x73\x69\x6f\x6e\x20\x7d\x7d\x3c\x2f\x65\x6d\x3e\x3c\x2f\x68\x31\x3e\x0a\x20\x20\x3c\x70\x3e\x3c\x63\x6f\x64\x65\x3e\x7b\x7b\x20\x2e\x50\x61\x63\x46\x69\x6c\x65\x6e\x61\x6d\x65\x20\x7d\x7d\x3c\x2f\x63\x6f\x64\x65\x3e\x3c\x2f\x70\x3e\x0a\x3c\x2f\x62\x6f\x64\x79\x3e\x0a\x3c\x2f\x68\x74\x6d\x6c\x3e\x0a"

func tmpl_html_home_tmpl_bytes() ([]byte, error) {
	return bindata_read(
		_tmpl_html_home_tmpl,
		"tmpl/html/home.tmpl",
	)
}

func tmpl_html_home_tmpl() (*asset, error) {
	bytes, err := tmpl_html_home_tmpl_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "tmpl/html/home.tmpl", size: 230, mode: os.FileMode(420), modTime: time.Unix(1432031703, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _tmpl_html_status_tmpl = "\x3c\x21\x64\x6f\x63\x74\x79\x70\x65\x20\x68\x74\x6d\x6c\x3e\x0a\x3c\x68\x74\x6d\x6c\x20\x6c\x61\x6e\x67\x3d\x22\x65\x6e\x22\x3e\x0a\x3c\x68\x65\x61\x64\x3e\x0a\x20\x20\x3c\x6d\x65\x74\x61\x20\x63\x68\x61\x72\x73\x65\x74\x3d\x22\x75\x74\x66\x2d\x38\x22\x3e\x0a\x20\x20\x3c\x74\x69\x74\x6c\x65\x3e\x7b\x7b\x20\x2e\x4e\x61\x6d\x65\x20\x7d\x7d\x20\x76\x7b\x7b\x20\x2e\x56\x65\x72\x73\x69\x6f\x6e\x20\x7d\x7d\x20\x2d\x20\x53\x74\x61\x74\x75\x73\x3c\x2f\x74\x69\x74\x6c\x65\x3e\x0a\x3c\x2f\x68\x65\x61\x64\x3e\x0a\x3c\x62\x6f\x64\x79\x3e\x0a\x20\x20\x3c\x68\x31\x3e\x7b\x7b\x20\x2e\x4e\x61\x6d\x65\x20\x7d\x7d\x20\x3c\x65\x6d\x3e\x76\x7b\x7b\x20\x2e\x56\x65\x72\x73\x69\x6f\x6e\x20\x7d\x7d\x3c\x2f\x65\x6d\x3e\x3c\x2f\x68\x31\x3e\x0a\x20\x20\x3c\x70\x3e\x3c\x63\x6f\x64\x65\x3e\x7b\x7b\x20\x2e\x50\x61\x63\x46\x69\x6c\x65\x6e\x61\x6d\x65\x20\x7d\x7d\x3c\x2f\x63\x6f\x64\x65\x3e\x3c\x2f\x70\x3e\x0a\x20\x20\x7b\x7b\x20\x69\x66\x20\x2e\x4b\x6e\x6f\x77\x6e\x50\x72\x6f\x78\x69\x65\x73\x20\x7d\x7d\x0a\x20\x20\x20\x20\x3c\x68\x33\x3e\x4b\x6e\x6f\x77\x6e\x20\x50\x72\x6f\x78\x69\x65\x73\x3c\x2f\x68\x33\x3e\x0a\x20\x20\x20\x20\x3c\x75\x6c\x3e\x0a\x20\x20\x20\x20\x7b\x7b\x72\x61\x6e\x67\x65\x20\x2e\x4b\x6e\x6f\x77\x6e\x50\x72\x6f\x78\x69\x65\x73\x7d\x7d\x0a\x20\x20\x20\x20\x20\x20\x3c\x6c\x69\x3e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x3c\x63\x6f\x64\x65\x3e\x7b\x7b\x2e\x41\x64\x64\x72\x65\x73\x73\x7d\x7d\x3c\x2f\x63\x6f\x64\x65\x3e\x20\x3c\x65\x6d\x3e\x2d\x20\x7b\x7b\x69\x66\x20\x2e\x49\x73\x41\x63\x74\x69\x76\x65\x7d\x7d\x41\x63\x74\x69\x76\x65\x7b\x7b\x65\x6c\x73\x65\x7d\x7d\x49\x6e\x61\x63\x74\x69\x76\x65\x7b\x7b\x65\x6e\x64\x7d\x7d\x3c\x2f\x65\x6d\x3e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x3c\x62\x72\x3e\x3c\x73\x6d\x61\x6c\x6c\x3e\x4c\x61\x73\x74\x20\x75\x70\x64\x61\x74\x65\x64\x20\x7b\x7b\x2e\x55\x70\x64\x61\x74\x65\x64\x2e\x46\x6f\x72\x6d\x61\x74\x20\x22\x32\x30\x30\x36\x2d\x30\x31\x2d\x30\x32\x54\x31\x35\x3a\x30\x34\x3a\x30\x35\x5a\x30\x37\x3a\x30\x30\x22\x7d\x7d\x3c\x2f\x73\x6d\x61\x6c\x6c\x3e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x7b\x7b\x69\x66\x20\x2e\x45\x72\x72\x6f\x72\x20\x7d\x7d\x3c\x62\x72\x3e\x3c\x73\x6d\x61\x6c\x6c\x3e\x7b\x7b\x20\x2e\x45\x72\x72\x6f\x72\x20\x7d\x7d\x3c\x2f\x73\x6d\x61\x6c\x6c\x3e\x7b\x7b\x65\x6e\x64\x7d\x7d\x0a\x20\x20\x20\x20\x20\x20\x3c\x2f\x6c\x69\x3e\x0a\x20\x20\x20\x20\x7b\x7b\x20\x65\x6e\x64\x20\x7d\x7d\x0a\x20\x20\x3c\x2f\x75\x6c\x3e\x0a\x20\x20\x7b\x7b\x20\x65\x6e\x64\x20\x7d\x7d\x0a\x3c\x2f\x62\x6f\x64\x79\x3e\x0a\x3c\x2f\x68\x74\x6d\x6c\x3e\x0a"

func tmpl_html_status_tmpl_bytes() ([]byte, error) {
	return bindata_read(
		_tmpl_html_status_tmpl,
		"tmpl/html/status.tmpl",
	)
}

func tmpl_html_status_tmpl() (*asset, error) {
	bytes, err := tmpl_html_status_tmpl_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "tmpl/html/status.tmpl", size: 625, mode: os.FileMode(420), modTime: time.Unix(1432032733, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _wpad_dat = "\x66\x75\x6e\x63\x74\x69\x6f\x6e\x20\x46\x69\x6e\x64\x50\x72\x6f\x78\x79\x46\x6f\x72\x55\x52\x4c\x28\x75\x72\x6c\x2c\x20\x68\x6f\x73\x74\x29\x0a\x7b\x0a\x20\x20\x69\x66\x20\x28\x69\x73\x49\x6e\x4e\x65\x74\x28\x68\x6f\x73\x74\x2c\x20\x22\x31\x32\x37\x2e\x30\x2e\x30\x2e\x30\x22\x2c\x20\x22\x32\x35\x35\x2e\x30\x2e\x30\x2e\x30\x22\x29\x29\x0a\x20\x20\x7b\x0a\x20\x20\x20\x20\x72\x65\x74\x75\x72\x6e\x20\x22\x44\x49\x52\x45\x43\x54\x22\x3b\x0a\x20\x20\x7d\x0a\x20\x20\x72\x65\x74\x75\x72\x6e\x20\x22\x50\x52\x4f\x58\x59\x20\x7b\x7b\x2e\x48\x54\x54\x50\x48\x6f\x73\x74\x7d\x7d\x22\x3b\x0a\x7d\x0a"

func wpad_dat_bytes() ([]byte, error) {
	return bindata_read(
		_wpad_dat,
		"wpad.dat",
	)
}

func wpad_dat() (*asset, error) {
	bytes, err := wpad_dat_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "wpad.dat", size: 148, mode: os.FileMode(420), modTime: time.Unix(1432024866, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"default.pac":           default_pac,
	"htdocs/index.html":     htdocs_index_html,
	"tmpl/html/home.tmpl":   tmpl_html_home_tmpl,
	"tmpl/html/status.tmpl": tmpl_html_status_tmpl,
	"wpad.dat":              wpad_dat,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type _bintree_t struct {
	Func     func() (*asset, error)
	Children map[string]*_bintree_t
}

var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"default.pac": &_bintree_t{default_pac, map[string]*_bintree_t{}},
	"htdocs": &_bintree_t{nil, map[string]*_bintree_t{
		"index.html": &_bintree_t{htdocs_index_html, map[string]*_bintree_t{}},
	}},
	"tmpl": &_bintree_t{nil, map[string]*_bintree_t{
		"html": &_bintree_t{nil, map[string]*_bintree_t{
			"home.tmpl":   &_bintree_t{tmpl_html_home_tmpl, map[string]*_bintree_t{}},
			"status.tmpl": &_bintree_t{tmpl_html_status_tmpl, map[string]*_bintree_t{}},
		}},
	}},
	"wpad.dat": &_bintree_t{wpad_dat, map[string]*_bintree_t{}},
}}

// Restore an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, path.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// Restore assets under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	if err != nil { // File
		return RestoreAsset(dir, name)
	} else { // Dir
		for _, child := range children {
			err = RestoreAssets(dir, path.Join(name, child))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

func assetFS() *assetfs.AssetFS {
	for k := range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: k}
	}
	panic("unreachable")
}
