# Go From First Principles → Production Database Engineering

A complete, self-paced roadmap. No timelines. Every stage has: **what you learn**, **the first-principles questions you should be able to answer**, **exercises**, and **projects**.

The spine of this roadmap is a single idea: *you don't understand a language feature until you know what the machine does when you use it, and what breaks when you use it wrong.*

---

## How to use this document

Three tracks run in parallel, but they're ordered so that each unlocks the next:

| Track | What it builds |
|---|---|
| **A. Language & Runtime** (Phases 0–5) | Go itself, from syntax to the scheduler and GC |
| **B. Backend & Networking** (Phases 6–8) | TCP, HTTP, APIs, real services |
| **C. Database Capstone** (Phase 9) | A Postgres-style RDBMS from scratch |

Track B is deliberately placed *before* the database, because your database will be a network server. Building an HTTP/TCP server first means the capstone's client-server layer is a known quantity, not a new problem.

**Working rules for the whole roadmap:**

1. **Write the code before reading the explanation where possible.** Guess what a program prints, run it, then reconcile.
2. **Every phase produces committed code.** A `go-lab/` repo with one directory per topic.
3. **Read standard library source.** Go's stdlib is unusually readable. When you learn `sync.Once`, open `sync/once.go`. It's 20 lines.
4. **Benchmark claims.** When I say "this allocates," prove it with `go test -bench . -benchmem`.
5. **Use the race detector by default:** `go test -race`, `go run -race`.

---

# PHASE 0 — The Machine Model & Toolchain

Before syntax. You need a mental model of what Go *is* before learning what it *says*.

### 0.1 What Go actually is

- Statically typed, compiled to native machine code (no VM, unlike JVM/CLR).
- Garbage collected, with a concurrent tri-color mark-and-sweep collector.
- Ships a runtime that is **linked into your binary** — the scheduler, GC, memory allocator, and channel implementation all travel with your executable. This is why a hello-world binary is ~1.5MB.
- Designed at Google around three constraints: slow C++ builds, difficulty of concurrent programming, and dependency hell. Every design decision traces back to one of these.

**Questions to answer:**
- Why does Go have no classes, inheritance, or exceptions? What replaces each?
- What does "the runtime is in your binary" mean for deployment vs Python/Java?
- Why is Go's compiler so fast? (Hint: no header files, explicit imports, no circular imports, single-pass-ish design.)

### 0.2 Toolchain

```bash
go build      # compile
go run        # compile to temp dir + execute
go test       # test runner (built into the toolchain, not a library)
go vet        # static analysis for likely bugs
go fmt        # canonical formatting — non-negotiable in Go culture
go mod        # dependency management
go doc        # documentation from source
go tool pprof # profiler
go tool trace # execution tracer
```

Learn these flags now, you'll use them all roadmap long:

```bash
go build -gcflags="-m"       # escape analysis decisions
go build -gcflags="-m -m"    # verbose escape analysis
go build -gcflags="-S"       # generated assembly
go test -bench=. -benchmem   # benchmarks with allocation counts
go test -race                # race detector
go test -fuzz=FuzzName       # fuzzing
GODEBUG=gctrace=1 ./binary   # GC activity log
```

### 0.3 Modules & project layout

- `go.mod`, `go.sum`, semantic import versioning, `replace` directives, vendoring.
- Package = directory. Exported = capitalized identifier. There is no `public`/`private` keyword.
- `internal/` directory — enforced by the compiler, not convention.
- Why Go rejects the `src/main/java/com/company/...` style. Flat is better.

**Exercise:** Create a module with three packages where `internal/store` is importable by your `cmd/` but provably not by an external module. Try to import it from a second module and read the compiler error.

### 0.4 Project: `wc` clone

Rebuild `wc` (word count): read files or stdin, count lines/words/bytes/runes, support flags.

Teaches: `os.Args`, the `flag` package, `os.Stdin`, `bufio.Scanner`, error handling on the file path, exit codes. Small, but you'll immediately hit the bytes-vs-runes distinction which sets up Phase 1.

---

# PHASE 1 — Types, Memory, and the Value Model

This is the phase most people skip and pay for later. Go's semantics are *entirely* explainable by "everything is a value, and assignment copies the value." Slices, maps, and channels only *seem* to break that rule.

### 1.1 Variables, zero values, and declaration

- `var x int`, `x := 5`, `const`, `iota`.
- **Every type has a zero value and it's always usable.** `var buf bytes.Buffer` is a ready-to-use buffer. `var mu sync.Mutex` is an unlocked mutex. `var s []int` is a valid empty slice you can `append` to. This is a deliberate design principle: *make the zero value useful*.
- Shadowing bugs with `:=` inside `if`/`for` blocks.
- Untyped constants: `const x = 1 << 40` works even though it doesn't fit in `int32`. Constants are arbitrary precision until assigned.

**Questions:**
- Why is `var m map[string]int` readable but not writable, while `var s []int` is appendable?
- What does `iota` reset on, and how do you build a bit-flag enum with it?

### 1.2 Numeric types and representation

- `int` is 64-bit on modern platforms but is *not* the same type as `int64`. No implicit conversion, ever.
- Two's complement, wrap-around on overflow (defined behavior in Go, unlike C's UB for signed overflow).
- `uint` underflow: `var x uint = 0; x--` gives you `18446744073709551615`. This bites in loop conditions.
- Floating point: IEEE-754, why `0.1 + 0.2 != 0.3`, `math.NaN()`, `math.Inf()`, and why NaN != NaN.
- Integer division truncates toward zero; `%` follows the dividend's sign.

**Exercise:** Write a function that safely adds two `int32`s and returns an error on overflow. Then find how the stdlib does it in `math/bits` and `strconv`.

### 1.3 Strings, bytes, and runes — the UTF-8 model

This is a first-principles goldmine and it shows up constantly in database work.

- A `string` is an immutable, read-only slice of bytes: internally a struct `{ptr *byte, len int}`.
- Indexing a string gives a **byte**, not a character: `"héllo"[1]` is `0xC3`, not `é`.
- `range` over a string decodes UTF-8 and gives you `(byteIndex, rune)`.
- `rune` is an alias for `int32` — a Unicode code point.
- `len(s)` is byte length. `utf8.RuneCountInString(s)` is code-point count. Neither is "number of characters a human sees" (that's grapheme clusters).
- String concatenation in a loop is O(n²) — use `strings.Builder`.
- `[]byte(s)` copies. `string(b)` copies. The compiler elides the copy in specific cases (`map[string(b)]`, `switch string(b)`) — check with `-gcflags=-m`.

**Exercise:**
```go
s := "héllo"
fmt.Println(len(s))                  // predict before running
for i, r := range s { fmt.Println(i, r, string(r)) }
fmt.Println([]byte(s))
fmt.Println([]rune(s))
```
Then: write `Reverse(string) string` that works correctly for `"héllo"` and explain why the naive byte-reversal corrupts it.

**Exercise:** Benchmark string concatenation with `+=` vs `strings.Builder` vs `bytes.Buffer` for 100k appends. Explain the allocation counts.

### 1.4 Arrays vs Slices — slice internals

**Arrays** are values. `[5]int` and `[6]int` are *different types*. Passing an array to a function copies it entirely. Arrays are rare in application code but critical in low-level code (fixed-size page headers in your database will use them).

**Slices** are the workhorse. A slice is a 3-word header:

```go
type slice struct {
    ptr *T   // pointer to backing array
    len int  // number of accessible elements
    cap int  // elements from ptr to end of backing array
}
```

Everything about slices follows from this:

- Slicing shares the backing array: `b := a[1:3]` — writes through `b` are visible through `a`.
- `append` may or may not allocate. If `len < cap`, it writes in place and returns a header with a bumped `len`. If `len == cap`, it allocates a new array (roughly 2× for small slices, ~1.25× growth factor for large), copies, and returns a header pointing elsewhere. **The caller's slice is now aliasing a different array.**
- The full slice expression `a[low:high:max]` caps capacity — the tool for preventing accidental aliasing when you hand a sub-slice to someone else.
- A `nil` slice and an empty slice behave identically for `len`, `cap`, `range`, and `append`, but differ under `== nil` and JSON marshaling (`null` vs `[]`).

**The classic bug — reproduce it yourself:**
```go
a := []int{1, 2, 3, 4, 5}
b := a[:2]
b = append(b, 99)
fmt.Println(a) // [1 2 99 4 5] — you just clobbered a[2]
```

**Exercise set:**
1. Write `func describe(s []int)` printing len, cap, and `&s[0]`. Call it while appending in a loop and watch when the pointer changes.
2. Implement `Delete(s []T, i int) []T` (order-preserving and non-preserving versions). Explain the memory-leak risk when `T` contains pointers, and fix it by zeroing the tail.
3. Write a function that takes a slice and must not let the caller see its mutations. Show two ways: copy, and three-index slicing.
4. Implement your own `Map`, `Filter`, `Reduce` with generics (revisit after 1.11).

### 1.5 Maps

- A map is a pointer to a `hmap` struct — that's why maps are "reference-like" and why the zero value (`nil`) can be read but not written.
- Implementation: array of buckets, each bucket holds 8 key/value pairs plus top-hash bytes for fast rejection; overflow buckets chain; incremental rehashing on growth. (Go 1.24+ uses Swiss Tables — worth reading about the change and *why*: better cache behavior, fewer indirections.)
- **Iteration order is deliberately randomized.** Not "unspecified in practice" — actively randomized at runtime so you can't depend on it.
- Map elements are **not addressable**: `m[k].field = v` is a compile error for struct values. You must either use `*T` values or read-modify-write.
- The comma-ok idiom: `v, ok := m[k]` — distinguishes "missing" from "present with zero value."
- Maps are not safe for concurrent use. Concurrent write triggers a runtime *throw*, not a panic you can recover from. `sync.Map` exists for specific read-heavy patterns and is usually the wrong choice.
- Deleting during range is safe and defined; adding during range is undefined regarding whether you see the new entry.

**Exercise:** Implement a hash map from scratch — array of buckets, FNV-1a hash, separate chaining, then open addressing with linear probing. Benchmark against the builtin. This directly prepares you for the database's hash join and buffer-pool page table.

### 1.6 Structs, memory layout, and alignment

- Structs are values. Assignment copies all fields.
- **Field ordering changes struct size** because of alignment padding:
```go
type Bad  struct { a bool; b int64; c bool }  // 24 bytes
type Good struct { b int64; a bool; c bool }  // 16 bytes
```
  Verify with `unsafe.Sizeof`, `unsafe.Alignof`, `unsafe.Offsetof`. This matters enormously when you have millions of tuples in a buffer pool.
- Embedding (`struct { io.Reader }`) — composition with promoted fields/methods. *Not* inheritance: there's no virtual dispatch to an overriding "subclass."
- Struct tags: `json:"name"` — arbitrary metadata read via reflection.
- Comparability: structs are `==`-comparable iff all fields are. Structs with slice/map/func fields are not.
- The empty struct `struct{}{}` occupies zero bytes — used for set semantics (`map[string]struct{}`) and signal-only channels (`chan struct{}`).

**Exercise:** Take a struct with 6 mixed-type fields. Compute its size by hand from alignment rules, then verify with `unsafe.Sizeof`. Reorder to minimize. Then find `fieldalignment` in `go vet`'s analyzers and run it.

### 1.7 Pointers

- `&x`, `*p`, `new(T)`. No pointer arithmetic (that's what `unsafe` is for).
- Pointers exist for two reasons: **mutation across function boundaries** and **avoiding copies**.
- Go automatically takes addresses for method calls on addressable values: `v.PointerMethod()` becomes `(&v).PointerMethod()`.
- Nil pointer dereference is a runtime panic, not undefined behavior.
- **Escape analysis**: whether a value lives on the stack or heap is decided by the compiler, not by `new` vs literal. `return &x` from a function forces `x` to the heap. See it with `-gcflags=-m`.
- Passing a large struct by value can be *faster* than by pointer (stack allocation, no GC pressure, better cache locality). Measure, don't assume.

**Exercise:** Write four variants of a function returning a struct: by value, by pointer, filling a caller-provided pointer, and appending to a caller-provided slice. Benchmark with `-benchmem` and explain the allocation differences using `-gcflags=-m`.

### 1.8 Functions, closures, defer

- First-class functions, function types, higher-order functions.
- Multiple return values — the foundation of Go's error handling.
- Named return values and when they help (deferred error mutation) vs hurt (readability).
- Variadic functions and the `slice...` spread. Note that `f(s...)` passes `s` itself, not a copy — mutation inside `f` is visible.
- **Closures capture variables by reference, not value.** Since Go 1.22 loop variables are per-iteration, which fixed the single most common Go bug in history. Know what the old behavior was — you will read pre-1.22 code.
- `defer`:
  - Arguments are evaluated **at defer time**, body runs at function return.
  - LIFO order.
  - Runs on panic too — this is what makes it safe for cleanup.
  - Runs *after* the return value is set, so a deferred closure can modify a named return value.
  - Open-coded defers (Go 1.14+) made them nearly free; the old "defer is slow, avoid in hot loops" advice is mostly obsolete — but `defer` in a *loop body* still accumulates until function exit, which is a real leak pattern.

**Exercise — predict every output:**
```go
func f() (result int) {
    defer func() { result *= 2 }()
    return 5
}

func g() {
    for i := 0; i < 3; i++ {
        defer fmt.Println(i)
    }
}

func h() {
    x := 1
    defer fmt.Println("deferred:", x)
    x = 2
    fmt.Println("normal:", x)
}
```

### 1.9 Methods and method sets

- Methods are functions with a receiver. They can be defined on any named type you own — including `type Celsius float64`.
- **Value receiver vs pointer receiver:**
  - Value receiver gets a copy. Mutations are lost.
  - Pointer receiver can mutate; also avoids copying large structs.
  - Rule of thumb: be consistent per type. If any method needs a pointer receiver, use pointer receivers for all of them.
- **Method sets** — the rule that trips everyone up:
  - The method set of `T` includes methods with value receivers.
  - The method set of `*T` includes methods with value *and* pointer receivers.
  - Therefore: `*T` satisfies more interfaces than `T`. If `Save()` has a pointer receiver, `T` does not implement `Saver`, but `*T` does.
- Method values (`f := obj.Method`) capture the receiver. Method expressions (`f := Type.Method`) take the receiver as the first argument.

**Exercise:** Build a type where a value fails to satisfy an interface but its pointer succeeds. Read the compiler error carefully — it's one of Go's more helpful ones.

### 1.10 Interfaces — the heart of Go's abstraction

First principles: an interface value is a **two-word pair**: `(type descriptor, data pointer)`.

```go
type iface struct {
    tab  *itab          // type + method table for this (interface, concrete) pair
    data unsafe.Pointer // pointer to the value
}
```

Everything follows:

- **Implicit satisfaction.** No `implements` keyword. A type satisfies an interface if it has the methods. This means you can define an interface *for* a type you don't own — the consumer defines the abstraction.
- **The nil interface trap:**
```go
var p *MyError = nil
var err error = p
fmt.Println(err == nil) // false! tab is non-nil, data is nil
```
  This is the #1 subtle Go bug. Understand it at the two-word level, and you'll never write `return err` where `err` is a typed nil.
- **Accept interfaces, return structs.** Callers can then use the concrete type's full API.
- **Keep interfaces small.** `io.Reader` has one method and is arguably the most valuable interface ever designed. "The bigger the interface, the weaker the abstraction."
- **The empty interface** `interface{}` / `any` holds anything and tells you nothing. Type assertions (`v, ok := x.(T)`) and type switches recover the type.
- Interface method calls are **dynamic dispatch** through the itab — not free, and they inhibit inlining. Relevant for hot loops (your query executor will care).
- Interface satisfaction assertion at compile time: `var _ Storer = (*FileStore)(nil)`.

**Exercise set:**
1. Implement `io.Reader` and `io.Writer` for a custom type (e.g., a ROT13 reader, a counting writer, a rate-limited writer). Chain them: `io.Copy(rot13Writer, io.LimitReader(file, 100))`.
2. Reproduce the typed-nil bug, then use `reflect` or a type switch to detect it.
3. Benchmark a direct method call vs an interface call in a tight loop with `-benchmem`. Look at `-gcflags=-m` to see the inlining difference.

### 1.11 Errors

Go's error handling is a design position, not an oversight: errors are values, and control flow is explicit.

- `error` is just `interface { Error() string }`.
- Sentinel errors: `var ErrNotFound = errors.New("not found")` — compared with `errors.Is`.
- Custom error types with extra fields — extracted with `errors.As`.
- **Wrapping:** `fmt.Errorf("loading config: %w", err)` builds a chain. `%w` wraps (preserves identity), `%v` flattens (destroys it).
- `errors.Is` walks the chain comparing identity. `errors.As` walks it looking for a type match.
- `errors.Join` (Go 1.20+) for multiple errors.
- Error messages: lowercase, no trailing punctuation, and they should compose — each layer adds context, so the final message reads as a path: `"handle request: query users: connect: dial tcp: refused"`.
- **Never `_ = err`.** Never ignore errors from `Close()` on a writer — that's where buffered write failures surface.

**Panic/recover:**
- `panic` is for programmer errors and truly unrecoverable states, not control flow.
- `recover` only works inside a deferred function, in the same goroutine.
- **A panic in any goroutine kills the entire program** unless recovered *in that goroutine*. You cannot recover a panic from a goroutine you spawned. Critical for server design.
- Legitimate uses of panic/recover: parser bailout across deep recursion (your SQL parser will use this), and HTTP middleware that converts a handler panic into a 500.

**Exercise:** Build a layered application (handler → service → repository). Define domain errors at the bottom, wrap at each layer, and at the top map them to HTTP status codes using `errors.Is`/`errors.As`. This is the exact structure you'll use in Phase 7.

### 1.12 Generics

- Type parameters: `func Map[T, U any](s []T, f func(T) U) []U`.
- Constraints are interfaces: `comparable`, `constraints.Ordered`, union elements `~int | ~float64`, the `~` approximation for named types.
- Type inference and when you must annotate explicitly.
- Implementation: Go uses **GC shape stenciling** — one instantiation per pointer-shape group, with a dictionary passed for type-specific info. Consequence: generics can be *slower* than a concrete implementation and are not a performance feature.
- When *not* to use generics: if an interface expresses it cleanly, use the interface. Generics are for containers and algorithms where the type genuinely varies.
- Learn `slices`, `maps`, `cmp` from the stdlib — they're generic and you should use them instead of hand-rolling.

**Exercise:** Implement a generic `Set[T comparable]`, a generic LRU cache `LRU[K comparable, V any]`, and a generic binary search tree with `cmp.Ordered`. The LRU comes back in Phase 9 as your buffer pool's eviction policy.

### Phase 1 Projects

**P1.1 — `jsonq`, a JSON query CLI**
Read JSON from stdin/file, support dotted path queries (`.users[0].name`), pretty-print, filter. Teaches: `encoding/json`, `map[string]any` navigation, type switches, recursive descent (a warm-up for your SQL parser).

**P1.2 — In-memory key-value store (single-threaded)**
A `Store` type with `Get/Set/Delete/Keys/Expire`. No concurrency yet. Add a simple text protocol over stdin. Teaches: maps, methods, interfaces for pluggable storage, error design.

**P1.3 — Text-indexing tool**
Walk a directory, tokenize files, build an inverted index `map[string][]DocID`, support boolean queries (`AND`/`OR`/`NOT`). Teaches: `path/filepath.WalkDir`, `bufio`, string handling, set operations, sorting with `sort`/`slices`.

**P1.4 — A tiny expression evaluator**
Lexer + recursive-descent parser + evaluator for arithmetic with precedence, parentheses, unary minus, variables. **This is a deliberate rehearsal for the SQL parser in Phase 9.** Do it now while it's small.

---

# PHASE 2 — Concurrency From First Principles

Go's headline feature. Learn the mechanism first, patterns second, and treat "share memory by communicating" as a guideline rather than a law — mutexes are frequently the right answer.

### 2.1 The concurrency model

- **Concurrency is not parallelism.** Concurrency is a program structure that deals with many things at once; parallelism is executing many things simultaneously. Watch/read Rob Pike's talk on this.
- **CSP (Communicating Sequential Processes)** — Hoare's model. Independent processes communicating over channels.
- Goroutines are **not OS threads.** They start with ~2–8KB of growable stack (vs ~1–8MB for an OS thread), are multiplexed onto threads by the Go runtime, and cost roughly a few hundred nanoseconds to create. A million goroutines is normal; a million threads is not.

### 2.2 The GMP scheduler

You should be able to draw this on a whiteboard.

- **G** = goroutine (the task, with its stack and state).
- **M** = machine (an OS thread).
- **P** = processor (a scheduling context; count = `GOMAXPROCS`, defaults to CPU count). A P holds a **local run queue** of Gs.
- An M must hold a P to run Go code. G runs on M via P.
- **Work stealing:** an idle P steals half the runnable Gs from another P's local queue, or pulls from the global queue.
- **Handoff:** when a G makes a blocking syscall, the M blocks with it and the P is handed to another M so the remaining Gs keep running.
- **Netpoller:** network I/O doesn't block an M — the G is parked and registered with epoll/kqueue/IOCP; when the fd is ready, the G becomes runnable. **This is why a Go server handles 100k connections with a handful of threads while using simple blocking-style code.** Internalize this before Phase 6.
- **Preemption:** cooperative at function-call safepoints historically; since Go 1.14 the runtime can asynchronously preempt via signals, so a tight loop with no calls no longer starves the scheduler.
- Stack growth: contiguous stacks that are copied and doubled when they overflow. That's why you can't hold long-lived raw pointers into a goroutine stack across `unsafe` boundaries.

**Exercise:** Spawn 100,000 goroutines that each sleep and increment an atomic counter. Measure memory with `runtime.ReadMemStats`. Then do the same with `GOMAXPROCS=1`. Then run a CPU-bound loop with no function calls and observe (with `GODEBUG=schedtrace=1000`) how preemption behaves.

### 2.3 Goroutines and lifecycle

- `go f(x)` — arguments evaluated immediately, function runs later.
- `main` returning kills all goroutines instantly. There is no "wait for children."
- **Goroutine leaks** are Go's memory leak: a goroutine blocked forever on a channel is never collected. Every goroutine you start must have a defined exit path.
- Rule: *the function that starts a goroutine is responsible for its shutdown.* Pass it a `context.Context` or a done channel.
- Detecting leaks: `runtime.NumGoroutine()`, `go.uber.org/goleak` in tests, `/debug/pprof/goroutine`.

### 2.4 Channels

Internals: a channel is a pointer to an `hchan` struct containing a ring buffer (for buffered channels), a send queue, a receive queue (both queues of parked goroutines, `sudog`s), and a mutex.

- **Unbuffered channel = synchronous rendezvous.** The send blocks until a receiver is ready; both are then released. It's a synchronization primitive that happens to move data.
- **Buffered channel = asynchronous up to capacity.** Send blocks only when full.
- Send on nil channel: blocks forever. Receive on nil channel: blocks forever. (Useful in `select` to disable a case.)
- Send on closed channel: **panic**. Receive from closed channel: returns zero value immediately, `v, ok := <-ch` gives `ok == false`.
- Close is a broadcast — every blocked receiver wakes. This is the standard "signal N goroutines" mechanism.
- **Only the sender closes.** Never the receiver. With multiple senders, you need a separate coordination mechanism (a `sync.WaitGroup` plus a closer goroutine, or a done channel).
- `range ch` iterates until close.
- Directional types in signatures: `chan<- T` (send-only), `<-chan T` (receive-only). Use them — they document and enforce ownership.

**Exercise set:**
1. Implement a ping-pong between two goroutines with an unbuffered channel. Add a counter and measure round-trips/sec. Then compare with buffered capacity 1, 10, 1000.
2. Build a generator (function returning `<-chan int`) that produces primes lazily and can be cancelled.
3. Deliberately cause: deadlock (all goroutines asleep), send-on-closed panic, and a goroutine leak. Read each runtime message.

### 2.5 select

- Blocks until one case is ready; if several are ready, chooses **uniformly at random** (prevents starvation).
- `default` makes it non-blocking.
- Timeout: `case <-time.After(d)`. Note `time.After` allocates a timer that isn't collected until it fires — in a hot loop use `time.NewTimer` and `Stop()`.
- Disabling a case by setting its channel variable to `nil` — an elegant pattern for state machines.
- `for { select { ... } }` is the canonical event-loop shape for a long-running goroutine.

### 2.6 sync package

Do not treat mutexes as second-class. For protecting shared state, a mutex is usually simpler and faster than a channel.

- `sync.Mutex` — not reentrant. Locking twice in the same goroutine deadlocks. Since Go 1.9 it has a starvation mode: a waiter waiting >1ms triggers FIFO handoff.
- `sync.RWMutex` — many readers or one writer. Only wins under genuinely read-heavy contention; the read path is more expensive than a plain Mutex, and writer starvation is possible.
- `sync.WaitGroup` — `Add` before `go`, `Done` in a `defer`, `Wait` in the parent. Never `Add` inside the goroutine.
- `sync.Once` — read the source, it's a masterclass in atomics + mutex (double-checked locking done correctly).
- `sync.Cond` — rarely needed, but you'll want it for a bounded buffer pool. Understand `Wait` must be in a `for` loop checking the predicate (spurious wakeups / lost races).
- `sync.Pool` — reuse of short-lived allocations, cleared at each GC. Correct use: reset the object before returning it. Your database's page and tuple buffers will use this.
- `sync/atomic` — `atomic.Int64`, `CompareAndSwap`, `atomic.Value`, `atomic.Pointer[T]`. The building block for lock-free structures.

**Exercise:** Implement a counter four ways — unsynchronized (prove the race with `-race`), mutex, atomic, and channel-owned (a goroutine owning the state). Benchmark all four at 1, 4, and 64 concurrent goroutines. Explain the shape of the results.

### 2.7 The Go Memory Model & data races

Read the official Go Memory Model document. Then answer:

- What is a **happens-before** relationship, and which operations establish one? (channel send/receive, mutex lock/unlock, `sync.Once`, `WaitGroup`, atomics with sequential consistency, goroutine creation.)
- Why is an unsynchronized read of a variable another goroutine writes *undefined*, not merely "possibly stale"? (Compiler reordering, CPU store buffers, torn reads on non-atomic multi-word values like interfaces and slice headers.)
- Why can a data race on an interface value cause a segfault, not just a wrong value?

**Exercise:** Write a program with a benign-looking race (a `bool` "done" flag). Run without `-race` — it may work forever. Run with `-race` — it fails immediately. Then build with optimizations disabled and see if behavior changes. This teaches you that "it works on my machine" means nothing for concurrency.

### 2.8 context

- `context.Context` carries **cancellation, deadlines, and request-scoped values** across API boundaries.
- `Background()`, `TODO()`, `WithCancel`, `WithTimeout`, `WithDeadline`, `WithValue`, `WithCancelCause` (1.20+), `AfterFunc` (1.21+).
- Contexts form a tree; cancelling a parent cancels all descendants.
- **Always call `cancel`**, even on the timeout variants — `defer cancel()` — or you leak the timer and the goroutine watching it.
- `ctx` is always the **first parameter**, never stored in a struct (with narrow exceptions).
- `ctx.Err()` returns `context.Canceled` or `context.DeadlineExceeded`.
- `WithValue` is for request-scoped metadata (request ID, auth subject) only — not for passing optional arguments. Use unexported key types to avoid collisions.

**Exercise:** Build a pipeline of three stages where cancelling the top-level context tears down every stage cleanly, verified with `goleak`.

### 2.9 Concurrency patterns

Implement each from scratch:

1. **Worker pool** — fixed N goroutines pulling from a jobs channel, writing to a results channel.
2. **Fan-out / fan-in** — one producer, N workers, merged output. Get the close semantics right with `WaitGroup` + closer goroutine.
3. **Pipeline** — stages connected by channels, each stage a function `func(<-chan In) <-chan Out`, all cancellable.
4. **Semaphore** — bounded concurrency via a buffered channel, or `golang.org/x/sync/semaphore` for weighted.
5. **errgroup** — `golang.org/x/sync/errgroup`: run N tasks, cancel all on the first error, collect it. Learn `SetLimit`.
6. **Rate limiter** — token bucket by hand, then compare to `golang.org/x/time/rate`.
7. **Single-flight** — deduplicate concurrent identical requests (`golang.org/x/sync/singleflight`). Implement it yourself first; it's a beautiful use of `sync.Mutex` + `WaitGroup`.
8. **Graceful shutdown** — signal handling, context cancellation, draining in-flight work with a timeout.
9. **Timeouts & circuit breaker** — state machine with atomic counters.
10. **Pub/sub broker** — topic → subscriber channels, with slow-consumer handling (drop, block, or buffer? — a real design decision).

### 2.10 Concurrency anti-patterns

- Unbounded goroutine creation per request (no backpressure → OOM).
- Sending to a channel nobody reads → leak.
- `time.After` in a hot select loop → timer churn.
- Holding a mutex across an I/O call or a channel operation → convoy/deadlock.
- Using `sync.Map` reflexively. It's optimized for a specific pattern (write-once, read-many, disjoint key sets per goroutine). Benchmark against `map` + `RWMutex` before adopting.
- Copying a struct containing a `sync.Mutex` (that's what `go vet`'s copylocks catches).

### Phase 2 Projects

**P2.1 — Concurrent web crawler**
BFS over links, bounded worker pool, visited set with mutex, depth limit, per-host rate limiting, full context cancellation, graceful shutdown on Ctrl-C. Verify zero goroutine leaks. This is the canonical Go concurrency project and it exercises nearly every pattern above.

**P2.2 — Concurrent KV store (upgrade P1.2)**
Add `RWMutex`, then shard the map into N shards keyed by hash to reduce contention. Add TTL expiry via a background sweeper goroutine. Benchmark single-lock vs sharded at increasing goroutine counts. Prove correctness under `-race`.

**P2.3 — Parallel file processor**
Walk a directory tree, hash every file (SHA-256), find duplicates. Bounded parallelism, streaming results, progress reporting, cancellable. Compare wall time at various worker counts and explain the point of diminishing returns (I/O bound vs CPU bound).

**P2.4 — Job queue with retries**
In-memory queue, N workers, exponential backoff, dead-letter queue, at-least-once semantics, graceful drain. Introduces durability questions you'll answer for real in Phase 9's WAL.

---

# PHASE 3 — Runtime, Memory & Performance

This phase is what separates "writes Go" from "engineers systems in Go." It is also directly load-bearing for your database.

### 3.1 Memory allocator

- Based on TCMalloc. Three-level hierarchy: **mcache** (per-P, lock-free), **mcentral** (per size-class, locked), **mheap** (global, manages arenas from the OS).
- **Size classes**: ~68 classes. An allocation of 17 bytes gets a 24-byte slot. Tiny allocator batches sub-16-byte pointer-free objects.
- Objects >32KB go straight to the heap as large objects.
- Spans, arenas (64MB), and page management.
- Consequence: allocation is *cheap* but not free, and internal fragmentation is real. Sizing your database page structs to size-class boundaries matters.

### 3.2 Garbage collector

- **Concurrent, tri-color, mark-and-sweep, non-moving, non-generational.**
- Tri-color invariant: white (unreached), grey (reached, children unscanned), black (reached, children scanned). The invariant: **no black object points to a white object.**
- **Write barrier** (Dijkstra-style, hybrid since 1.8) preserves the invariant while the mutator runs concurrently with the marker.
- Phases: sweep termination → mark (concurrent, with a brief STW to enable write barriers) → mark termination (STW) → sweep (concurrent, lazy).
- **Stop-the-world pauses are sub-millisecond**, but *assist* work is charged to allocating goroutines — a high allocation rate slows your own code.
- `GOGC` (default 100 = heap doubles before next cycle), `GOMEMLIMIT` (soft memory cap, 1.19+ — the fix for containerized OOM kills).
- Go's GC is **non-generational**, which is unusual. Understand the argument: escape analysis + value types mean fewer short-lived heap objects than in Java, so a nursery buys less.
- **Reducing GC pressure**: fewer pointers (a `[]int` is scanned trivially; a `[]*T` is scanned deeply), object pooling, arena-style batch allocation, preallocated slices/maps with capacity hints.

**Exercise:** Write a program allocating 10M small objects. Run with `GODEBUG=gctrace=1`. Then: (a) preallocate with capacity, (b) switch `[]*T` to `[]T`, (c) add a `sync.Pool`. Measure GC cycles, heap size, and pause times at each step.

### 3.3 Escape analysis, stack vs heap

- The compiler proves whether a value's lifetime is bounded by the function. If yes → stack (free deallocation). If no → heap.
- Common escape causes: returning a pointer, storing in an interface, capture by a closure that outlives, sending on a channel, slice/map of pointers, value of unknown size at compile time, `fmt.Println` (variadic `...any` forces interface boxing).
- Read `-gcflags='-m -m'` output fluently. This skill pays off forever.

**Exercise:** Take a hot function that allocates, read the escape analysis output, and restructure to eliminate the allocation (e.g., accept a `[]byte` buffer to fill instead of returning a new one — exactly how `strconv.AppendInt` and `time.Time.AppendFormat` are designed). Benchmark before/after.

### 3.4 Profiling & observability of your own code

Non-negotiable toolkit:

- **Benchmarks**: `func BenchmarkX(b *testing.B)`, `b.ResetTimer()`, `b.ReportAllocs()`, `b.RunParallel`, sub-benchmarks with `b.Run`. Beware dead-code elimination — assign to a package-level sink.
- **benchstat** (`golang.org/x/perf/cmd/benchstat`) — compare benchmark runs with statistical significance. Never claim a speedup from a single run.
- **CPU profile**: `go test -cpuprofile`, or `net/http/pprof` in a running server. Read flat vs cumulative, and use `go tool pprof -http=:8080` for the flame graph.
- **Heap profile**: `-memprofile`, `inuse_space` vs `alloc_space` — the difference between "leaking" and "churning."
- **Block & mutex profiles**: where goroutines wait on synchronization. Essential for your database's latch contention.
- **Execution trace**: `go tool trace` — see goroutine scheduling, GC, syscalls, and network blocking on a timeline. The single best tool for understanding concurrency behavior.

**Exercise:** Take your P2.2 sharded KV store. Profile it under load. Find the contention with the mutex profile, the allocations with the heap profile, and the scheduling behavior with the tracer. Optimize one thing, prove it with benchstat.

### 3.5 unsafe, reflect, and going below the language

- `unsafe.Pointer` conversion rules (there are exactly 6 valid patterns — learn them). `uintptr` is not a pointer and does not keep objects alive.
- `unsafe.Sizeof/Alignof/Offsetof`, `unsafe.Slice`, `unsafe.String`, `unsafe.SliceData` (1.20+) — the modern, safer way to do zero-copy `[]byte` ↔ `string`.
- Zero-copy conversions matter in a database: you will read a page from disk into `[]byte` and want to interpret bytes as a tuple header without copying.
- `reflect`: `Type`, `Value`, `Kind`, settability rules, struct tags. How `encoding/json` works. Slow — use for configuration and serialization, not hot paths.
- `go:linkname`, `go:noescape`, build tags, and assembly (`.s` files) — awareness level; know they exist and where the stdlib uses them (`math/bits`, `crypto`).

**Exercise:** Write a zero-copy `BytesToString` and `StringToBytes` using `unsafe`. Write a test that demonstrates *why* the resulting string must never be mutated. Then benchmark against the safe versions.

### 3.6 Compiler behavior worth knowing

- **Inlining**: budget-based; check with `-gcflags=-m`. `//go:noinline` for benchmarking.
- **Bounds check elimination**: `_ = s[len(s)-1]` hints, and the `for i := range s` form. Verify with `-gcflags=-d=ssa/check_bce/debug=1`.
- **Devirtualization** of interface calls when the concrete type is provable.
- **PGO** (profile-guided optimization, 1.21+): drop a `default.pgo` CPU profile in your main package and the compiler makes better inlining decisions. Free 2–10% on real workloads — you will use this on your database.

---

# PHASE 4 — Standard Library, I/O & Engineering Practice

### 4.1 The io abstraction

`io.Reader` and `io.Writer` are the most important interfaces in Go. Learn the whole family:

- `Reader`, `Writer`, `Closer`, `Seeker`, `ReaderAt`, `WriterAt`, `ReaderFrom`, `WriterTo`, `ByteReader`, `StringWriter`.
- **`ReaderAt`/`WriterAt` are what your database will use** — positional, concurrency-safe I/O without a shared file offset. This is why `os.File` implements them.
- Composition: `io.MultiReader`, `io.MultiWriter`, `io.TeeReader`, `io.LimitReader`, `io.Pipe`, `io.Copy` (and why `io.Copy` is fast — it uses `ReaderFrom`/`WriterTo`/`sendfile` when available).
- `io.EOF` is a *value*, not an error condition to log. `io.ErrUnexpectedEOF` is different and meaningful.
- The `Read` contract: may return `0 < n < len(p)`; may return `n > 0` **and** `err == io.EOF` in the same call. Handle `n` before `err`. Most people get this wrong. Use `io.ReadFull` when you need exactly N bytes — you will, constantly, in the database.

### 4.2 Buffered and byte-level I/O

- `bufio.Reader`/`Writer`/`Scanner`/`ReadWriter`. Scanner's default 64KB token limit and how to raise it. Custom `SplitFunc` — you'll write one for a wire protocol.
- **Always `Flush()` a `bufio.Writer`**, and check the error. `defer w.Flush()` silently drops errors — handle it explicitly for durability-critical writes.
- `bytes.Buffer`, `bytes.Reader`, `strings.Reader`, `strings.Builder`.
- `encoding/binary`: `binary.LittleEndian.PutUint32`, `binary.Read/Write`, `binary.AppendUvarint`, varint encoding. **This is the core of your database's on-disk format.** Master it now: fixed-width vs varint tradeoffs, endianness choice and why LittleEndian on x86/ARM, alignment in serialized structs.

**Exercise:** Define a binary record format: `[magic:4][version:2][flags:2][length:4][payload:N][crc32:4]`. Write encoder and decoder using `encoding/binary` directly on a `[]byte` (no `binary.Read` reflection path). Fuzz the decoder to make sure malformed input never panics or over-reads. **This exercise is a direct component of Phase 9.**

### 4.3 Files, OS, and durability

- `os.File`, `Open` vs `OpenFile` with flags (`O_RDWR|O_CREATE|O_APPEND|O_SYNC`), permissions.
- `ReadAt`/`WriteAt`, `Seek`, `Truncate`, `Stat`.
- **`Sync()` = fsync.** Understand the stack: your write → Go's buffer → OS page cache → disk controller cache → platter/flash. `fsync` is the only thing that gives you durability, and it's expensive (milliseconds). **This single fact drives the entire design of write-ahead logging.**
- Atomic file replacement: write to temp file → fsync file → rename → fsync directory. Know why the directory fsync is needed.
- File locking (`flock` via `syscall`), and why a database needs it to prevent two processes opening the same data dir.
- `io/fs` and `fs.FS` — the filesystem abstraction, `os.DirFS`, `embed.FS`.
- `mmap` via `golang.org/x/exp/mmap` or raw `syscall.Mmap`. Understand the tradeoff: mmap gives you the OS page cache for free, but you lose control over eviction and I/O ordering — the reason "Are You Sure You Want to Use MMAP in Your Database Management System?" (Crotty, Leis, Pavlo) exists. **Read that paper before Phase 9.4.**

### 4.4 Serialization

- `encoding/json`: struct tags, `omitempty`, custom `MarshalJSON`/`UnmarshalJSON`, `json.Decoder` for streaming, `json.RawMessage` for deferred parsing, and the performance cost of reflection.
- `encoding/gob` — Go-native, self-describing, no cross-language use.
- Protobuf/gRPC (Phase 8).
- Custom binary formats (4.2) — what you'll actually use for pages and WAL records.

### 4.5 Time

- `time.Time` contains wall clock **and** monotonic reading. Subtraction uses monotonic (immune to NTP jumps); `t.Round(0)` strips it.
- **Never measure elapsed time by subtracting `time.Now()` values that crossed a serialization boundary** — the monotonic part is lost.
- `time.Duration` is an `int64` of nanoseconds. `time.Timer`, `Ticker` (must be `Stop`ped), `time.AfterFunc`.
- Making time testable: inject a `Clock` interface. You'll need this for MVCC timestamps and lock timeouts.

### 4.6 Testing — treat this as a first-class skill

- **Table-driven tests** — the Go idiom. `map[string]struct{...}` or `[]struct{name string; ...}` with `t.Run(name, ...)`.
- `t.Helper()`, `t.Cleanup()`, `t.Parallel()` and its interaction with subtests.
- `testing.TB` interface so helpers work in tests and benchmarks.
- **Golden files** for complex output (your query planner's EXPLAIN output).
- **Fuzzing**: `func FuzzParse(f *testing.F)`, seed corpus, `f.Fuzz(func(t *testing.T, data []byte){...})`. Non-negotiable for parsers and binary decoders.
- **Property-based testing** ideas: for your B+Tree, insert random keys and assert the invariants (sorted, balanced, all leaves at same depth) after every operation.
- `httptest` for HTTP, `net.Pipe` for connection-level tests.
- Test doubles via interfaces. Go doesn't need a mocking framework; hand-written fakes are usually better. Know `go.uber.org/mock` exists.
- Integration tests with build tags (`//go:build integration`) and `testcontainers-go` for real Postgres in CI.
- Coverage: `go test -cover`, `-coverprofile`, and why 100% coverage is not the goal.
- **Deterministic simulation testing** — advanced but transformative for a database: run your system on a seeded fake clock, fake filesystem, and fake network so failures are reproducible. Read about FoundationDB's approach and TigerBeetle's VOPR.

**Exercise:** Retrofit P1.4 (expression evaluator) with a full table-driven suite plus a fuzzer. Let the fuzzer find at least one crash. Fix it. Add the crasher to your seed corpus.

### 4.7 Engineering hygiene

- Project layout: `cmd/`, `internal/`, `pkg/` (and the argument against `pkg/`).
- Dependency injection by hand — constructor functions taking interfaces. No framework needed.
- `log/slog` (1.21+): structured logging, handlers, levels, `slog.With` for context, custom handlers. Replace `log` entirely.
- Configuration: flags → env → file, precedence order, validation at startup, fail fast.
- Linting: `golangci-lint` with `errcheck`, `staticcheck`, `govet`, `ineffassign`, `bodyclose`, `sqlclosecheck`, `fieldalignment`.
- `//go:generate`, `embed` package for static assets.
- Build: cross-compilation (`GOOS`/`GOARCH`), `-ldflags "-X main.version=..."` for build metadata, `-trimpath`, static vs cgo-linked binaries, minimal Docker images (`FROM scratch` / `distroless`).
- Semantic versioning, `go.mod` major-version suffixes, API compatibility promises.

### Phase 4 Projects

**P4.1 — Log-structured file store**
An append-only file storing `key → value` records with an in-memory hash index of `key → file offset`. Support get/put/delete (tombstones), file rotation, and compaction into a new segment. Crash-safe: reconstruct the index by replaying the log at startup.

This is essentially **Bitcask**, and it is the single best bridge project between Go and databases. You'll implement: binary record encoding, CRC checksums, `ReadAt` for point lookups, fsync policy, segment merging, and startup recovery. Everything in Phase 9 will feel familiar afterward.

**P4.2 — `mini-tar` / archive tool**
Pack and unpack a directory into a custom archive format with a header table, per-file metadata, and checksums. Add streaming compression via `compress/gzip` wrapping your writer. Teaches io composition deeply.

**P4.3 — Structured log aggregator CLI**
Tail multiple JSON log files, parse, filter by level/time/field, aggregate counts, output as table or JSON. Concurrency + io + time + generics all at once.

---

# PHASE 5 — Idiomatic Go & Design

A short but important phase: consolidate style before writing large systems.

- **API design**: accept interfaces, return structs; small interfaces; functional options pattern for constructors with many optional parameters; avoid config structs with 20 fields.
- **Package design**: name packages for what they provide, not what they contain (`http`, not `utils`). Avoid `util`, `common`, `helpers`, `base`, `manager`. Package names should read well at the call site: `bytes.Buffer`, not `bytes.BytesBuffer`.
- **Naming**: short names for short scopes (`i`, `r`, `buf`), longer for package-level. Getters have no `Get` prefix. Interfaces with one method end in `-er`.
- **Comments and docs**: doc comments start with the identifier name and form complete sentences. `Example` functions in `_test.go` become runnable documentation.
- **Concurrency in APIs**: document whether a type is safe for concurrent use. If not, say so. Don't add locks to types that don't need them.
- **Errors in APIs**: expose sentinel errors or typed errors for anything callers must branch on. Don't force string matching.
- **Read these**: *Effective Go*, *Go Code Review Comments*, Google's Go Style Guide, Dave Cheney's *Practical Go*, Rob Pike's *Go Proverbs*.
- **Read this source code** (in order of value): `io`, `sync`, `sort`/`slices`, `net/http`, `database/sql`, `context`, then `go/scanner` + `go/parser` (before writing your SQL parser), then `bbolt` (before writing your B+Tree).

**Exercise:** Take one of your earlier projects and refactor its public API using the functional options pattern, small interfaces, and proper error types. Write the package doc comment first, as a design tool.

---

# PHASE 6 — Networking From First Principles

Now Track B. Everything here directly feeds the database's client-server layer.

### 6.1 The network stack underneath you

- OSI/TCP-IP layers, but focus on what you control: sockets, TCP, and application protocols.
- TCP: the three-way handshake, sequence numbers, ACKs, sliding window, congestion control (slow start, AIMD), Nagle's algorithm and `TCP_NODELAY`, delayed ACKs, and why their interaction causes 40ms latency spikes in request/response protocols.
- **TCP is a byte stream, not a message stream.** There is no `recv()` that gives you "one message." You must frame messages yourself: length-prefixing, delimiters, or fixed-size records. This single fact is why every protocol has a framing layer.
- Connection lifecycle, `TIME_WAIT`, half-close, keep-alives.
- Backpressure: what happens when the receiver doesn't read — the sender's `Write` blocks. Understand this before designing a server.

### 6.2 The net package

- `net.Listen("tcp", ":8080")`, `Accept()` loop, `net.Conn` (which is an `io.ReadWriteCloser`).
- **The Go server model**: one goroutine per connection, blocking reads. Because of the netpoller (2.2), this scales to tens of thousands of connections without threads. You write simple sequential code and get epoll performance.
- `SetDeadline`, `SetReadDeadline`, `SetWriteDeadline` — the *only* reliable way to time out a network read. Combine with context by having a goroutine set a deadline on `ctx.Done()`.
- UDP with `net.PacketConn` — no connection, no ordering, no reliability.
- Unix domain sockets — faster local IPC, used by Postgres for local connections.
- `net.Dialer` with timeouts, `net.Resolver`, and DNS behavior.
- TLS with `crypto/tls`: `tls.Listen`, certificates, `tls.Config`, mutual TLS.

### 6.3 Protocol design & implementation

Design decisions you must be able to reason about:
- Text vs binary. Human-debuggable vs compact/fast.
- Framing: length-prefix (fast, needs a max-size guard) vs delimiter (simple, needs escaping) vs self-describing.
- Request/response correlation: sequential vs pipelined vs multiplexed with request IDs.
- Versioning and forward compatibility.
- Error signaling.

### Phase 6 Projects

**P6.1 — TCP echo & chat server**
Start with echo. Then a chat server: a hub goroutine owning the client set, per-client read and write goroutines, broadcast via channels, graceful client disconnect, and a slow-client policy. Add `SetReadDeadline` heartbeats.

**P6.2 — Redis clone, done properly (RESP protocol)**

You already put "mini-Redis in Go" on your résumé. Rebuild it from the ground up so it's genuinely yours again, and go further than last time:

- Implement the **RESP2** protocol parser and serializer from the spec (simple strings, errors, integers, bulk strings, arrays, and inline commands). Then RESP3 if you want.
- Commands: `PING`, `ECHO`, `SET`/`GET`/`DEL`/`EXISTS`, `SET ... EX/PX/NX/XX`, `INCR`/`DECR`, `EXPIRE`/`TTL`, `KEYS`, `TYPE`.
- Data types beyond strings: lists (`LPUSH`/`RPUSH`/`LRANGE`), hashes (`HSET`/`HGET`/`HGETALL`), sets (`SADD`/`SMEMBERS`), and sorted sets with a **skip list** (`ZADD`/`ZRANGE`/`ZRANGEBYSCORE`) — the skip list is a genuinely valuable data structure to implement and a nice contrast to the B+Tree you build later.
- Expiration: lazy on access + an active sampling sweeper, like real Redis.
- Persistence: **AOF** (append-only file with fsync policies: always/everysec/no) and **RDB-style snapshots**. Compare durability and recovery time. This is a direct rehearsal for WAL vs checkpoint in Phase 9.
- Pub/Sub, transactions (`MULTI`/`EXEC`), pipelining, keyspace notifications.
- Benchmark with the real `redis-benchmark` tool. Profile and optimize until you're within a reasonable factor of real Redis. Then use `go tool trace` to see where you lose.
- Test with the real `redis-cli` — if the official client can talk to your server, your protocol implementation is correct.

**P6.3 — Simple HTTP/1.1 server without `net/http`**
Parse the request line and headers off a raw TCP connection, handle `Content-Length` and chunked bodies, write a valid response, support keep-alive. **Do this before using `net/http`.** Once you've written the parser, `net/http` stops being magic.

---

# PHASE 7 — Backend Development in Go

### 7.1 net/http from the inside

- `http.Handler` is one method: `ServeHTTP(ResponseWriter, *Request)`. Everything in the ecosystem is that interface.
- `http.HandlerFunc` — an adapter that turns a function into a `Handler`. Read its two-line definition; it's the clearest example of Go's design philosophy in the stdlib.
- `http.ServeMux`, and the Go 1.22 enhanced routing: `mux.HandleFunc("GET /users/{id}", h)` with `r.PathValue("id")`. **The stdlib router is now good enough for most services** — learn it before reaching for chi/echo/gin.
- `http.Server` configuration you must set in production: `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `MaxHeaderBytes`. Defaults are unlimited and that's a DoS vector.
- `Server.Shutdown(ctx)` for graceful shutdown; combine with `signal.NotifyContext`.
- Request internals: `r.Context()`, `r.Body` (must be closed, is a stream), `http.MaxBytesReader`, `r.URL.Query()`, form parsing.
- `ResponseWriter` mechanics: headers must be set *before* `WriteHeader`; the first `Write` implies 200; `http.Flusher` for streaming and SSE; `http.Hijacker` for WebSocket upgrades.
- The client side: **never use `http.DefaultClient`** in production — no timeout. Configure `Transport` (connection pooling, `MaxIdleConnsPerHost`), always `defer resp.Body.Close()`, and always drain the body if you want connection reuse.

### 7.2 Middleware

Middleware is just `func(http.Handler) http.Handler` — decorator pattern, no framework required.

Build a chain from scratch: request ID, structured logging, panic recovery, timeout, CORS, gzip, auth, rate limiting, metrics. Understand ordering (recovery outermost, logging next, auth after that). Then implement a `Chain` helper so `Chain(h, mw1, mw2, mw3)` applies in readable order.

### 7.3 API design & implementation

- REST semantics: resources, correct verbs, status codes that mean something, idempotency, pagination (offset vs cursor — and why cursor wins at scale), filtering, sorting, partial responses.
- Request validation: decode into a DTO, validate explicitly, return field-level errors. Consider `go-playground/validator` but know how to do it by hand.
- Consistent error envelope: `{"error": {"code": "...", "message": "...", "details": [...]}}`, mapped from your domain errors via `errors.As`.
- Content negotiation, compression, ETags and conditional requests, `Cache-Control`.
- API versioning strategies.
- OpenAPI: generate from code or code from spec (`oapi-codegen`).

### 7.4 Persistence with a real database

This is where you meet Postgres as a *user*, right before you build one.

- `database/sql`: it's an abstraction over drivers with a built-in **connection pool**. `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime` — and the failure modes when they're wrong.
- `Query` vs `QueryRow` vs `Exec`. `rows.Close()` and why a leaked `Rows` holds a connection hostage. `rows.Err()` after the loop — the error people forget.
- **Prepared statements**: what the database does with them, why they prevent SQL injection structurally (the query plan is fixed before data arrives), and their cost with a pooler.
- `NULL` handling: `sql.NullString`, `*string`, `sql.Null[T]` (1.22+).
- Transactions: `BeginTx`, isolation level selection, `defer tx.Rollback()` after a successful commit is a no-op — the correct idiom.
- `pgx` — the Postgres-native driver. Native protocol, binary format, `CopyFrom` for bulk load, `pgxpool`, listen/notify. Understand what it does that `database/sql` can't express.
- `sqlc` (generate type-safe Go from SQL) vs `sqlx` vs GORM. Form an opinion; for a systems-minded engineer, `sqlc` or raw `pgx` is usually right.
- Migrations: `golang-migrate` or `goose`. Up/down, versioning, and why you never edit an applied migration.
- **Schema and query craft you'll need for Phase 9**: indexes (B-tree, hash, GIN, partial, covering), `EXPLAIN (ANALYZE, BUFFERS)`, query plans (seq scan vs index scan vs bitmap heap scan; nested loop vs hash join vs merge join), when the planner picks each, `VACUUM` and bloat, `pg_stat_statements`.
  Spend real time in `EXPLAIN ANALYZE`. In Phase 9 you're going to *build* the thing producing those plans.

### 7.5 Auth, security, and production concerns

- Password hashing with `bcrypt`/`argon2` (`golang.org/x/crypto`). Never MD5/SHA.
- Sessions (server-side, opaque token) vs JWT (stateless, self-contained) — real tradeoffs: revocation, size, key rotation.
- OAuth2 flows, `golang.org/x/oauth2`.
- Common vulnerabilities and their Go-specific mitigations: SQL injection (parameterized queries), XSS (`html/template` auto-escapes contextually — read how), CSRF, SSRF, path traversal (`filepath.Clean` + validation), open redirects, timing attacks (`crypto/subtle.ConstantTimeCompare`).
- Secrets management, TLS termination, security headers.
- `crypto/rand` vs `math/rand` — and never confusing them.

### 7.6 Observability

- **Structured logging** with `log/slog`: consistent fields, request-scoped loggers via context, log levels that mean something, no logging of secrets.
- **Metrics** with Prometheus (`prometheus/client_golang`): counters, gauges, histograms; RED (Rate, Errors, Duration) for services and USE (Utilization, Saturation, Errors) for resources; cardinality discipline.
- **Tracing** with OpenTelemetry: spans, context propagation across services, sampling.
- **Health checks**: liveness vs readiness, and why they're different.
- `net/http/pprof` exposed on an internal-only port.

### 7.7 Architecture

- Layering: transport (HTTP) → service (business logic) → repository (persistence). Dependencies point inward; the service layer knows nothing about HTTP.
- Domain errors defined in the service layer, translated at the transport layer.
- Interfaces defined by the *consumer*, in the consumer's package.
- Testing strategy per layer: unit tests on services with fake repositories, integration tests on repositories against a real Postgres in a container, end-to-end tests through `httptest.Server`.
- Background workers, outbox pattern for reliable event publishing, idempotency keys.

### Phase 7 Projects

**P7.1 — Full REST API service**
Pick a domain with real complexity — a task/project tracker, or a URL shortener with analytics. Requirements:
- Stdlib routing (1.22+), full middleware chain, graceful shutdown.
- Postgres via `pgx` + `sqlc`, migrations, transactions with correct isolation.
- Auth (register/login, hashed passwords, sessions or JWT with refresh), authorization checks.
- Validation, consistent errors, cursor pagination, filtering.
- Structured logs, Prometheus metrics, `/healthz` and `/readyz`, pprof.
- Rate limiting, request timeouts, `MaxBytesReader`.
- Tests at all three layers; integration tests against a containerized Postgres.
- Dockerfile (multi-stage, distroless), docker-compose, config via env.

**P7.2 — Real-time feature**
Add WebSockets (`nhooyr.io/websocket` or `gorilla/websocket`) or Server-Sent Events to P7.1: live updates when an entity changes. Hub pattern from P6.1, backed by Postgres `LISTEN/NOTIFY` or Redis pub/sub.

**P7.3 — Background job system**
A worker service consuming from a queue (your own Postgres-backed queue using `SELECT ... FOR UPDATE SKIP LOCKED` — a genuinely instructive pattern), with retries, backoff, dead letter handling, and observability. Understanding `SKIP LOCKED` requires understanding row locking, which is a preview of Phase 9.8.

---

# PHASE 8 — Distributed & Advanced Backend (optional but valuable)

Take this before or after the capstone. It's most useful *after* you understand transactions.

- **gRPC & Protocol Buffers**: schema definition, code generation, the four RPC types, interceptors, deadlines, streaming, and when gRPC beats REST.
- **Message queues**: NATS or Kafka. Delivery semantics (at-most-once, at-least-once, exactly-once and why it's mostly a lie), consumer groups, partitioning, ordering guarantees.
- **Caching**: cache-aside, write-through, write-behind; invalidation strategies; stampede protection with singleflight; Redis as an external cache (which you now understand deeply, having built one).
- **Consistency and consensus**: CAP as it actually applies, linearizability vs sequential vs eventual consistency, Raft (read the paper, then implement leader election and log replication in Go — this is one of the highest-value exercises in distributed systems), quorum reads/writes, vector clocks, CRDTs.
- **Distributed transactions**: two-phase commit and its blocking problem, sagas, outbox.
- **Reading**: *Designing Data-Intensive Applications* (Kleppmann) — chapters 5–9 especially. This book and your database project reinforce each other perfectly.

**P8.1 — Raft implementation**
Leader election, log replication, safety, persistence, snapshotting. Test with a simulated network that drops, delays, and partitions. Then bolt it onto your Phase 6 Redis clone to make it a replicated KV store.

---

# PHASE 9 — CAPSTONE: A Postgres-Style Relational Database in Go

Call it whatever you like — I'll refer to it as **`pgo`**. The goal is a single-node, disk-backed, ACID-compliant relational database with a SQL front-end, a cost-aware query planner, MVCC concurrency control, ARIES-style write-ahead logging, and the **PostgreSQL wire protocol** so that `psql` and any Postgres client library can connect to it.

That last point is the difference between a toy and something people take seriously. When you can run `psql -h localhost -p 5433 -d pgo` and it works, you have built a database.

### Ground rules for the capstone

1. **Build it stage by stage, and keep it working at every stage.** Each stage below ends with a demoable capability.
2. **Test relentlessly.** Every layer gets unit tests, the parser and page decoder get fuzzers, the B+Tree gets property tests, and the storage engine gets crash tests.
3. **Read the theory alongside.** Primary sources listed at each stage.
4. **Instrument from day one.** Add counters for page reads/writes, cache hits, WAL flushes. You'll need them to reason about performance later.
5. **Design docs.** Before each stage, write a short document describing your on-disk format or algorithm. This is how real database work happens, and it will be by far the most impressive part of the repo to anyone reviewing it.

### Prerequisite reading (do this first)

- **CMU 15-445/645 Intro to Database Systems** (Andy Pavlo) — lectures are free on YouTube, with slides and notes. This is the single best resource and it maps almost exactly onto the stages below.
- **CMU 15-721 Advanced Database Systems** — for when you want to go deeper on query execution and concurrency control.
- *Database Internals* — Alex Petrov. Part I is storage engines, Part II is distributed. Read Part I now.
- *Architecture of a Database System* — Hellerstein, Stonebraker, Hamilton. A 100-page survey; read it twice.
- PostgreSQL documentation, Chapter "Internals," plus `src/backend/access/nbtree/README` and `src/backend/access/heap/README` in the Postgres source. They're written for humans.
- Source to read as you go: **`bbolt`** (a clean, small B+Tree store in Go), **SQLite's architecture docs**, **`badger`** (LSM in Go), and Postgres itself for reference.

---

## Stage 9.0 — Architecture & Skeleton

Sketch the whole system before writing code. Your layers, bottom to top:

```
                 ┌─────────────────────────────────┐
                 │   Postgres Wire Protocol (9.13) │
                 ├─────────────────────────────────┤
                 │   Session / Connection Manager   │
                 ├─────────────────────────────────┤
                 │   Parser → AST (9.5)             │
                 ├─────────────────────────────────┤
                 │   Analyzer / Binder + Catalog    │
                 ├─────────────────────────────────┤
                 │   Planner & Optimizer (9.7)      │
                 ├─────────────────────────────────┤
                 │   Execution Engine (9.8)         │
                 ├──────────────┬──────────────────┤
                 │  Access      │  Transaction     │
                 │  Methods     │  Manager (9.9)   │
                 │  (9.3, 9.4)  │  MVCC + Locks    │
                 ├──────────────┴──────────────────┤
                 │   Buffer Pool Manager (9.2)      │
                 ├─────────────────────────────────┤
                 │   WAL / Recovery (9.10)          │
                 ├─────────────────────────────────┤
                 │   Disk Manager / Page I/O (9.1)  │
                 └─────────────────────────────────┘
```

Set up the repo: `cmd/pgo` (server), `cmd/pgoctl` (admin CLI), `internal/storage`, `internal/buffer`, `internal/index`, `internal/catalog`, `internal/sql/{lexer,parser,ast,analyzer}`, `internal/plan`, `internal/exec`, `internal/txn`, `internal/wal`, `internal/wire`.

Decide upfront and write it down: page size (8KB, matching Postgres), endianness (little), maximum tuple size, identifier limits, and your supported type set for v1 (`INT4`, `INT8`, `FLOAT8`, `BOOL`, `TEXT`, `TIMESTAMP`, `NULL`).

---

## Stage 9.1 — Disk Manager & Page Layout

**Concepts:** Why databases manage their own storage rather than using files-per-row; the page as the unit of I/O; slotted pages; internal vs external fragmentation; tuple identifiers.

**Build:**

- A `DiskManager` that reads and writes fixed-size 8KB pages by `PageID` using `ReadAt`/`WriteAt` on a single heap file. Include `AllocatePage`, `DeallocatePage`, and a free-space map.
- **Page header**: `{checksum uint32, pageLSN uint64, flags uint16, lowerOffset uint16, upperOffset uint16, special uint16}` — model it on Postgres's `PageHeaderData`.
- **Slotted page layout**: header at the front, an array of slot pointers `{offset uint16, length uint16, flags uint16}` growing downward from the front, tuple data growing upward from the back, free space in the middle. This layout is what lets tuples be variable-length and lets you compact a page without changing tuple identifiers.
- **TupleID (`TID`)** = `(PageID, SlotNumber)`. Indexes will store these. Understand why the slot indirection is essential: you can move a tuple within a page during compaction and every index pointing at it stays valid.
- Tuple serialization: null bitmap, fixed-width fields inline, variable-length fields with length prefixes. Handle alignment. Write `Serialize(schema, values) []byte` and `Deserialize(schema, []byte) []Value`.
- Page checksums (CRC32C via `hash/crc32` with the Castagnoli table — hardware-accelerated) validated on read.

**Go skills exercised:** `encoding/binary`, byte-slice manipulation, `unsafe` for zero-copy reads, struct alignment, `os.File` positional I/O, fuzz testing.

**Deliverable:** Write 10,000 tuples to a heap file, read them all back, verify checksums, and handle a page that fills up. Fuzz the page decoder against random bytes — it must never panic.

---

## Stage 9.2 — Buffer Pool Manager

**Concepts:** Why a database doesn't just use the OS page cache (control over eviction ordering, forced writes for WAL, knowledge of access patterns, avoiding double-buffering); pinning; dirty pages; replacement policies.

**Build:**

- A fixed-size pool of frames, each holding one page.
- **Page table**: `map[PageID]FrameID` guarded by a latch. Consider a sharded map or a lock-free approach later.
- `FetchPage(pid) (*Page, error)` — returns a **pinned** page. `UnpinPage(pid, isDirty)`. Pin count > 0 means the page cannot be evicted. This is a manual refcount in a garbage-collected language, and getting it right is genuinely hard — leaked pins are your equivalent of a memory leak. Add a debug mode that tracks pin call sites.
- **Replacement policy**: implement LRU first, then **Clock (second-chance)**, then **LRU-K**. Benchmark them on synthetic workloads (sequential scan, zipfian point lookups, mixed) and write up which wins where. Postgres uses clock-sweep; know why.
- Dirty page tracking and `FlushPage`/`FlushAll`.
- **Critical WAL interaction (implement in 9.10, design for it now):** a dirty page may not be written to disk until the WAL records up to its `pageLSN` are durable. This is the WAL rule. Your buffer pool must expose the hook.
- Latching: a page latch (`sync.RWMutex`) per frame, distinct from transaction *locks*. **Latches protect physical structures for the duration of an operation; locks protect logical data for the duration of a transaction.** Be able to explain this difference precisely.

**Go skills exercised:** `sync.RWMutex`, `sync.Cond` (waiting for a free frame), atomics for pin counts, `sync.Pool`, mutex profiling, generics for the replacer interface.

**Deliverable:** A buffer pool with 100 frames serving a 10,000-page file, with hit-rate metrics, tested concurrently under `-race` with no leaked pins.

---

## Stage 9.3 — Heap Files & Table Access

**Build:**

- `TableHeap` — a linked or directory-organized collection of pages holding tuples for one table.
- `InsertTuple`, `GetTuple(tid)`, `UpdateTuple`, `DeleteTuple`. Note: with MVCC coming in 9.9, "update" will become "insert new version + mark old version," and "delete" becomes "mark." Design for that now — don't overwrite in place.
- **Free space map**: track which pages have room, so inserts don't scan linearly.
- **Sequential scan iterator** — the foundation of your executor. Make it an interface so index scans slot in later.
- Handle tuples larger than a page: either reject them in v1, or implement **TOAST**-style overflow pages (Postgres's approach: compress, then move oversized attributes to a side table). Overflow pages are a good stretch goal.

**Deliverable:** Create a table, insert a million rows, scan them, measure throughput, and profile where time goes.

---

## Stage 9.4 — B+Tree Index

The centerpiece data structure. Budget serious time here; it's harder than it looks and the concurrency version is harder still.

**Concepts:** Why B+Trees and not binary trees (fanout vs disk I/O — a tree with fanout 250 reaches a billion keys in 4 levels); leaf vs internal nodes; clustered vs secondary indexes; the difference between Postgres's non-clustered heap+index and MySQL InnoDB's clustered index.

**Build in order:**

1. **Node layout on a page.** Internal nodes: `n` keys and `n+1` child PageIDs. Leaf nodes: keys and values (TIDs), plus a sibling pointer for range scans. Use the slotted-page machinery from 9.1.
2. **Search** — descend from root, binary search within each node.
3. **Insert with splits** — leaf split, propagate the separator key up, root split increases tree height. Handle duplicate keys (either allow with a uniquifier, or use `(key, TID)` as the true key — Postgres does the latter).
4. **Delete with merges/redistribution** — the hardest part. Many production systems (including bbolt and, historically, some Postgres paths) simplify by allowing underfull nodes and reclaiming lazily. Implement the full version, then evaluate whether the simplification is right for you.
5. **Range scans** via leaf sibling pointers, forward and backward.
6. **Bulk loading** — building a tree bottom-up from sorted data is dramatically faster than repeated inserts. Implement it; it's how `CREATE INDEX` works.
7. **Concurrency: latch crabbing (coupling).** Acquire the child's latch before releasing the parent's. Read operations take read latches and release the parent immediately; write operations hold ancestors until they know a split won't propagate ("safe node" analysis). Then implement the **optimistic** version: descend with read latches assuming no split, and retry pessimistically if you were wrong. Test hard under `-race` with concurrent readers and writers.
8. **Variable-length keys**, composite keys, and NULL ordering.

**Property tests:** after every operation, assert — all leaves at the same depth; keys sorted within and across nodes; every internal key is a correct separator; occupancy invariants; sibling chain complete and consistent with an in-order traversal.

**Optional comparison:** implement a simple **LSM tree** (memtable → SSTables → leveled compaction, with bloom filters) and benchmark it against your B+Tree on read-heavy vs write-heavy workloads. This is exactly the design tension between Postgres and RocksDB, and having built both makes you fluent in it. Read the badger/LevelDB design docs.

**Also worth building:** an **extendible hash index** for equality-only lookups — it's how Postgres's hash indexes work and it's a nice contrast in structure.

**Deliverable:** `CREATE INDEX` works, point lookups use it, range predicates use it, and it survives a concurrent stress test.

---

## Stage 9.5 — SQL Front-End: Lexer, Parser, AST

You rehearsed this with the expression evaluator in P1.4. Now do it for real.

**Build:**

- **Lexer** — hand-written, producing tokens with positions. Handle identifiers (quoted and unquoted, case-folding rules — Postgres lowercases unquoted identifiers), keywords, numeric literals, string literals with escapes, operators, comments (`--` and `/* */`), and parameter placeholders (`$1`). Emit good error positions.
- **Recursive-descent parser** for statements: `SELECT` (with `DISTINCT`, target list, `FROM` with joins, `WHERE`, `GROUP BY`, `HAVING`, `ORDER BY`, `LIMIT`/`OFFSET`), `INSERT`, `UPDATE`, `DELETE`, `CREATE TABLE`, `CREATE INDEX`, `DROP`, `BEGIN`/`COMMIT`/`ROLLBACK`, `EXPLAIN`.
- **Expression parsing with precedence climbing (Pratt parsing)** — operators, `AND`/`OR`/`NOT`, comparisons, `IN`, `BETWEEN`, `LIKE`, `IS NULL`, function calls, `CASE`, subqueries, casts.
- **AST** as a set of Go structs implementing a `Node` interface. Add a `String()` method so you can round-trip: parse → print → parse and get the same AST. That round-trip is your best parser test.
- Read Go's own `go/scanner` and `go/parser` first — they're the best-documented hand-written lexer/parser you have easy access to.
- **Fuzz the parser.** It must never panic on any input, only return errors.

**Go skills exercised:** interfaces and type switches over AST nodes, recursion, `panic`/`recover` for parser bailout, string handling, position tracking, fuzzing, the visitor pattern.

**Deliverable:** Parse a corpus of a few hundred real SQL statements. Print helpful syntax errors with a caret pointing at the offending token.

---

## Stage 9.6 — Catalog & Semantic Analysis (Binder)

**Build:**

- **System catalog** stored in the database itself, as regular tables: `pg_class`-style table metadata, `pg_attribute`-style column metadata, index metadata, and statistics. Bootstrapping this (the catalog tables must exist before you can read the catalog) is a genuinely interesting chicken-and-egg problem — solve it with hardcoded bootstrap schemas.
- **Binder/analyzer**: resolve table and column names against the catalog, attach types to every expression, check that the types are compatible, resolve `*` into an explicit column list, validate aggregate usage (no aggregates in `WHERE`, non-aggregated columns must appear in `GROUP BY`), assign output names.
- **Type system**: implicit and explicit casts, a type-coercion lattice, and NULL semantics — three-valued logic. Get `NULL = NULL → NULL`, `NULL AND FALSE → FALSE`, `NULL AND TRUE → NULL` exactly right. This trips up everyone; write an exhaustive truth-table test.
- Constraint metadata: `NOT NULL`, `PRIMARY KEY`, `UNIQUE`, `DEFAULT`, and later `FOREIGN KEY`/`CHECK`.

**Deliverable:** Every statement either produces a fully typed, resolved logical query or a precise semantic error message.

---

## Stage 9.7 — Query Planner & Optimizer

The intellectual high point of the project.

**Build:**

1. **Logical plan** — a tree of relational-algebra operators: Scan, Filter, Project, Join, Aggregate, Sort, Limit, Distinct, Union.
2. **Rule-based rewriting** (the transformations that are always wins):
   - Predicate pushdown — move filters as close to the scans as possible.
   - Projection pushdown — stop carrying columns nobody needs.
   - Constant folding and expression simplification.
   - Subquery decorrelation — turning a correlated subquery into a join. Hard, high value.
   - Join reordering to eliminate cross products.
   - `LIMIT` pushdown into sorts (top-N).
3. **Statistics** — per-column: number of distinct values, null fraction, min/max, and an **equi-depth histogram**. Add a `ANALYZE` command that samples the table and populates them. Optionally implement HyperLogLog or a count-min sketch for distinct-value estimation.
4. **Cardinality estimation** — the selectivity of predicates from histograms; the independence assumption for conjunctions and why it's wrong; join cardinality estimation. Read about why estimation errors compound exponentially through joins ("How Good Are Query Optimizers, Really?" by Leis et al. — read this paper, it's excellent and readable).
5. **Cost model** — I/O cost (sequential vs random page reads) plus CPU cost per tuple. Mirror Postgres's `seq_page_cost`/`random_page_cost`/`cpu_tuple_cost` parameters so you can compare directly.
6. **Physical planning / access path selection** — sequential scan vs index scan vs index-only scan (when the index covers all needed columns) vs bitmap heap scan. Choose based on estimated selectivity, and know where the crossover point is and why it's around a few percent selectivity for a non-clustered index.
7. **Join ordering** — start with left-deep trees via **dynamic programming (System R style)**, then add a greedy or genetic fallback for large join counts (Postgres switches to GEQO past 12 relations).
8. **Join algorithm selection** — nested loop, index nested loop, sort-merge, hash join. Cost each.
9. **`EXPLAIN`** — print the plan tree with estimated rows and costs. Then **`EXPLAIN ANALYZE`** with actual rows and timings. Comparing estimated vs actual on your own optimizer is the fastest way to learn what optimizers get wrong.

**Optional advanced:** a Cascades-style top-down optimizer with a memo structure and transformation rules, which is what modern systems (and CockroachDB, written in Go — read its optimizer) use.

**Deliverable:** For a 3-table join with predicates, your planner picks a plan a human would defend, and `EXPLAIN` shows why.

---

## Stage 9.8 — Execution Engine

**Build:**

- **Volcano/iterator model** first: every operator implements `Next() (Tuple, error)`, pulling from children. It's elegant and it's how Postgres works.
```go
type Executor interface {
    Init() error
    Next() (*Tuple, bool, error)
    Close() error
    Schema() *Schema
}
```
- Operators: `SeqScan`, `IndexScan`, `Filter`, `Projection`, `NestedLoopJoin`, `IndexNestedLoopJoin`, `HashJoin`, `SortMergeJoin`, `Sort`, `Limit`, `Aggregate` (hash and sorted variants), `Distinct`, `Insert`, `Update`, `Delete`, `Values`, `NestedLoop` for subqueries.
- **External sort** — when data exceeds memory: generate sorted runs, then k-way merge with a heap (`container/heap`). Essential and instructive.
- **Grace hash join / partitioned hash join** — when the build side exceeds memory, partition both sides by hash and join partition-pairs.
- **Aggregation**: `COUNT`, `SUM`, `AVG`, `MIN`, `MAX`, with `GROUP BY` and `HAVING`; hash aggregation with spill-to-disk.
- **Expression evaluation** — a tree-walking evaluator first. Then, for a real performance lesson, implement a **compiled/closure-based evaluator**: convert the expression tree into a tree of Go closures once, then call it per tuple. Benchmark the difference; it's typically large, and it teaches you why real systems JIT-compile expressions.
- **Vectorized execution (advanced, high value):** switch from one-tuple-at-a-time to batches of ~1024 values in columnar arrays. Rewrite a few operators (filter, projection, hash aggregate) in this style and benchmark against the tuple-at-a-time versions. Read "MonetDB/X100: Hyper-Pipelining Query Execution" and Pavlo's 15-721 lectures on this. The speedup you measure yourself will teach you more than any explanation.
- **Parallel query execution** — this is where Go shines. An `Exchange` operator that fans a scan out across goroutines and merges results. Use worker pools with bounded parallelism, context cancellation for early termination (a `LIMIT` satisfied means every worker must stop), and be careful about buffer pool pressure. Benchmark speedup vs core count and find where it stops scaling.

**Go skills exercised:** interfaces and dynamic dispatch cost, closures, `container/heap`, generics, worker pools, context cancellation, `sync.Pool` for tuple buffers, profiling and inlining analysis.

**Deliverable:** Run a TPC-H-like query (a 3-way join with aggregation and sort) end to end, and profile it.

---

## Stage 9.9 — Transactions & Concurrency Control

**Concepts:** ACID precisely defined. The isolation anomalies: dirty read, non-repeatable read, phantom read, lost update, write skew. The ANSI isolation levels and how they map to anomalies. Why "serializable" is the only level with an intuitive meaning, and why almost nobody runs it.

**Build:**

1. **Transaction manager** — transaction IDs (XIDs), state (active/committed/aborted), the commit log (`pg_xact`/CLOG equivalent: two bits per transaction recording its outcome).
2. **MVCC, Postgres-style** — this is the design you specifically want:
   - Every tuple carries `xmin` (creating transaction) and `xmax` (deleting transaction).
   - `UPDATE` = insert a new tuple version + set `xmax` on the old one. Nothing is overwritten in place.
   - A **snapshot** = `{xmin, xmax, xip[]}` — the set of transactions in flight when the snapshot was taken.
   - **Visibility rule**: a tuple is visible to a snapshot if its `xmin` committed before the snapshot and is not in `xip`, and its `xmax` is either unset or belongs to a transaction not visible to the snapshot. Implement this function carefully and test it exhaustively — it is the heart of the entire system.
   - Isolation levels fall out of *when you take a snapshot*: `READ COMMITTED` takes a new snapshot per statement; `REPEATABLE READ` takes one per transaction. That's essentially the whole difference. Understanding this makes isolation levels stop being memorized trivia.
   - Compare with the alternative: undo-log MVCC (MySQL/Oracle style) where the current version is in place and old versions live in an undo segment. Know the tradeoffs (Postgres: fast rollback, cheap versioning, but bloat and vacuum; MySQL: compact heap, but rollback and long readers are expensive).
3. **Lock manager** — for writes and for explicit locking. Lock modes and a compatibility matrix, hierarchical locking with intention locks (IS/IX/SIX), a wait-for graph with **deadlock detection** (cycle detection, victim selection) and/or timeout-based prevention (wait-die / wound-wait). Implement deadlock detection as a background goroutine.
4. **Two-phase locking (2PL)** and strict 2PL — why releasing locks early permits cascading aborts.
5. **Serializable Snapshot Isolation (SSI)** — Postgres's true `SERIALIZABLE`. Track read-write dependencies between concurrent transactions and abort when a dangerous structure (two consecutive rw-edges) forms. Read Cahill's paper and the Ports & Grittner paper on the Postgres implementation. This is advanced; treat it as a stretch goal, but it's the most intellectually satisfying part of the whole database.
6. **Write skew** — construct the classic example (the "at least one doctor on call" scenario), show that it succeeds under `REPEATABLE READ` and is correctly aborted under your SSI implementation. Nothing will teach you isolation levels faster than watching this happen in a database you wrote.
7. **VACUUM** — dead tuples accumulate. Build a vacuum process that finds tuples invisible to all live snapshots and reclaims their space, updates the free space map, and removes corresponding index entries. Then implement **XID wraparound** handling conceptually (32-bit XIDs wrap; Postgres freezes old tuples) — even if you use 64-bit XIDs to sidestep it, understand the problem.

**Go skills exercised:** heavy concurrency — mutexes, condition variables, atomics, goroutine coordination, deadlock detection with graph algorithms, `-race`, and testing concurrent correctness (this is where deterministic simulation testing pays off enormously).

**Deliverable:** A test suite of concurrent transaction schedules that asserts exactly which anomalies are possible at each isolation level. Compare your results against real Postgres running the same schedules — they should match.

---

## Stage 9.10 — Write-Ahead Logging & Crash Recovery

**Concepts:** Why durability requires logging rather than just writing pages (a page write is not atomic — torn pages are real). The two WAL rules. ARIES.

**Build:**

1. **Log records** — a compact binary format: `{LSN, prevLSN, xid, type, pageID, undoInfo, redoInfo}`. Types: BEGIN, UPDATE, INSERT, DELETE, COMMIT, ABORT, END, CLR (compensation log record), CHECKPOINT_BEGIN, CHECKPOINT_END.
2. **LSN (Log Sequence Number)** — monotonically increasing, stored in each page header as `pageLSN`.
3. **The WAL rules**:
   - **Write-ahead:** a log record describing a change must reach durable storage *before* the changed page does.
   - **Force-at-commit:** all log records for a transaction must be durable before the commit is acknowledged.
   Both reduce to: `fsync` at the right moments, and never before you have to.
4. **Group commit** — batch the fsyncs of concurrent committing transactions. One fsync for N transactions. This is a huge throughput win and a beautiful use of Go concurrency (a committer goroutine, a channel of waiters, `sync.Cond` or a channel-of-channels handshake). Benchmark commits/sec with and without.
5. **Checkpoints** — fuzzy checkpoints that don't stop the world: record the active transaction table and dirty page table, flush what you can, and write the checkpoint record.
6. **ARIES recovery, three passes**:
   - **Analysis** — scan forward from the last checkpoint to rebuild the dirty page table and transaction table; determine losers (uncommitted at crash) and winners.
   - **Redo** — repeat history. Replay *every* logged update whose `pageLSN < record.LSN`, including those of loser transactions. Understand why redoing losers is correct and necessary.
   - **Undo** — roll back losers in reverse LSN order, writing **CLRs** so that a crash during recovery doesn't undo the same thing twice.
7. **Crash testing** — the highest-value testing you will do. Build a harness that: runs a workload, kills the process at a random point (or, better, uses a fake filesystem that can drop un-fsynced writes and simulate torn pages), restarts, runs recovery, and verifies that every acknowledged commit survived and no uncommitted change persisted. Run it thousands of times with different seeds. Read about how TigerBeetle, FoundationDB, and the Jepsen analyses approach this.
8. **Torn page protection** — Postgres's `full_page_writes`: the first modification to a page after a checkpoint logs the entire page image. Implement it and understand the write amplification cost.

**Go skills exercised:** binary encoding, buffered writes with explicit `Sync()`, `sync.Cond` for group commit, careful error handling on the durability path (an unchecked `Flush` error here is a data-loss bug), fault injection, and testing infrastructure.

**Deliverable:** `kill -9` your server mid-workload, restart, and demonstrate that recovery restores exactly the committed state. Do it 1000 times automatically.

---

## Stage 9.11 — Constraints, DDL, and Data Integrity

- `PRIMARY KEY` and `UNIQUE` enforced via unique indexes (and the concurrency subtlety: a uniqueness check must consider uncommitted concurrent inserts — this is where you need index-level locking or careful MVCC-aware checks).
- `NOT NULL`, `DEFAULT`, `CHECK` constraints evaluated on insert/update.
- `FOREIGN KEY` with referential actions (`CASCADE`, `RESTRICT`, `SET NULL`), implemented as triggers or as explicit checks. The concurrency problem here is real: FK checks under snapshot isolation can produce anomalies, which is why Postgres uses special row locks for them.
- `ALTER TABLE ADD COLUMN` — and the beautiful trick Postgres uses to make it instantaneous by storing the default in the catalog rather than rewriting every tuple.
- `DROP`, `TRUNCATE`, transactional DDL (Postgres has it, MySQL famously doesn't — understand why it requires DDL changes to go through the same MVCC catalog machinery).

---

## Stage 9.12 — Connection & Session Management

- One goroutine per connection — the Postgres process-per-connection model, but cheaper.
- Session state: current transaction, isolation level, prepared statements, portals, search path, timezone.
- Connection limits and backpressure — what happens when you're at max connections.
- Statement timeout, idle-in-transaction timeout, cancellation (Postgres uses a separate connection with a cancel key; implement it and route it to a `context.CancelFunc`).
- Graceful shutdown: stop accepting, let in-flight transactions finish or abort them, checkpoint, close cleanly.

---

## Stage 9.13 — PostgreSQL Wire Protocol

**This is the stage that makes the project real.** Implement the frontend/backend protocol v3 so that `psql`, `pgx`, `psycopg2`, and JDBC can all connect.

- Read the official documentation: PostgreSQL docs, Part IV, "Frontend/Backend Protocol." It is thorough and well-written.
- **Startup**: `StartupMessage`, parameter negotiation, `AuthenticationOk` (start there, then add `AuthenticationCleartextPassword`, then **SCRAM-SHA-256**, which is a genuinely worthwhile crypto implementation exercise), `ParameterStatus`, `BackendKeyData`, `ReadyForQuery`.
- **Simple query protocol**: `Query` → `RowDescription` → `DataRow`* → `CommandComplete` → `ReadyForQuery`. Plus `ErrorResponse` and `NoticeResponse` with correct field codes.
- **Extended query protocol**: `Parse` → `Bind` → `Describe` → `Execute` → `Sync`. Named vs unnamed statements and portals. This is what real client libraries use, and it's where prepared statements live.
- **Type OIDs** — map your types to Postgres's OIDs (`23` = int4, `25` = text, `16` = bool, `701` = float8, …) so clients decode correctly. Support both text and binary result formats.
- `COPY` protocol for bulk load (optional but very useful for loading test data).
- Cancellation requests on a second connection.
- SSL request handling (at minimum, respond correctly to a negotiation attempt).

**Go skills exercised:** binary protocol framing over `net.Conn`, `bufio` with explicit flushing, `encoding/binary`, state machines, `crypto` for SCRAM, careful error handling on a wire protocol where a single misplaced byte desynchronizes everything.

**Deliverable:** `psql -h localhost -p 5433 -U you pgo` connects, `\d` lists tables, `SELECT` returns rows, `EXPLAIN` works, and a Go program using `pgx` runs transactions against it. Then point a real ORM at it and see what breaks.

---

## Stage 9.14 — Performance Engineering

Now bring the whole of Phase 3 to bear on a real system.

- Build a benchmark harness: a **TPC-C-like** OLTP workload (mixed read/write transactions) and a **TPC-H-like** analytical workload (heavy joins and aggregates). Also implement `pgbench`-compatible workloads if you can.
- Profile the OLTP path: expect to find hot spots in the buffer pool page table, tuple deserialization, the lock manager, and WAL fsyncs. Use CPU profiles, mutex profiles, and `go tool trace`.
- Optimizations to try and *measure*:
  - Reduce allocations on the tuple path (`sync.Pool`, buffer reuse, `-gcflags=-m` driven refactoring).
  - Sharded or lock-free buffer pool page table.
  - Zero-copy tuple access with `unsafe` (interpret page bytes directly rather than deserializing).
  - Compiled expression evaluation (from 9.8).
  - Group commit tuning.
  - Prefetching / readahead for sequential scans.
  - `GOGC` and `GOMEMLIMIT` tuning — a database with a large buffer pool has very different GC characteristics than a typical service. Consider allocating the buffer pool as one large `[]byte` so the GC has almost nothing to scan.
  - **PGO**: collect a profile from your benchmark, drop it in as `default.pgo`, rebuild, re-benchmark.
- Every optimization gets a benchstat comparison. Keep a `PERF.md` documenting what worked, what didn't, and why. That document is worth more than the optimizations.
- Compare against real Postgres on the same workload. You will lose — by a lot — and understanding *exactly where* you lose is the most educational part of the entire project.

---

## Stage 9.15 — Stretch Goals

Pick what interests you:

- **Streaming replication** — ship WAL to a replica, replay it, support read-only queries on the replica with conflict handling. Then physical vs logical replication.
- **Logical decoding** — turn WAL into a change stream (this is what Debezium consumes).
- **Point-in-time recovery** — base backup + WAL archive + replay to a target LSN or timestamp.
- **Columnar storage** for analytics, with dictionary and run-length encoding, plus a vectorized scan path.
- **Partitioning** — range and hash partitions with partition pruning in the planner.
- **A cost-based Cascades optimizer** with a memo.
- **Distributed layer** — Raft (from P8.1) replicating your WAL, giving you a fault-tolerant single-writer cluster. This is essentially the CockroachDB/YugabyteDB architecture.
- **Foreign data wrappers** / an extension interface.

---

# Consolidated Project List

| # | Project | Primary skills |
|---|---|---|
| P0.1 | `wc` clone | CLI, flags, io basics |
| P1.1 | `jsonq` JSON query CLI | encoding/json, type switches, recursion |
| P1.2 | In-memory KV store | maps, methods, interfaces, errors |
| P1.3 | Text indexer | file walking, strings, sorting, sets |
| P1.4 | Expression evaluator | lexing, recursive descent, AST — parser rehearsal |
| P2.1 | Concurrent web crawler | goroutines, channels, context, worker pools |
| P2.2 | Sharded concurrent KV store | mutexes, sharding, TTL sweeper, benchmarking |
| P2.3 | Parallel file hasher | bounded parallelism, I/O vs CPU bound |
| P2.4 | Job queue with retries | queues, backoff, graceful drain |
| P4.1 | **Bitcask-style log store** | binary format, CRC, fsync, compaction, recovery |
| P4.2 | Archive tool | io composition, streaming |
| P4.3 | Log aggregator | concurrency + io + time |
| P6.1 | TCP chat server | net, hub pattern, deadlines |
| P6.2 | **Redis clone (RESP, AOF, skip list)** | protocol design, persistence, data structures |
| P6.3 | HTTP/1.1 server from scratch | protocol parsing, framing |
| P7.1 | **Full REST API service** | net/http, Postgres, auth, observability, testing |
| P7.2 | Real-time WebSocket/SSE feature | hub pattern, pub/sub |
| P7.3 | Postgres-backed job system | SKIP LOCKED, row locking |
| P8.1 | Raft implementation | consensus, distributed testing |
| **P9** | **`pgo` — Postgres-style RDBMS** | everything above, plus database internals |

The bolded ones are the portfolio pieces. P4.1 → P6.2 → P9 form a deliberate escalation: durable log store, then a networked data store with persistence, then a full relational engine.

---

# Reading & Resource Map

**Go language and runtime**
- *The Go Programming Language* — Donovan & Kernighan. Still the best book.
- *Learning Go* (2nd ed.) — Jon Bodner. Modern, idiom-focused.
- *100 Go Mistakes and How to Avoid Them* — Teiva Harsanyi. Exceptionally good for the "why does this break" instinct.
- *Concurrency in Go* — Katherine Cox-Buday.
- *Efficient Go* — Bartłomiej Płotka. Performance and profiling.
- Official: *Effective Go*, *Go Code Review Comments*, the Go Memory Model, the Go Blog (especially the slices, interfaces, and GC posts), release notes for every version from 1.18 onward.
- Talks: Rob Pike "Concurrency is not Parallelism," "Go Proverbs"; Bryan Boreham on the scheduler; Google's "Go Runtime" deep dives.
- Source: the Go standard library, and `runtime/` when you want the truth about the scheduler and GC.

**Backend**
- *Let's Go* and *Let's Go Further* — Alex Edwards. The most practical Go web-service books available.
- Postgres official documentation (you'll read it twice — once as a user, once as an implementer).
- *Designing Data-Intensive Applications* — Kleppmann.

**Databases**
- **CMU 15-445** (Andy Pavlo) — lectures, notes, and the BusTub assignments. The BusTub project structure (buffer pool → B+Tree → executors → concurrency control) maps directly onto stages 9.2, 9.4, 9.8, 9.9. Even though it's C++, doing the assignments' *design* in Go is excellent.
- **CMU 15-721** — advanced: vectorized execution, MVCC survey, modern optimizers.
- *Database Internals* — Petrov.
- *Architecture of a Database System* — Hellerstein et al.
- *Readings in Database Systems* ("the Red Book") — Bailis, Hellerstein, Stonebraker.
- Papers worth reading in full: ARIES (Mohan et al.), "Access Path Selection in a Relational DBMS" (Selinger et al. — the original optimizer paper), "An Empirical Evaluation of In-Memory Multi-Version Concurrency Control" (Wu et al.), "Serializable Snapshot Isolation in PostgreSQL" (Ports & Grittner), "How Good Are Query Optimizers, Really?" (Leis et al.), "Are You Sure You Want to Use MMAP in Your DBMS?" (Crotty et al.), "MonetDB/X100" (Boncz et al.).
- Blogs: Andy Pavlo's, Jepsen analyses (Kyle Kingsbury), the CockroachDB engineering blog (Go-specific database engineering — very relevant to you), the TigerBeetle blog (on deterministic testing and safety).
- Go codebases to study: `bbolt` (B+Tree, small enough to read fully), `badger` (LSM), `dolt` (a Git-like SQL database in Go), `go-mysql-server` (a SQL engine in Go — its analyzer and planner are instructive), CockroachDB's `pkg/sql/opt` (a production Cascades optimizer in Go).

---

# Progress Checkpoints

You've genuinely completed a phase when you can do these without looking anything up:

- **Phase 1**: Explain the slice aliasing bug, the typed-nil interface bug, and method sets, with code, from memory.
- **Phase 2**: Draw the GMP scheduler. Explain why a Go server handles 100k connections with 8 threads. Fix a goroutine leak found by pprof.
- **Phase 3**: Read `-gcflags=-m` output and eliminate an allocation. Find a mutex bottleneck with a profile and prove the fix with benchstat.
- **Phase 4**: Design a binary file format with checksums and write a fuzz-hardened decoder for it.
- **Phase 6/7**: Explain what happens, layer by layer, between a client's `SELECT` and your server's response — TCP framing, goroutine scheduling, the netpoller, connection pooling, and back.
- **Phase 9**: Explain exactly why an `UPDATE` in Postgres creates a new tuple version, what a `VACUUM` reclaims, why WAL must be fsynced before the data page, and what the three passes of ARIES recovery each accomplish. Then point at your own code that does each one.

---

# A Note on Pacing

The database is the destination, but Phases 1–4 are what make it possible. Every time you're tempted to skip ahead, remember that stage 9.4 (concurrent B+Tree) is essentially a final exam on Phase 2, and stage 9.14 is a final exam on Phase 3.

The one shortcut I'd endorse: if you want to feel the database work early, build P4.1 (the Bitcask log store) as soon as you finish Phase 4. It's small, it's finishable, and it gives you the storage-engine instincts that make Phase 9 feel like an extension rather than a leap.
