package perftest

import commandspb "code.vegaprotocol.io/vega/protos/vega/commands/v1"

// BatchOrders stores pending messages ready to be sent in a batch
type BatchOrders struct {
	cancels []*commandspb.OrderCancellation
	amends  []*commandspb.OrderAmendment
	orders  []*commandspb.OrderSubmission
}

// GetMessageCount returns the number of meesages waiting to be sent for this one user
func (b *BatchOrders) GetMessageCount() int {
	return len(b.cancels) + len(b.amends) + len(b.orders)
}

// Empty clears any pending messages
func (b *BatchOrders) Empty() {
	b.cancels = b.cancels[:0]
	b.amends = b.amends[:0]
	b.orders = b.orders[:0]
}
