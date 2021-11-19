package discordgo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEpochMsTime(t *testing.T) {
	s := struct {
		CreatedAt EpochMsTime `json:"created_at,omitempty"`
	}{
		CreatedAt: EpochMsTime(time.Date(2021, 11, 19, 6, 0, 0, 0, time.UTC)),
	}

	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("error in marshal: %v", err)
	}
	actual := string(b)
	expected := `{"created_at":1637301600000}`
	if actual != expected {
		t.Fatalf("EpochMsTime did not marshal correctly:\n\tExpected: %v\n\tActual:%v", expected, actual)
	}

	sprime := struct {
		CreatedAt EpochMsTime `json:"created_at,omitempty"`
	}{}

	if err := json.Unmarshal(b, &sprime); err != nil {
		t.Fatalf("EpochMsTime ")
	}

	if !s.CreatedAt.Equal(sprime.CreatedAt) {
		t.Fatalf("Struct with EpochMsTime did not survive marshal/unmarshal:\n\tBefore: %v\n\tAfter:%v", s, sprime)
	}
}
