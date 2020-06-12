package visual

import (
	"fmt"
	"go_protoc"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestVisualize(t *testing.T) {
	conn, err := grpc.Dial("localhost:6008", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := go_protoc.NewChordUpdateClient(conn)
	gap := 6
	for i := 0; i < 10; i++ {
		_, err := client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_NEW_NODE,
			Value: "",
		})
		if err != nil {
			t.Fatal(err)
		}

		<-time.After(500 * time.Millisecond)
	}

	for i := 0; i < 10; i++ {
		_, err := client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_SET_SUCC,
			Value: fmt.Sprintf("%v", uint64((i*gap+gap)%(10*gap))),
		})
		if err != nil {
			t.Fatal(err)
		}

		<-time.After(500 * time.Millisecond)
	}

	for i := 0; i < 10; i++ {
		_, err := client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_SET_PRED,
			Value: fmt.Sprintf("%v", uint64((10*gap+i*gap-gap)%(10*gap))),
		})
		if err != nil {
			t.Fatal(err)
		}

		<-time.After(500 * time.Millisecond)
	}

	for i := 0; i < 10; i++ {
		_, err := client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_SET_HLGHT,
			Value: "1",
		})
		if err != nil {
			t.Fatal(err)
		}

		<-time.After(500 * time.Millisecond)
	}

	for i := 0; i < 10; i++ {
		_, err := client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_SET_HLGHT,
			Value: "0",
		})
		if err != nil {
			t.Fatal(err)
		}

		<-time.After(500 * time.Millisecond)
	}

	for i := 0; i < 10; i++ {
		_, err := client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_SET_KEY,
			Value: fmt.Sprintf("%v", i*gap),
			Key:   fmt.Sprintf("%v", i*gap),
		})
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(500 * time.Millisecond)

		_, err = client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_SET_KEY,
			Value: fmt.Sprintf("%v", i*gap+1),
			Key:   fmt.Sprintf("%v", i*gap),
		})
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(500 * time.Millisecond)

		_, err = client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_SET_KEY,
			Value: fmt.Sprintf("%v", i*gap),
			Key:   fmt.Sprintf("%v", i*gap+2),
		})
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(500 * time.Millisecond)

		_, err = client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_SET_KEY,
			Value: fmt.Sprintf("%v", i*gap),
			Key:   fmt.Sprintf("%v", i*gap+3),
		})
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(500 * time.Millisecond)
	}

	for i := 0; i < 10; i++ {
		_, err := client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_DEAD_NODE,
			Value: fmt.Sprintf("%v", i*gap),
		})
		if err != nil {
			t.Fatal(err)
		}

		<-time.After(500 * time.Millisecond)
	}

	// double delete should not return error
	for i := 0; i < 10; i++ {
		_, err := client.Update(context.Background(), &go_protoc.ChordMessage{
			Nid:   uint64(i * gap),
			Verb:  go_protoc.ChordMessage_DEAD_NODE,
			Value: fmt.Sprintf("%v", i*gap),
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
