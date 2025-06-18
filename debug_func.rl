fn applyTwice(f, x) { f(f(x)) }
fn increment(x) { x + 1 }
applyTwice(increment, 5) 