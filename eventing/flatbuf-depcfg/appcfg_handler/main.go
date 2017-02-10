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

	auth := builder.CreateString("Admin:asdasd")
	metaBucket := builder.CreateString("eventing")
	sourceBucket := builder.CreateString("default")

	appcfg.DepCfgStart(builder)
	appcfg.DepCfgAddBuckets(builder, buckets)
	appcfg.DepCfgAddTickDuration(builder, 5000)
	appcfg.DepCfgAddWorkerCount(builder, 3)
	appcfg.DepCfgAddAuth(builder, auth)
	appcfg.DepCfgAddMetadataBucket(builder, metaBucket)
	appcfg.DepCfgAddSourceBucket(builder, sourceBucket)
	depcfg := appcfg.DepCfgEnd(builder)

	appCode := builder.CreateString("func OnUpdate() {log(meta.cas)}")
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

	fmt.Println("code: ", string(c.AppCode()), " name: ", string(c.AppName()), c.Id())

	d := new(appcfg.DepCfg)
	dd := c.DepCfg(d)
	fmt.Println("auth:", string(dd.Auth()), " src:", string(dd.SourceBucket()), " meta:", string(dd.MetadataBucket()))

	b := new(appcfg.Bucket)
	for i := 0; i < dd.BucketsLength(); i++ {
		if dd.Buckets(b, i) {
			fmt.Println("name:", string(b.BucketName()), " alias:", string(b.Alias()))
		}
	}
}
