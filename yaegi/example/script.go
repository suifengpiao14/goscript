package vocabulary

import (
	"errors"

	"github.com/spf13/cast"
	"github.com/suifengpiao14/goscript/yaegi"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func SetLimit(index int, size int) (offset int, limit int) {
	limit = size
	offset = index * size
	return offset, limit
}

func CallSetLimit(input string) (outputDTO *yaegi.OutputDTO) {
	index := cast.ToInt(gjson.Get(input, "func.vocabulary.SetLimit.input.index").String())
	size := cast.ToInt(gjson.Get(input, "func.vocabulary.SetLimit.input.size").String())
	var out string
	var err error
	outputDTO = new(yaegi.OutputDTO)
	{ // 避免局部变量冲突
		offset, size := SetLimit(index, size)
		out, err = sjson.Set(out, "func.vocabulary.SetLimit.output.offset", offset)
		if err != nil {
			outputDTO.Err = err
			return outputDTO
		}
		out, err = sjson.Set(out, "func.vocabulary.SetLimit.output.size", size)
		if err != nil {
			outputDTO.Err = err
			return outputDTO
		}

	}
	outputDTO = &yaegi.OutputDTO{
		Data: out,
		Err:  errors.New("hhah"),
	}
	return outputDTO
}
