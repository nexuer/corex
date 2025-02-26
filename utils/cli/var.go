package cli

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

type DurationPBValue durationpb.Duration

func (d *DurationPBValue) String() string {
	if d.Seconds == 0 && d.Nanos == 0 {
		return ""
	}
	return (*durationpb.Duration)(d).AsDuration().String()
}

func (d *DurationPBValue) Set(value string) error {
	v, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	*d = DurationPBValue(*durationpb.New(v))
	return err
}

func (d *DurationPBValue) Type() string {
	return "duration"
}

func NewDurationPBValue(val *durationpb.Duration, p *durationpb.Duration) *DurationPBValue {
	if val == nil {
		return (*DurationPBValue)(p)
	}
	*p = *durationpb.New(val.AsDuration())
	return (*DurationPBValue)(p)
}
