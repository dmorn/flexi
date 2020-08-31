// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"sync"
	"sort"
	"fmt"
)

type idPool struct {
	sync.Mutex
	free []int
	out  []int
}

// Get returns a unique integer. Subsequent Get calls will not
// return the same integer unless it is returned to the pool
// with Put.
func (p *idPool) Get() int {
	p.Lock()
	defer p.Unlock()
	if p.out == nil {
		p.out = []int{}
	}
	if len(p.free) == 0 {
		max := 0
		if len(p.out) > 0 {
			max = p.out[0]
			max++
		}
		p.free = []int{max}
	}

	id := p.free[0]
	p.free = p.free[1:]
	p.out = append(p.out, id)
	sort.Sort(sort.Reverse(sort.IntSlice(p.out)))
	return id
}

// Put returns i to the pool, meaning that subsequent Get
// calls might return it.
func (p *idPool) Put(i int) {
	p.Lock()
	defer p.Unlock()

	remove := -1
	for j, v := range p.out {
		if v == i {
			remove = j
			break
		}
	}
	if remove == -1 {
		// Caller asked to remove an indentifier
		// that was not in the out list.
		// Is it critical?
		return
	}
	p.out = append(p.out[:remove], p.out[remove+1:]...)

	if p.free == nil {
		p.free = []int{}
	}
	p.free = append(p.free, i)
	sort.Ints(p.free)
}

// Have works like Get without returning any integer.
// Use it to let p know you have i, and no other should.
// Returns an error if p knew already that i was out.
func (p *idPool) Have(i int) error {
	p.Lock()
	defer p.Unlock()
	if p.out == nil {
		p.out = []int{}
	}
	for _, v := range p.out {
		if i == v {
			return fmt.Errorf("%d is already tracked by the pool", i)
		}
	}
	p.out = append(p.out, i)
	sort.Sort(sort.Reverse(sort.IntSlice(p.out)))
	return nil
}
