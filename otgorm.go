package otgorm

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gorm.io/gorm"
	"strings"
)

const (
	parentSpanGormKey = "opentracingParentSpan"
	spanGormKey       = "opentracingSpan"

	callBackBeforeName = "opentracing:before"
	callBackAfterName  = "opentracing:after"
)

// 生成子 span
func before(db *gorm.DB) {
	// 先从父级 span 生成子 span
	span, _ := opentracing.StartSpanFromContext(db.Statement.Context, parentSpanGormKey)

	// 利用 db 实例去传递 span
	db.InstanceSet(spanGormKey, span)
	return
}

// 获取追踪数据
func after(db *gorm.DB) {
	// 从 db 实例中取出 span
	val, ok := db.InstanceGet(spanGormKey)
	if !ok {
		return
	}
	sp := val.(opentracing.Span)
	method := strings.ToUpper(strings.Split(db.Statement.SQL.String(), " ")[0])
	if db.Error != nil || db.Statement.Error != nil {
		ext.Error.Set(sp, true)
	} else {
		ext.Error.Set(sp, false)
	}
	ext.DBStatement.Set(sp, db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...))
	sp.SetTag("db.table", db.Statement.Table)
	sp.SetTag("db.method", method)
	sp.SetTag("db.count", db.Statement.RowsAffected)
	sp.SetTag("db.err", db.Error)
	sp.SetTag("db.statement.err", db.Statement.Error)

	sp.Finish()
}

type Plugin struct{}

var _ gorm.Plugin = &Plugin{}

func (p *Plugin) Name() string {
	return "opentracingPlugin"
}

func (p *Plugin) Initialize(db *gorm.DB) (err error) {
	db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, before)
	db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, before)
	db.Callback().Delete().Before("gorm:before_delete").Register(callBackBeforeName, before)
	db.Callback().Update().Before("gorm:setup_reflect_value").Register(callBackBeforeName, before)
	db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, before)
	db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, before)

	// 结束后
	db.Callback().Create().After("gorm:after_create").Register(callBackAfterName, after)
	db.Callback().Query().After("gorm:after_query").Register(callBackAfterName, after)
	db.Callback().Delete().After("gorm:after_delete").Register(callBackAfterName, after)
	db.Callback().Update().After("gorm:after_update").Register(callBackAfterName, after)
	db.Callback().Row().After("gorm:row").Register(callBackAfterName, after)
	db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, after)
	return
}
