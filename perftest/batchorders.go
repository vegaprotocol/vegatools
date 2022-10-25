package perftest

import commandspb "code.vegaprotocol.io/vega/protos/vega/commands/v1"

type BatchOrders struct {
	cancels []*commandspb.OrderCancellation
	amends  []*commandspb.OrderAmendment
	orders  []*commandspb.OrderSubmission
}

func (b *BatchOrders) GetMessageCount() int {
	return len(b.cancels) + len(b.amends) + len(b.orders)
}

func (b *BatchOrders) Empty() {
	b.cancels = b.cancels[:0]
	b.amends = b.amends[:0]
	b.orders = b.orders[:0]
}
