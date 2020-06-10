package visual

import (
    "fmt"
    "go_protoc"
    "log"
    "testing"
    "time"

    "golang.org/x/net/context"
    "google.golang.org/grpc"
)

func TestVisualize(t *testing.T) {
    conn, err := grpc.Dial("localhost:6008", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := go_protoc.NewChordUpdateClient(conn)
    gap := 6
    for i := 0; i < 10; i++ {
        reply, err := client.Update(context.Background(), &go_protoc.ChordMessage{
            Nid:   uint64(i * gap),
            Verb:  go_protoc.ChordMessage_NEW_NODE,
            Value: "",
        })
        if err != nil {
            log.Fatal(err)
        }
        fmt.Println(reply.GetReply())

        <-time.After(500 * time.Millisecond)
    }

    for i := 0; i < 10; i++ {
        reply, err := client.Update(context.Background(), &go_protoc.ChordMessage{
            Nid:   uint64(i * gap),
            Verb:  go_protoc.ChordMessage_SET_SUCC,
            Value: fmt.Sprintf("%v", uint64((i*gap+gap)%(10*gap))),
        })
        if err != nil {
            log.Fatal(err)
        }
        fmt.Println(reply.GetReply())

        <-time.After(500 * time.Millisecond)
    }

    for i := 0; i < 10; i++ {
        reply, err := client.Update(context.Background(), &go_protoc.ChordMessage{
            Nid:   uint64(i * gap),
            Verb:  go_protoc.ChordMessage_SET_PRED,
            Value: fmt.Sprintf("%v", uint64((10*gap+i*gap-gap)%(10*gap))),
        })
        if err != nil {
            log.Fatal(err)
        }
        fmt.Println(reply.GetReply())

        <-time.After(500 * time.Millisecond)
    }

    for i := 0; i < 10; i++ {
       fmt.Print("set highlight!")
       reply, err := client.Update(context.Background(), &go_protoc.ChordMessage{
           Nid:   uint64(i * gap),
           Verb:  go_protoc.ChordMessage_SET_HLGHT,
           Value: "1",
       })
       if err != nil {
           log.Fatal(err)
       }
       fmt.Println(reply.GetReply())

       <-time.After(500 * time.Millisecond)
    }

    for i := 0; i < 10; i++ {
       reply, err := client.Update(context.Background(), &go_protoc.ChordMessage{
           Nid:   uint64(i * gap),
           Verb:  go_protoc.ChordMessage_SET_HLGHT,
           Value: "0",
       })
       if err != nil {
           log.Fatal(err)
       }
       fmt.Println(reply.GetReply())

       <-time.After(500 * time.Millisecond)
    }
}
