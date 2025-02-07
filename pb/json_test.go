package pb_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/proto/proto3_proto"
	"github.com/izumin5210/hx"
	"github.com/izumin5210/hx/pb"
)

func TestJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/echo":
			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var (
				msg proto3_proto.Message
				buf bytes.Buffer
			)

			err := (&jsonpb.Unmarshaler{}).Unmarshal(r.Body, &msg)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = (&jsonpb.Marshaler{}).Marshal(&buf, &msg)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Write(buf.Bytes())

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	want := &proto3_proto.Message{
		Name:     "It, Works!",
		Score:    120,
		Hilarity: proto3_proto.Message_SLAPSTICK,
		Children: []*proto3_proto.Message{
			{Name: "foo", HeightInCm: 170},
			{Name: "bar", TrueScotsman: true},
		},
	}

	t.Run("simple", func(t *testing.T) {
		var got proto3_proto.Message
		err := hx.Post(context.Background(), ts.URL+"/echo",
			pb.JSON(want),
			hx.WhenSuccess(pb.AsJSON(&got)),
			hx.WhenFailure(hx.AsError()),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
		assertProtoMessage(t, want, &got)
	})

	t.Run("custom encoder", func(t *testing.T) {
		var got, overwrited proto3_proto.Message
		overwrited = *want
		overwrited.Name = "It, Works!!!!!!!!!!!!!!!!!!!!!!"

		jsonCfg := &pb.JSONConfig{
			EncodeFunc: func(_ proto.Message) (io.Reader, error) {
				var buf bytes.Buffer
				err := (&jsonpb.Marshaler{}).Marshal(&buf, &overwrited)
				if err != nil {
					return nil, err
				}
				return &buf, nil
			},
		}
		err := hx.Post(context.Background(), ts.URL+"/echo",
			jsonCfg.JSON(want),
			hx.WhenSuccess(pb.AsJSON(&got)),
			hx.WhenFailure(hx.AsError()),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
		assertProtoMessage(t, &overwrited, &got)
	})

	t.Run("custom decoder", func(t *testing.T) {
		var got, overwrited proto3_proto.Message
		overwrited = *want
		overwrited.Name = "It, Works!!!!!!!!!!!!!!!!!!!!!!"

		jsonCfg := &pb.JSONConfig{
			DecodeFunc: func(r io.Reader, m proto.Message) error {
				(*m.(*proto3_proto.Message)) = *want
				return nil
			},
		}
		err := hx.Post(context.Background(), ts.URL+"/echo",
			pb.JSON(&overwrited),
			hx.WhenSuccess(jsonCfg.AsJSON(&got)),
			hx.WhenFailure(hx.AsError()),
		)
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
		assertProtoMessage(t, want, &got)
	})
}
