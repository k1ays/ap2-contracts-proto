# Contracts Workspace

This folder simulates the dedicated proto repository required by Assignment 2.

Contents:
- `payment/v1/payment.proto` - Payment gRPC contract.
- `order/v1/order.proto` - Order streaming gRPC contract.

Recommended GitHub split:
1. Push `contracts-proto` to a dedicated proto repository.
2. Push `contracts-generated` to a dedicated generated-code repository.
3. Point service modules to the generated repository tags.

In this local workspace the services use a `replace` directive to consume `../contracts-generated`.
