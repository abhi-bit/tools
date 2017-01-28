package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/looplab/fsm"
)

type vbucketState int

type dcpStreamStatus int

const (
	Init vbucketState = iota
	DcpStreamRequested
	DcpStreamCreated
	SteadyState
	DcpStreamStopRequested
	DcpStreamStopped
)

const (
	InitVbucket dcpStreamStatus = iota
	StreamRequested
	StreamCreated
	StreamRunning
	StreamStopRequested
	StreamStopped
)

const (
	NumVbuckets = 2
)

type vbucketStateRepr struct {
	CurrentVbOwner  string
	DcpStreamStatus string
	History         []string
}

type vbFsm struct {
	fsm     *fsm.FSM
	history []string
}

var vbFsms map[int]*vbFsm

func main() {

	initVbFsms()

	for vbNo := 0; vbNo < NumVbuckets; vbNo++ {
		f := vbFsms[vbNo]

		vbRepr := readBlob(vbNo)

		log.Printf("vbNo: %d state: %s\n", vbNo, f.fsm.Current())
		triggerEvent(f, "streamRequest", vbRepr, vbNo)

		log.Printf("vbNo: %d state: %s\n", vbNo, f.fsm.Current())
		triggerEvent(f, "streamResponse", vbRepr, vbNo)

		log.Printf("vbNo: %d state: %s\n", vbNo, f.fsm.Current())
		triggerEvent(f, "updateMetaData", vbRepr, vbNo)

		log.Printf("vbNo: %d state: %s\n", vbNo, f.fsm.Current())
		triggerEvent(f, "giveUpOwnership", vbRepr, vbNo)

		log.Printf("vbNo: %d state: %s\n", vbNo, f.fsm.Current())
		triggerEvent(f, "streamEndReceived", vbRepr, vbNo)

		log.Printf("vbNo: %d state: %s\n", vbNo, f.fsm.Current())
		triggerEvent(f, "reinit", vbRepr, vbNo)

		log.Printf("vbNo: %d state: %s\n", vbNo, f.fsm.Current())
		triggerEvent(f, "streamRequest", vbRepr, vbNo)

		log.Printf("vbNo: %d state: %s\n", vbNo, f.fsm.Current())
		triggerEvent(f, "streamResponse", vbRepr, vbNo)
	}
}

func initVbFsms() {
	vbFsms = make(map[int]*vbFsm)

	for i := 0; i < NumVbuckets; i++ {

		f := &vbFsm{}
		f.history = make([]string, 0)
		f.history = append(f.history, Init.String())

		vbRepr := &vbucketStateRepr{
			CurrentVbOwner:  "",
			DcpStreamStatus: "",
			History:         f.history,
		}
		writeBlob(vbRepr, i)

		f.fsm = fsm.NewFSM(
			Init.String(),
			fsm.Events{
				{Name: "streamRequest", Src: []string{Init.String()}, Dst: DcpStreamRequested.String()},
				{Name: "streamResponse", Src: []string{DcpStreamRequested.String()}, Dst: DcpStreamCreated.String()},
				{Name: "updateMetaData", Src: []string{DcpStreamCreated.String()}, Dst: SteadyState.String()},
				{Name: "updateMetaData", Src: []string{SteadyState.String()}, Dst: SteadyState.String()},
				{Name: "giveUpOwnership", Src: []string{SteadyState.String()}, Dst: DcpStreamStopRequested.String()},
				{Name: "streamEndReceived", Src: []string{DcpStreamStopRequested.String()}, Dst: DcpStreamStopped.String()},
				{Name: "reinit", Src: []string{DcpStreamStopped.String()}, Dst: Init.String()},
			},
			fsm.Callbacks{
				"enter_Init": func(e *fsm.Event) {
					writeBlobHelper(e, InitVbucket)
				},
				"enter_DcpStreamRequested": func(e *fsm.Event) {
					writeBlobHelper(e, StreamRequested)
				},
				"enter_DcpStreamCreated": func(e *fsm.Event) {
					writeBlobHelper(e, StreamCreated)
				},
				"enter_SteadyState": func(e *fsm.Event) {
					writeBlobHelper(e, StreamRunning)
				},
				"enter_DcpStreamStopRequested": func(e *fsm.Event) {
					writeBlobHelper(e, StreamStopRequested)
				},
				"enter_DcpStreamStopped": func(e *fsm.Event) {
					writeBlobHelper(e, StreamStopped)
				},
			},
		)
		vbFsms[i] = f

	}
}

func writeBlobHelper(e *fsm.Event, status dcpStreamStatus) {
	vbRepr := e.Args[0].(*vbucketStateRepr)
	vbNo := e.Args[1].(int)

	vbRepr.DcpStreamStatus = status.String()
	vbRepr.CurrentVbOwner = fmt.Sprintf("%d", vbNo)
	vbRepr.History = append(vbRepr.History, e.Event)

	writeBlob(vbRepr, vbNo)
}

func triggerEvent(f *vbFsm, event string, arg1 interface{}, arg2 interface{}) error {
	err := f.fsm.Event(event, arg1, arg2)
	if err != nil {
		log.Printf("Err: %v\n", err)
	} else {
		f.history = append(f.history, event)
	}
	return err
}

func writeBlob(vbRepr *vbucketStateRepr, vbNo int) {
	encodedVb, err := json.Marshal(vbRepr)
	if err != nil {
		log.Println(err)
		return
	}

	fileName := fmt.Sprintf("%d.json", vbNo)
	err = ioutil.WriteFile(fileName, encodedVb, 0644)
	if err != nil {
		log.Println(err)
		return
	}
}

func readBlob(vbNo int) *vbucketStateRepr {
	fileName := fmt.Sprintf("%d.json", vbNo)

	encodedVb, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println(err)
		return nil
	}

	var vbRepr vbucketStateRepr
	err = json.Unmarshal(encodedVb, &vbRepr)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &vbRepr
}

func checkVbucketState(s *vbucketStateRepr) vbucketState {
	if s.CurrentVbOwner == "" && (s.DcpStreamStatus == "" || s.DcpStreamStatus == StreamStopped.String()) {
		return Init
	} else if s.CurrentVbOwner != "" && s.DcpStreamStatus == StreamRequested.String() {
		return DcpStreamRequested
	} else if s.CurrentVbOwner != "" && s.DcpStreamStatus == StreamCreated.String() {
		return DcpStreamCreated
	} else if s.CurrentVbOwner != "" && s.DcpStreamStatus == StreamRunning.String() {
		return SteadyState
	} else if s.CurrentVbOwner != "" && s.DcpStreamStatus == StreamStopRequested.String() {
		return DcpStreamStopRequested
	} else if s.CurrentVbOwner != "" && s.DcpStreamStatus == StreamStopped.String() {
		return DcpStreamStopped
	} else {
		return 0
	}
}

func (s dcpStreamStatus) String() string {
	switch s {
	case StreamRequested:
		return "StreamRequested"
	case StreamCreated:
		return "StreamCreated"
	case StreamRunning:
		return "StreamRunning"
	case StreamStopRequested:
		return "StreamStopRequested"
	case StreamStopped:
		return "StreamStopped"
	default:
		return ""
	}
}

func (s vbucketState) String() string {
	switch s {
	case Init:
		return "Init"
	case DcpStreamRequested:
		return "DcpStreamRequested"
	case DcpStreamCreated:
		return "DcpStreamCreated"
	case SteadyState:
		return "SteadyState"
	case DcpStreamStopRequested:
		return "DcpStreamStopRequested"
	case DcpStreamStopped:
		return "DcpStreamStopped"
	default:
		return ""
	}
}
