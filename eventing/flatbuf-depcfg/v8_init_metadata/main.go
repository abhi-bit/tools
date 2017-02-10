package main

import (
	"fmt"
	"io/ioutil"

	"github.com/abhi-bit/tools/eventing/flatbuf-depcfg/v8init"
	flatbuffers "github.com/google/flatbuffers/go"
)

func main() {
	builder := flatbuffers.NewBuilder(0)

	appName := builder.CreateString("credit_score")
	kvHostPort := builder.CreateString("localhost:12000")
	depCfg := builder.CreateString("{\"tickDuration\": 5000, \"workerCount\": 3}")

	v8init.InitStart(builder)
	v8init.InitAddAppname(builder, appName)
	v8init.InitAddKvhostport(builder, kvHostPort)
	v8init.InitAddDepcfg(builder, depCfg)
	initMsg := v8init.InitEnd(builder)

	builder.Finish(initMsg)
	buf := builder.FinishedBytes()

	err := ioutil.WriteFile("v8_init_payload", buf, 0644)
	if err != nil {
		fmt.Printf("Error while writing encoded data to disk, err: %v", err)
		return
	}
}
