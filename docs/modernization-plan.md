# Kanban Backend Modernization Plan

## Goals

- Improve maintainability through SOLID-oriented structure.
- Reduce production risks in startup, request handling, and data consistency.
- Make behavior explicit and test-friendly.

## Guiding Principles

- Single Responsibility: split transport, use-case logic, and persistence responsibilities.
- Open/Closed: centralize reusable request/response policies to avoid scattered edits.
- Liskov Substitution: keep repository contracts predictable and return complete entity state.
- Interface Segregation: keep handler dependencies minimal and use-case-specific.
- Dependency Inversion: rely on abstractions at boundaries, keep concrete details at composition root.

## Roadmap

## Phase 1 (Implemented in this iteration)

1. Runtime hardening:
   - Graceful server shutdown and clean startup error flow.
   - Config loading via explicit errors instead of panic.
2. HTTP modernization:
   - Shared strict JSON decoder with unknown-field and trailing-data checks.
   - Unified body size limits for write endpoints.
3. SOLID (ISP/DIP) on transport layer:
   - Handlers depend on narrow interfaces instead of full repositories where possible.
4. Persistence contract consistency:
   - Update queries return full entity snapshot, not partial fields.
5. Router hardening:
   - Add request ID, recoverer, and per-request timeout middleware.
6. Position consistency on create flows:
   - Serialize column/task position assignment via SQL row locks in owner-scoped create paths.

## Phase 2

1. Application service layer:
   - Introduce use-case services between handlers and repositories.
   - Move auth/business rules out of handlers.
2. Consistency under concurrency:
   - Extend serialization/retry strategy to all position-sensitive mutations (including move/reorder scenarios).
   - Add race-oriented integration tests.
3. Observability:
   - Structured logging with request correlation.
   - Readiness endpoint with DB ping.

## Phase 3

1. CI/CD quality gates:
   - Remove duplicate workflows and add staticcheck/lint gates.
   - Stabilize coverage command and enforce threshold.
2. Security:
   - Add dependency/image vulnerability scanning in CI.
   - Introduce auth rate limiting.

## Success Metrics

- No panic-based config failures in runtime path.
- Stable API responses after update operations with complete entity fields.
- All write endpoints reject unknown JSON fields and oversized payloads.
- Existing test suite green (`test`, `vet`, `race`).
