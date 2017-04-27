package main

import (
	"fmt"

	"github.com/couchbase/gocb"
)

func main() {
	c, err := gocb.Connect("couchbase://localhost:12000")
	if err != nil {
		fmt.Printf("connect err: %v\n", err)
		return
	}
	b, err := c.OpenBucket("default", "asdasd")
	if err != nil {
		fmt.Printf("bucket open err: %v\n", err)
		return
	}

	timers := make([]string, 0)
	timers = append(timers, "12:00:00")

	docF := b.MutateIn("abc0", 0, uint32(0))
	docF.UpsertEx("_eventing.timer", timers, gocb.SubdocFlagXattr|gocb.SubdocFlagCreatePath)
	docF.UpsertEx("_eventing.cas", "${Mutation.CAS}", gocb.SubdocFlagXattr|gocb.SubdocFlagCreatePath|gocb.SubdocFlagUseMacros)
	_, err = docF.Execute()
	fmt.Printf("err: %v\n", err)
}
