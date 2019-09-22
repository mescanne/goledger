package book

// Reclassify counter-accounts in two-posting transactions in the book.
//
// Where the refacct is used in two-posting transactions in the ref book
// the payee will be used to determine the counter account.
//
// The book will have its similar postings adjusted to have the same counter
// account based on the dist() function. The dist function will be used
// to identify the most desirable matching payee.
//
// If dist() returns 0.0, then there is no match and no re-classification.
//
func (b *Book) ReclassifyByAccount(ref *Book, refacct string, dist func(a, b string) float64) {

	// Find payee -> counter account mappings
	payee_mapping_all := make(map[string]map[string]int)
	for _, trans := range ref.Transactions() {

		// Ensure only two-postings
		if len(trans) != 2 {
			continue
		}

		// Get counter account or continue
		cacct := ""
		if trans[0].GetAccount() == refacct {
			cacct = trans[1].GetAccount()
		} else if trans[1].GetAccount() == refacct {
			cacct = trans[0].GetAccount()
		} else {
			continue
		}

		// Add to mapping
		daccts, ok := payee_mapping_all[trans.GetPayee()]
		if ok {
			cnt, ok := daccts[cacct]
			if !ok {
				daccts[cacct] = 1
			} else {
				daccts[cacct] = cnt + 1
			}
		} else {
			payee_mapping_all[trans.GetPayee()] = map[string]int{cacct: 1}
		}
	}

	// Choose the most common (or one of the most common) accounts
	payee_mapping := make(map[string]string)
	for payee, caccts := range payee_mapping_all {
		max := 0
		macct := ""
		for cacct, k := range caccts {
			if k > max {
				k = max
				macct = cacct
			}
		}
		payee_mapping[payee] = macct
	}

	for _, trans := range b.Transactions() {
		if len(trans) != 2 {
			continue
		}

		// Get counter account or continue
		ctrans := -1
		if trans[0].GetAccount() == refacct {
			ctrans = 1
		} else if trans[1].GetAccount() == refacct {
			ctrans = 0
		} else {
			continue
		}

		max_match := 0.0
		max_cacct := ""
		for p, c := range payee_mapping {
			d := dist(trans.GetPayee(), p)
			if d > max_match {
				max_match = d
				max_cacct = c
			}
		}

		if max_match > 0.0 {
			trans[ctrans].acct = max_cacct
		}
	}
}
