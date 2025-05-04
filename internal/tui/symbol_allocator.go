package tui

import (
	"container/list"
)

type SymbolAllocator struct {
	symbols     []rune
	pidToSymbol map[int]rune
	symbolToPID map[rune]int
	lru         *list.List // list of pids, front = most recently used
	maxSymbols  int
}

func NewSymbolAllocator(symbols []rune) *SymbolAllocator {
	return &SymbolAllocator{
		symbols:     symbols,
		pidToSymbol: make(map[int]rune),
		symbolToPID: make(map[rune]int),
		lru:         list.New(),
		maxSymbols:  len(symbols),
	}
}

// AccessPID returns the symbol and its index for a given PID, assigning one if needed.
func (sa *SymbolAllocator) AccessPID(pid int) (rune, int) {
    // If already assigned, update LRU and return
    if sym, exists := sa.pidToSymbol[pid]; exists {
        sa.updateLRU(pid)
        return sym, sa.symbolIndex(sym)
    }

    // If we have a free symbol
    if len(sa.pidToSymbol) < sa.maxSymbols {
        for i, sym := range sa.symbols {
            if _, used := sa.symbolToPID[sym]; !used {
                sa.assignSymbol(pid, sym)
                return sym, i
            }
        }
    }

    // Evict least recently used
    lruElem := sa.lru.Back()
    if lruElem != nil {
        oldPID := lruElem.Value.(int)
        oldSym := sa.pidToSymbol[oldPID]
        delete(sa.pidToSymbol, oldPID)
        delete(sa.symbolToPID, oldSym)
        sa.lru.Remove(lruElem)

        sa.assignSymbol(pid, oldSym)
        return oldSym, sa.symbolIndex(oldSym)
    }

    // Should never happen if symbols > 0
    return '?', -1
}

func (sa *SymbolAllocator) assignSymbol(pid int, sym rune) {
	sa.pidToSymbol[pid] = sym
	sa.symbolToPID[sym] = pid
	sa.lru.PushFront(pid)
}

func (sa *SymbolAllocator) updateLRU(pid int) {
	for e := sa.lru.Front(); e != nil; e = e.Next() {
		if e.Value.(int) == pid {
			sa.lru.MoveToFront(e)
			return
		}
	}
	// If not found, add it
	sa.lru.PushFront(pid)
}

func (sa *SymbolAllocator) symbolIndex(sym rune) int {
    for i, s := range sa.symbols {
        if s == sym {
            return i
        }
    }
    return -1
}
