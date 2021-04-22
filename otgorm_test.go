package otgorm_test

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	otgorm "opentracing-gorm"
	"testing"
	"time"
)

func newJaegerTracer(serverName, agentHostPort string) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		ServiceName: serverName,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  agentHostPort,
		},
	}
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, nil, err
	}
	// 设置全局 tracer
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer, err
}

type Product struct {
	gorm.Model
	Code  string
	Price float64
}

func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	_ = db.Use(&otgorm.Plugin{})
	_ = db.AutoMigrate(&Product{})
	db.Create(&Product{Code: "L1212"})
	return db
}

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
	session.First(&product, 1)                 // 根据整形主键查找
	session.First(&product, "code = ?", "D42") // 查找 code 字段值为 D42 的记录

	// Update - 将 product 的 price 更新为 200
	session.Model(&product).Update("Price", 200)
	// Update - 更新多个字段
	session.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // 仅更新非零值字段
	session.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - 删除 product
	session.Delete(&product, 1)
}
