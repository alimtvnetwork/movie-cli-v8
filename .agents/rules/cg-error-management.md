# Error Management Standard

## Core Principles
1. **Never Swallow Errors:** All errors must be handled, wrapped, and propagated.
2. **Universal Response Envelope:** All HTTP responses must use the structured Status, Attributes, Results envelope.
3. **appfault Package:** Go backends must return *appfault.AppError. Use ppfault.Wrap() or ppfault.New(). Never use rrors.New().
4. **Three-Tier Architecture:** Errors are handled via Delegated Server -> Go Backend -> Frontend Error Modal.
