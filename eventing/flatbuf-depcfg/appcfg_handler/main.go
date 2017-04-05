package main

import (
	"fmt"
	"io/ioutil"

	"github.com/abhi-bit/tools/eventing/flatbuf-depcfg/appcfg"
	flatbuffers "github.com/google/flatbuffers/go"
)

func main() {
	builder := flatbuffers.NewBuilder(0)

	bucketName := builder.CreateString("default")
	bucketAlias := builder.CreateString("credit_score")

	appcfg.BucketStart(builder)
	appcfg.BucketAddBucketName(builder, bucketName)
	appcfg.BucketAddAlias(builder, bucketAlias)
	csBucket := appcfg.BucketEnd(builder)

	appcfg.DepCfgStartBucketsVector(builder, 1)
	builder.PrependUOffsetT(csBucket)
	buckets := builder.EndVector(1)

	metaBucket := builder.CreateString("eventing")
	sourceBucket := builder.CreateString("default")
	rbacpass := builder.CreateString("asdasd")
	rbacrole := builder.CreateString("admin")
	rbacuser := builder.CreateString("eventing")

	appcfg.DepCfgStart(builder)
	appcfg.DepCfgAddBuckets(builder, buckets)
	appcfg.DepCfgAddMetadataBucket(builder, metaBucket)
	appcfg.DepCfgAddSourceBucket(builder, sourceBucket)
	appcfg.DepCfgAddRbacpass(builder, rbacpass)
	appcfg.DepCfgAddRbacrole(builder, rbacrole)
	appcfg.DepCfgAddRbacuser(builder, rbacuser)
	depcfg := appcfg.DepCfgEnd(builder)

	appCode := builder.CreateString("func OnUpdate(doc, meta) {log(meta.cas)}")
	appName := builder.CreateString("credit_score")

	appcfg.ConfigStart(builder)
	appcfg.ConfigAddId(builder, 1)
	appcfg.ConfigAddAppName(builder, appName)
	appcfg.ConfigAddAppCode(builder, appCode)
	appcfg.ConfigAddDepCfg(builder, depcfg)
	config := appcfg.ConfigEnd(builder)

	builder.Finish(config)

	buf := builder.FinishedBytes()

	err := ioutil.WriteFile("depcfg", buf, 0644)
	if err != nil {
		fmt.Printf("Failed to write flatbuf encoded message to disk, err: %v\n", err)
		return
	}

	c := appcfg.GetRootAsConfig(buf, 0)

	fmt.Println("code:", string(c.AppCode()), " name:", string(c.AppName()), " appid:", c.Id())

	d := new(appcfg.DepCfg)
	dd := c.DepCfg(d)
	fmt.Println("src:", string(dd.SourceBucket()), " meta:", string(dd.MetadataBucket()),
		" rbacuser:", string(dd.Rbacuser()), " rbacpass:", string(dd.Rbacpass()),
		" rbacrole:", string(dd.Rbacrole()))

	b := new(appcfg.Bucket)
	for i := 0; i < dd.BucketsLength(); i++ {
		if dd.Buckets(b, i) {
			fmt.Println("name:", string(b.BucketName()), " alias:", string(b.Alias()))
		}
	}
}
