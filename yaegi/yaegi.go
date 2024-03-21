package yaegi

import (
	"fmt"
	"io/fs"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	_ "github.com/spf13/cast"
	_ "github.com/syyongx/php2go"
	_ "github.com/tidwall/gjson"
	_ "github.com/tidwall/sjson"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const (
	SCRIPT_LANGUAGE_GO = "go"
)

type OutputDTO struct {
	Data string
	Err  error
}

func init() {
	Symbols["github.com/suifengpiao14/goscript/yaegi/yaegi"] = map[string]reflect.Value{
		// type definitions
		"OutputDTO": reflect.ValueOf((*OutputDTO)(nil)),
	}
}

type ScriptGo struct {
	engine                *interp.Interpreter
	code                  []string
	symbols               map[string]map[string]reflect.Value
	_SourcecodeFilesystem fs.FS
}

func (sgo ScriptGo) Language() string {
	return SCRIPT_LANGUAGE_GO
}

func (sgo *ScriptGo) WriteCode(codes ...string) {
	sgo.code = append(sgo.code, codes...)
}

func (sgo *ScriptGo) Compile() (err error) {
	engine := interp.New(interp.Options{
		SourcecodeFilesystem: sgo._SourcecodeFilesystem,
	})
	engine.Use(stdlib.Symbols)
	engine.Use(Symbols)     //注册当前包结构体
	engine.Use(sgo.symbols) // 使用当前符号
	for _, code := range sgo.code {
		_, err = engine.Eval(code)
		if err != nil {
			err = errors.WithMessage(err, "init dynamic go script error")
			return err
		}
	}
	sgo.engine = engine
	return nil
}

//Use 注册引用
func (sgo *ScriptGo) Use(symbols map[string]map[string]reflect.Value) {
	if sgo.symbols == nil {
		sgo.symbols = symbols
	}
	for k, v := range symbols {
		sgo.symbols[k] = v
	}
}

//Use 注册引用
func (sgo *ScriptGo) SetSourcecodeFilesystem(fs fs.FS) (err error) {
	if sgo._SourcecodeFilesystem != nil {
		err = errors.Errorf("SourcecodeFilesystem  has been set up")
		return err
	}
	sgo._SourcecodeFilesystem = fs
	return nil
}

//CallFuncScript 实际调用时的脚本语句
func (sgo *ScriptGo) CallFuncScript(funcName string, input string) (callScript string) {
	arr := strings.Split(funcName, ".")
	arr[len(arr)-1] = fmt.Sprintf("Call%s", arr[len(arr)-1]) // callfunc.tpl go 模板 生成调用函数时有前缀Call
	realFuncName := strings.Join(arr, ".")
	callScript = fmt.Sprintf("%s(`%s`)", realFuncName, input)
	return callScript
}

func (sgo *ScriptGo) Run(script string) (out string, err error) {
	rv, err := sgo.GetSymbolFromScript(script, nil)
	if err != nil {
		return "", err
	}
	outputDTOT := reflect.TypeOf(&OutputDTO{})
	if rv.CanConvert(outputDTOT) {
		dtoV := rv.Convert(outputDTOT)
		dto := dtoV.Interface().(*OutputDTO)
		out, err = dto.Data, dto.Err
		if err != nil {
			return "", err
		}
		return out, nil
	}
	out = rv.String()
	return out, nil
}

var (
	ERROR_GET_SYMBOL_SELECTOR_UNDEFINED = errors.New("undefined selector: ") // 这个文本是固定的，从yaegi/interp 包内拷贝，此处定义成错误，方便调用方判断错误类型
	ERROR_GET_SYMBOL_CAN_NOT_CONVERT    = errors.New("can not convert: ")
)

// GetSymbolFromScript 从动态脚本中获取特定符号(对象、函数、变量等)
func (sgo *ScriptGo) GetSymbolFromScript(selector string, dstType reflect.Type) (destSymbol reflect.Value, err error) {
	if sgo.engine == nil {
		err = sgo.Compile()
		if err != nil {
			return destSymbol, err
		}
	}
	destSymbol, err = sgo.engine.Eval(selector)
	if err != nil && strings.Contains(err.Error(), ERROR_GET_SYMBOL_SELECTOR_UNDEFINED.Error()) { // 不存在当前元素 时 忽略错误，程序容许只动态实现一部分
		err = nil
		return destSymbol, ERROR_GET_SYMBOL_SELECTOR_UNDEFINED
	}

	if err != nil {
		err = errors.WithMessage(err, selector)
		return destSymbol, err
	}
	if dstType == nil {
		return destSymbol, nil
	}
	if !destSymbol.CanConvert(dstType) {
		err = errors.Errorf("dynamic func %s ,must can convert to %s", selector, fmt.Sprintf("%s.%s", dstType.PkgPath(), dstType.Name()))
		return destSymbol, err
	}
	destSymbol = destSymbol.Convert(dstType)
	return destSymbol, nil
}

func NewScriptGo() (sgo *ScriptGo) {
	return &ScriptGo{}
}

var Symbols = stdlib.Symbols

//go:generate go install github.com/traefik/yaegi/cmd/yaegi
//go:generate yaegi extract github.com/tidwall/gjson
//go:generate yaegi extract github.com/tidwall/sjson
//go:generate yaegi extract github.com/spf13/cast
//go:generate yaegi extract github.com/syyongx/php2go
//go:generate yaegi extract github.com/suifengpiao14/goscript/yaegi/funcs
