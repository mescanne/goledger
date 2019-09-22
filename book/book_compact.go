package book

import (
	"math/big"
	"sort"
)

// TODO: Use append rather than targetIdx. Far simpler.

func (b *Book) compact() {

	// Skip empty
	if len(b.post) == 0 {
		return
	}

	// Sort
	sort.Slice(b.post, func(i, j int) bool {
		return b.post[i].isLess(&b.post[j])
	})

	// Initialize new array
	targetIdx := 0

	// Compact - adding up common entries
	p := b.post
	for i := 1; i < len(p); i++ {

		if p[i].date == p[targetIdx].date &&
			p[i].payee == p[targetIdx].payee &&
			p[i].acct == p[targetIdx].acct &&
			p[i].ccy == p[targetIdx].ccy {

			// Add in the numbers
			p[targetIdx].val.Add(p[targetIdx].val, p[i].val)

		} else {

			// TODO: Incorporate zero-check here.
			// TODO: Incorporate tracking position of transactions here as well.
			//       This will allow us to track everything and iterate in a much
			//       more elegant way.

			// Otherwise start up a target.
			targetIdx++

			// If it's not our current index, then initialise it
			if targetIdx < i {
				p[targetIdx] = p[i]
				p[targetIdx].tnote = ""
				p[targetIdx].note = ""
			}
		}
	}

	// Shrink
	b.post = p[:targetIdx+1]

	// Compact - removing zero entries
	p = b.post
	targetIdx = 0
	var zero big.Int
	for i := 0; i < len(p); i++ {

		if p[i].val.Num().Cmp(&zero) == 0 {
			continue
		}

		if targetIdx < i {
			p[targetIdx] = p[i]
		}

		targetIdx++
	}

	// Shrink again, except exclude targetIdx this time
	b.post = p[:targetIdx]

	// Transaction re-index
	p = b.post
	lastIdx := 0
	i := 1
	b.trans = b.trans[:0]
	for {
		if i == len(p) || p[i].date != p[lastIdx].date || p[i].payee != p[lastIdx].payee {
			b.trans = append(b.trans, p[lastIdx:i])
			if i == len(p) {
				break
			}
			lastIdx = i
		}
		i++
	}

	// Shrink transactions (if needed)
	//b.trans = b.trans[:targetIdx]

	// Calculate balance
	amts := make(map[[2]string]*big.Rat)
	p = b.post
	for i := range p {
		key := [2]string{p[i].acct, p[i].ccy}
		v, ok := amts[key]
		if !ok {
			v = big.NewRat(0, 1)
			v.Set(p[i].val)
			amts[key] = v
		} else {
			v.Add(v, p[i].val)
		}
		p[i].bal.Set(v)
	}
}
