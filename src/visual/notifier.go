package visual

import (
    "go_protoc"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "log"
)

const (
    NEW_NODE = go_protoc.ChordMessage_NEW_NODE
    SET_SUCC = go_protoc.ChordMessage_SET_SUCC
    SET_PRED = go_protoc.ChordMessage_SET_PRED
    SET_HLGHT = go_protoc.ChordMessage_SET_HLGHT
    NEWS = go_protoc.ChordMessage_NEWS
)

type ChordMsg struct {
    Id uint64
    Verb go_protoc.ChordMessage_ChordVerb
    Value string
}

func SendMessage(addr string, msg ChordMsg) {
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := go_protoc.NewChordUpdateClient(conn)
    client.Update(context.Background(), &go_protoc.ChordMessage{
        Nid:   msg.Id,
        Verb:  msg.Verb,
        Value: msg.Value,
    })
}
