package core

// PaymentSource describes the funding source for a tier renewal on-chain.
type PaymentSource struct {
    Source    string  `json:"source"`       // "vault", "hot", "stake"
    AmountGRC float64 `json:"amount_grc"`   // amount debited from the sender in GRC
}

// TxTierRenew is the canonical on-chain representation of a tier renewal.
//
// It is referenced by the HTTP layer (net package) and by the Chain engine.
type TxTierRenew struct {
    Sender           string        `json:"sender"`
    Nonce            uint64        `json:"nonce"`
    Tier             string        `json:"tier"`
    BillingCycle     string        `json:"billing_cycle"`
    Payment          PaymentSource `json:"payment"`
    EarnAppliedGRC   float64       `json:"earn_applied_grc"`
    StakeDiscountGRC float64       `json:"stake_discount_grc"`
    SurplusToTimeGRC float64       `json:"surplus_to_time_grc"`
    Timestamp        int64         `json:"timestamp"`
    Signature        string        `json:"signature"`
}
