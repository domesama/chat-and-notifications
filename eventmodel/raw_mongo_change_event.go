package eventmodel

import "github.com/goccy/go-json"

// These are the change event types & output format from MongoDB Debezium Change Streams

type ChangeEventType string

const (
	EventTypeCreate ChangeEventType = "c"

	EventTypeUpdate ChangeEventType = "u"

	EventTypeDelete ChangeEventType = "d"
)

type RawMongoChangePayload struct {
	Before      *string         `json:"before"`
	After       *string         `json:"after"`
	EventType   ChangeEventType `json:"op"`
	TimeStampMS int64           `json:"ts_ms"`
	// In case we need more info from the following change in future
	// UpdateDescription any     `json:"updateDescription"`
	// Transaction       any     `json:"transaction"`
	// Source      `json:"source"`
}

type Source struct {
	Version               string `json:"version"`
	Connector             string `json:"connector"`
	Name                  string `json:"name"`
	Snapshot              string `json:"snapshot"`
	DB                    string `json:"db"`
	Sequence              any    `json:"sequence"`
	TimestampMilliSeconds int64  `json:"ts_ms"`
	TimestampMicroSecond  int64  `json:"ts_us"`
	TimestampNanoSecond   int64  `json:"ts_ns"`
	Collection            string `json:"collection"`
	Ord                   int    `json:"ord"`
	LSID                  string `json:"lsid"`
	TxnNumber             int    `json:"txnNumber"`
	WallTime              any    `json:"wallTime"`
}

func (r RawMongoChangePayload) ToBytes() (b []byte, err error) {
	b, err = json.Marshal(r)
	return
}
