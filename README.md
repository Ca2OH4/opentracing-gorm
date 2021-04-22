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
var gDB *gorm.DB

func init() {
    gDB = initDB()
}

func initDB() *gorm.DB {
    db, err := gorm.Open("sqlite3", ":memory:")
    if err != nil {
        panic(err)
    }
    // register callbacks must be called for a root instance of your gorm.DB
    otgorm.AddGormCallbacks(db)
    return db
}

func Handler(ctx context.Context) {
    span, ctx := opentracing.StartSpanFromContext(ctx, "handler")
    defer span.Finish()

    // clone db with proper context
    db := otgorm.SetSpanToGorm(ctx, gDB)

    // sql query
    db.First
}
```

Call to the `Handler` function would create sql span with table name, sql method and sql statement as a child of handler span.

## License

[MIT](LICENSE)
