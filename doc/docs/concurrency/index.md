# Concurrency

Three chapters on goroutine safety and lifecycle. Covers channel-based worker pools, panic recovery with defer/recover, and select-based run loops with context cancellation.

## Chapters
- [chapter-20-channel-delivery-tests.md](chapter-20-channel-delivery-tests.md) -- Channel Delivery Tests — fan-out work to workers via buffered channels with graceful channel-close shutdown
- [chapter-21-panic-recovery.md](chapter-21-panic-recovery.md) -- Panic Recovery in Tests — `defer recover()` converts panics to errors for safe testing of edge cases
- [chapter-22-goroutine-run-loops.md](chapter-22-goroutine-run-loops.md) -- Goroutine Run Loops — select-based event loops with context cancellation and started-channel synchronization

