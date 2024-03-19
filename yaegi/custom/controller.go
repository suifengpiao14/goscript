package custom

import (
	"context"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/suifengpiao14/pathtransfer"
	"github.com/suifengpiao14/torm"
	"github.com/tidwall/gjson"
)

// ExecSouceFn 执行资源函数
type ExecSouceFn func(ctx context.Context, identify string, input []byte) (out []byte, err error)

type InjectObject struct {
	ExecSQLTPL    ExecSouceFn
	PathTransfers pathtransfer.Transfers
	Torms         torm.Torms
}

func (ijctO InjectObject) GetTorm(tormName string) (tor *torm.Torm, err error) {
	tor, err = ijctO.Torms.GetByName(tormName)
	return tor, err
}

type DynamicLogicFn func(ctx context.Context, injectObject InjectObject, input []byte) (out []byte, err error)

func NewPaginationLogicFn(listTormName string, totalTormName string) (logicFn DynamicLogicFn) {
	return func(ctx context.Context, injectObject InjectObject, input []byte) (out []byte, err error) {
		return Pagination(ctx, listTormName, totalTormName, injectObject, input)
	}
}

func Pagination(ctx context.Context, listTormName string, totalTormName string, injectObject InjectObject, input []byte) (out []byte, err error) {

	totalJson, err := injectObject.ExecSQLTPL(ctx, totalTormName, input)
	if err != nil {
		return nil, err
	}
	totalTorm, err := injectObject.GetTorm(totalTormName)
	if err != nil {
		return nil, err
	}
	_, outTransfers := totalTorm.Transfers.SplitInOut()
	totalPath := ""
	if len(outTransfers) > 0 {
		totalPath = outTransfers[0].Src.Path
	}
	if totalPath == "" {
		err = errors.Errorf("not foud total torm output path")
		return nil, err
	}
	total := cast.ToInt(gjson.GetBytes(totalJson, totalPath))
	if total == 0 {
		return nil, nil
	}
	paginationJson, err := injectObject.ExecSQLTPL(ctx, listTormName, input)
	if err != nil {
		return nil, err
	}
	out, err = jsonpatch.MergePatch([]byte(paginationJson), []byte(totalJson))
	if err != nil {
		return nil, err
	}
	return out, nil

}
