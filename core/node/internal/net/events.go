package net

import "time"

type EventType string

const (
    EventNewBlock       EventType = "NewBlock"
    EventTransfer       EventType = "Transfer"
    EventMint           EventType = "Mint"
    EventRedeem         EventType = "Redeem"
    EventValuationTick  EventType = "ValuationTick"
    EventWindowUpdate   EventType = "WindowUpdate"
    EventTreasuryUpdate EventType = "TreasuryUpdate"
    EventNodeStatus     EventType = "NodeStatus"
)

type Event struct {
    ID        string      `json:"id"`
    Type      EventType   `json:"type"`
    Version   string      `json:"version"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}
