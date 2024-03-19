package goscript

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/goscript/yaegi"
)

const (
	SCRIPT_LANGUAGE_GO = yaegi.SCRIPT_LANGUAGE_GO
)

type Script struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type Scripts []Script

func (ss Scripts) GroupByLanguage() (groupd map[string]Scripts) {
	groupd = map[string]Scripts{}
	for _, script := range ss {
		if _, ok := groupd[script.Language]; !ok {
			groupd[script.Language] = make(Scripts, 0)
		}
		groupd[script.Language] = append(groupd[script.Language], script)
	}
	return groupd
}

type ScriptI interface {
	Language() string
	Compile() (err error)
	WriteCode(codes ...string)
	Run(script string) (out string, err error)
	CallFuncScript(funcName string, input string) (callFuncScript string)                            //最终调用函数代码
	GetSymbolFromScript(selector string, dstType reflect.Type) (destSymbol reflect.Value, err error) //从脚本中获取符号
}

type ScriptIs []ScriptI

func (sis *ScriptIs) Add(ss ...ScriptI) {
	*sis = append(*sis, ss...)
}

var (
	ERROR_NOT_FOUND_SCRIPTI_BY_LANGUAGE = errors.New("not found script by language")
)

func (sis *ScriptIs) GetByLanguage(language string) (scriptI ScriptI, err error) {
	for _, s := range *sis {
		if strings.EqualFold(language, s.Language()) {
			return s, nil
		}
	}
	err = errors.WithMessagef(ERROR_NOT_FOUND_SCRIPTI_BY_LANGUAGE, "language :%s", language)
	return nil, err
}

func NewScriptEngine(language string) (scriptI ScriptI, err error) {
	m := map[string]func() ScriptI{
		SCRIPT_LANGUAGE_GO: func() ScriptI {
			return yaegi.NewScriptGo()
		},
	}
	fn, ok := m[strings.ToLower(language)]
	if !ok {
		err = errors.Errorf("not found script engine by language:%s", language)
		return nil, err
	}
	return fn(), nil
}
