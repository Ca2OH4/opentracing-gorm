# opentracing gorm

[OpenTracing](http://opentracing.io/) instrumentation for [GORM](http://gorm.io/).

## Install

```
go get -u github.com/smacker/opentracing-gorm
```

## Usage

1. _ = db.Use(&otgorm.Plugin{})
   2.span := opentracing.StartSpan("gormTracing unit test")
   3.ctx := opentracing.ContextWithSpan(context.Background(), span)
   4.session := db.WithContext(ctx)

Example:

```go
func TestOTGorm(t *testing.T) {
_, closer, err := newJaegerTracer(
"test-OTGorm", "127.0.0.1:6831",
)
defer closer.Close()
if err != nil {
t.Fatal(err)
}
db := initDB()
// 生成新的Span - 注意将span结束掉，不然无法发送对应的结果
span := opentracing.StartSpan("gormTracing unit test")
defer span.Finish()

// 把生成的Root Span写入到Context上下文，获取一个子Context
// 通常在Web项目中，Root Span由中间件生成
ctx := opentracing.ContextWithSpan(context.Background(), span)

// 将上下文传入DB实例，生成Session会话
// 这样子就能把这个会话的全部信息反馈给Jaeger
session := db.WithContext(ctx)

// Create
session.Create(&Product{Code: "D42", Price: 100})

// Read
var product Product
session.First(&product, 1) // 根据整形主键查找
session.First(&product, "code = ?", "D42") // 查找 code 字段值为 D42 的记录

// Update - 将 product 的 price 更新为 200
session.Model(&product).Update("Price", 200)
// Update - 更新多个字段
session.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // 仅更新非零值字段
session.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

// Delete - 删除 product
session.Delete(&product, 1)
}           
```

Call to the `Handler` function would create sql span with table name, sql method and sql statement as a child of handler
span.

## License

[MIT](LICENSE)
