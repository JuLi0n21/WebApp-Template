# Template for a Webapp

## Technologies

- Backend: golang
  - sqlc
  - protobuff restgateway
  - jwt based auth
- Frontend: Vue
  - base.

## How it works

1. Define Services in the `./Proto` dir generate server stubs and swagger file using `make generate` and client service using `make api`
2. Implement server logic
3. Implement client logic
4. Profit
