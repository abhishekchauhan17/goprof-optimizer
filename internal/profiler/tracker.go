package profiler

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"unsafe"
)

// atomicCounter is used to generate unique IDs for suggestions, etc.
var atomicCounter uint64

func nextID(prefix string) string {
	n := atomic.AddUint64(&atomicCounter, 1)
	return fmt.Sprintf("%s-%d", prefix, n)
}

func trackAllocation(p *Profiler, obj any, tag string) {
	typ := reflect.TypeOf(obj)
	if typ == nil {
		return
	}

	typeName := typ.String()
	if tag == "" {
		tag = "default"
	}

	size := estimateSize(obj, typ)
	if size == 0 {
		// Avoid polluting stats with meaningless entries.
		return
	}

	key := typeName + "|" + tag

	p.mu.Lock()
	defer p.mu.Unlock()

	stat, ok := p.allocs[key]
	if !ok {
		stat = &AllocationStat{
			TypeName: typeName,
			Tag:      tag,
		}
		p.allocs[key] = stat
	}

	stat.AllocCount++
	stat.TotalAllocBytes += size
	if stat.AllocCount > 0 {
		stat.AverageAllocBytes = stat.TotalAllocBytes / stat.AllocCount
	}
}

// estimateSize attempts to estimate the size of obj in bytes. This is a
// heuristic approximation intended for relative comparisons, not exact
// accounting. It avoids deep traversals to keep overhead low.
func estimateSize(obj any, typ reflect.Type) uint64 {
	v := reflect.ValueOf(obj)

	switch typ.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return 0
		}
		// Size of the value pointed to; we ignore pointer shell size.
		elemType := typ.Elem()
		return estimateSize(v.Elem().Interface(), elemType)

	case reflect.Slice, reflect.Array:
		// For slices/arrays, approximate as len * element size.
		elemType := typ.Elem()
		elemSize := elemType.Size()
		if elemSize == 0 {
			return 0
		}
		return uint64(v.Len()) * uint64(elemSize)

	case reflect.Map:
		// Very rough approximation: len * (key + value) sizes.
		keySize := typ.Key().Size()
		valSize := typ.Elem().Size()
		entrySize := keySize + valSize
		if entrySize == 0 {
			return 0
		}
		return uint64(v.Len()) * uint64(entrySize)

	case reflect.String:
		return uint64(len(v.String()))

	case reflect.Struct:
		return uint64(typ.Size())

	default:
		// For all other kinds (int, float, etc.), use static size.
		return uint64(typ.Size())
	}
}

// unsafeSize can be used as a fallback for some types if needed. We keep it
// here for potential tuning.
func unsafeSize(typ reflect.Type) uint64 {
	return uint64(typ.Size())
}

// ptrSize indicates pointer size on this architecture. Useful for rough
// calculations if needed.
var ptrSize = uint64(unsafe.Sizeof(uintptr(0)))
