package nats

import (
	"errors"
	"fmt"
	"testing"

	"github.com/nats-io/nats.go"
)

func TestNewNats(t *testing.T) {
	conn, err := nats.Connect("nats://localhost:4222", nats.UserInfo("ruser", "T0pS3cr3t"))
	if err != nil {
		t.Fatal(err)
	}

	js, err := conn.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	//js.AddStream(&nats.StreamConfig{
	//	MaxAge: 86400,
	//	MaxMsgsPerSubject: 1,
	//	Retention:
	//})
	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: "test",
	})

	if err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		t.Fatal(err)
	}

	rev, err := kv.Create("install", []byte{})
	if err != nil {
		t.Fatal(err)
	}

	v, err := kv.Get("install")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(v.Revision())
	//kv.Create()
	rev, err = kv.Update("install", []byte{}, 0)
	if err != nil {
		t.Fatal(err)
	}
	list, err := kv.ListKeys()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(rev)
	for k := range list.Keys() {
		fmt.Println(k)
	}

}
