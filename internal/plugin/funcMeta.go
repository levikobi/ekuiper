package plugin

import (
	"fmt"
	"github.com/lf-edge/ekuiper/internal/conf"
	"github.com/lf-edge/ekuiper/internal/pkg/filex"
	"io/ioutil"
	"path"
	"strings"
)

type (
	fileFunc struct {
		Name    string        `json:"name"`
		Example string        `json:"example"`
		Hint    *fileLanguage `json:"hint"`
	}
	fileFuncs struct {
		About   *fileAbout  `json:"about"`
		Name    string      `json:"name"`
		FiFuncs []*fileFunc `json:"functions"`
	}
	uiFunc struct {
		Name    string    `json:"name"`
		Example string    `json:"example"`
		Hint    *language `json:"hint"`
	}
	uiFuncs struct {
		About   *about    `json:"about"`
		Name    string    `json:"name"`
		UiFuncs []*uiFunc `json:"functions"`
	}
)

func isInternalFunc(fiName string) bool {
	internal := []string{`accumulateWordCount.json`, `countPlusOne.json`, `echo.json`, `internal.json`, "windows.json", "image.json", "geohash.json"}
	for _, v := range internal {
		if v == fiName {
			return true
		}
	}
	return false
}
func newUiFuncs(fi *fileFuncs) *uiFuncs {
	if nil == fi {
		return nil
	}
	uis := new(uiFuncs)
	uis.About = newAbout(fi.About)
	uis.Name = fi.Name
	for _, v := range fi.FiFuncs {
		ui := new(uiFunc)
		ui.Name = v.Name
		ui.Example = v.Example
		ui.Hint = newLanguage(v.Hint)
		uis.UiFuncs = append(uis.UiFuncs, ui)
	}
	return uis
}

var gFuncmetadata map[string]*uiFuncs

func (m *Manager) readFuncMetaDir() error {
	gFuncmetadata = make(map[string]*uiFuncs)
	confDir, err := conf.GetConfLoc()
	if nil != err {
		return err
	}

	dir := path.Join(confDir, "functions")
	files, err := ioutil.ReadDir(dir)
	if nil != err {
		return err
	}
	for _, file := range files {
		fname := file.Name()
		if !strings.HasSuffix(fname, ".json") {
			continue
		}

		if err := m.readFuncMetaFile(path.Join(dir, fname)); nil != err {
			return err
		}
	}
	return nil
}

func (m *Manager) uninstalFunc(name string) {
	if ui, ok := gFuncmetadata[name+".json"]; ok {
		if nil != ui.About {
			ui.About.Installed = false
		}
	}
}
func (m *Manager) readFuncMetaFile(filePath string) error {
	fiName := path.Base(filePath)
	fis := new(fileFuncs)
	err := filex.ReadJsonUnmarshal(filePath, fis)
	if nil != err {
		return fmt.Errorf("filePath:%s err:%v", filePath, err)
	}
	if nil == fis.About {
		return fmt.Errorf("not found about of %s", filePath)
	} else if isInternalFunc(fiName) {
		fis.About.Installed = true
	} else {
		_, fis.About.Installed = m.registry.Get(FUNCTION, strings.TrimSuffix(fiName, `.json`))
	}
	gFuncmetadata[fiName] = newUiFuncs(fis)
	conf.Log.Infof("funcMeta file : %s", fiName)
	return nil
}
func GetFunctions() (ret []*uiFuncs) {
	for _, v := range gFuncmetadata {
		ret = append(ret, v)
	}
	return ret
}