package book

import (
	"math/big"
	"regexp"
	"sort"
	"strings"
)

func findMaxCommonPrefix(left string, right string, divider string) string {
	prefix := 0
	i := 0
	for ; i < len(left) && i < len(right); i++ {
		if left[i] != right[i] {
			break
		}
		if strings.HasPrefix(left[i:], divider) {
			prefix = i
		}
	}

	if (i < len(left) && strings.HasPrefix(left[i:], divider)) ||
		(i < len(right) && strings.HasPrefix(right[i:], divider)) ||
		(i == len(left) && i == len(right)) {
		prefix = i
	}

	return left[0:prefix]
}

func splitByStrings(prefixes map[string]bool, acct string, includeSelf bool, divider string) (int, string) {
	last := 0
	o := make([]string, 0, 5)
	i := 0
	for {
		if i == len(acct) {
			if includeSelf {
				if _, ok := prefixes[acct]; ok {
					o = append(o, acct[last:])
					last = i
				}
			}
			break
		}

		if strings.HasPrefix(acct[i:], divider) {
			if _, ok := prefixes[acct[:i]]; ok {
				o = append(o, acct[last:i])
				last = i + len(divider)
			}
		}

		i++
	}
	o = append(o, acct[last:])
	return len(o) - 1, acct[last:]
}

func splitByStringsOld(prefixes map[string]bool, acct string, includeSelf bool, divider string) []string {
	last := 0
	o := make([]string, 0, 5)
	i := 0
	for {
		if i == len(acct) {
			if includeSelf {
				if _, ok := prefixes[acct]; ok {
					o = append(o, acct[last:])
					last = i
				}
			}
			break
		}

		if strings.HasPrefix(acct[i:], divider) {
			if _, ok := prefixes[acct[:i]]; ok {
				o = append(o, acct[last:i])
				last = i + len(divider)
			}
		}

		i++
	}
	o = append(o, acct[last:])
	return o
}

func (b *Book) Accumulate(toCCY string, divider string, credit *regexp.Regexp, hidden string) []Transaction {

	// only need some of compact functionality
	b.compact()

	posts := b.post

	if len(posts) == 0 {
		return b.Transactions()
	}

	// Accumulated Postings.
	eposts := make([]Posting, 0, len(posts)) // Existing Terminal
	sposts := make([]Posting, 0, len(posts)) // Non-Terminal

	// This is for converted account matching.
	type PostKey struct {
		acct string
		val  *big.Rat
	}
	cposts := make([]PostKey, 0, len(posts))

	// Accumulated -- converted currency only
	accum := make(map[string]bool)

	transIdx := 0
	i := 0
	for {
		// If this is different than previous, accumulate
		if i == len(posts) || posts[i].date != posts[transIdx].date || posts[i].payee != posts[transIdx].payee {

			// Assign level to normal postings
			for j := transIdx; j < i; j++ {
				//eposts[j].acctlevels = splitByStringsOld(accum, eposts[j].acct, true, divider)
				eposts[j].acctlevel, eposts[j].acctterm = splitByStrings(accum, eposts[j].acct, true, divider)
			}
			// Calculate new postings
			for k, _ := range accum {

				// Find the level
				//newparts := splitByStringsOld(accum, k, false, divider)
				newlevels, newterm := splitByStrings(accum, k, false, divider)

				// Accumulate across postings
				lkey := k + divider
				v := big.NewRat(0, 1)
				for _, cpost := range cposts {
					if cpost.acct != k && !strings.HasPrefix(cpost.acct, lkey) {
						continue
					}
					v.Add(v, cpost.val)
				}

				// Add new posting
				sposts = append(sposts, Posting{
					date:  posts[transIdx].date,
					payee: posts[transIdx].payee,
					tnote: posts[transIdx].tnote,
					acct:  k,
					ccy:   toCCY,
					val:   v,
					note:  "",
					bal:   big.NewRat(0, 1),
					//acctlevels: newparts,
					acctlevel: newlevels,
					acctterm:  newterm,
				})
			}

			// Start again tracking
			accum = make(map[string]bool)
			cposts = cposts[:0]

			// New transaction index is this one
			transIdx = i
		}

		if i == len(posts) {
			break
		}

		p := posts[i]

		// Create new accumulated posting
		npost := p

		// Add to the posts
		eposts = append(eposts, npost)

		ncpost := PostKey{
			acct: p.acct,
			val:  p.val,
		}

		// Adjust currency if needed
		if p.ccy != toCCY {
			ncpost.val = big.NewRat(0, 1)
			ncpost.val.Mul(p.val, b.GetPrice(p.date, p.ccy, toCCY))

			// Accumulate for the base
			accum[ncpost.acct] = true
		}

		cposts = append(cposts, ncpost)

		// If not the first posting in the transaction...
		// ... check if there's an accumulation point
		if i != transIdx {
			key := findMaxCommonPrefix(eposts[i-1].acct, npost.acct, divider)
			if key != "" {
				accum[key] = true
			}
		}

		i++
	}

	// Add the new non-terminal posts to the terminal posts
	eposts = append(eposts, sposts...)

	// Sort it
	sort.Slice(eposts, func(i, j int) bool {
		return eposts[i].isLess(&eposts[j])
	})

	// Filter out hidden, adjust credit
	newp := make([]Posting, 0, len(eposts))
	for _, p := range eposts {
		if hidden != "" && p.GetAccount() == hidden {
			continue
		}
		newp = append(newp, p)
		if credit != nil && credit.MatchString(p.GetAccount()) {
			nval := big.NewRat(0, 1)
			nval.Neg(p.val)
			newp[len(newp)-1].val = nval
		}
	}
	eposts = newp

	// Turn into transactions
	trans := make([]Transaction, 0, len(eposts))
	lastIdx := 0
	i = 1
	for {
		if i == len(eposts) || eposts[i].date != eposts[lastIdx].date || eposts[i].payee != eposts[lastIdx].payee {
			trans = append(trans, eposts[lastIdx:i])
			if i == len(eposts) {
				break
			}
			lastIdx = i
		}
		i++
	}

	return trans
}
