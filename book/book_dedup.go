package book

// Remove transactions from book that already exists in the main book.
//
// This matches only (date, payee) unique combinations and doesn't inspect
// the actual postings within the transaction.
func (b *Book) RemoveDuplicatesOf(main *Book) {
	reftrans := main.Transactions()
	refidx := 0
	b.FilterTransaction(func(date Date, payee string, posts Transaction) bool {

		// Move forward as much as possible
		for refidx < len(reftrans) {
			if reftrans[refidx][0].date < date {
				refidx++
				continue
			}
			if reftrans[refidx][0].date == date && reftrans[refidx][0].payee < payee {
				refidx++
				continue
			}
			break
		}

		// Skip if at end
		if refidx == len(reftrans) {
			return true
		}

		// If the same, then we have a match. Remove this one.
		if reftrans[refidx][0].payee == payee && reftrans[refidx][0].date == date {
			return false
		}

		// Otherwise keep it.
		return true
	})
}
