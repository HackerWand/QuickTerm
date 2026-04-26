export class ThrottleMap {
  private timers = new Map<string, ReturnType<typeof setTimeout>>()
  private lastCallTimes = new Map<string, number>()
  private pendingCalls = new Map<string, () => void>()

  constructor(private delay: number) {}

  call(key: string, fn: () => void): void {
    const now = Date.now()
    const lastCallTime = this.lastCallTimes.get(key) ?? 0
    const elapsed = now - lastCallTime

    if (elapsed >= this.delay) {
      this.lastCallTimes.set(key, now)
      fn()
    } else {
      this.pendingCalls.set(key, fn)
      if (!this.timers.has(key)) {
        const timer = setTimeout(() => {
          this.timers.delete(key)
          const pending = this.pendingCalls.get(key)
          if (pending) {
            this.pendingCalls.delete(key)
            this.lastCallTimes.set(key, Date.now())
            pending()
          }
        }, this.delay - elapsed)
        this.timers.set(key, timer)
      }
    }
  }

  cancel(key: string): void {
    const timer = this.timers.get(key)
    if (timer) {
      clearTimeout(timer)
      this.timers.delete(key)
    }
    this.pendingCalls.delete(key)
    this.lastCallTimes.delete(key)
  }

  cancelAll(): void {
    for (const timer of this.timers.values()) {
      clearTimeout(timer)
    }
    this.timers.clear()
    this.pendingCalls.clear()
    this.lastCallTimes.clear()
  }
}
