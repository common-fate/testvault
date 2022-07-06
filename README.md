# testvault

This folder contains an API which provides 'vaults'. These vaults represent resources we can grant access to, similar to Okta groups or AWS SSO Permission Sets. This API is used with the `testvault` provider for Granted and is used for end-to-end tests of Granted.

For more information refer to the API documentation in [openapi.yml](./openapi.yml).

## Development

Create a DynamoDB table for testing as follows:

```bash
go run cmd/devcli/main.go db create -n testvault -e TESTVAULT_TABLE_NAME
```

_Note: the above command will fail if the table has already been created. That is fine, you can use the same table._

Add the following entry to your .env file in the root of this repo:

```
TESTVAULT_TABLE_NAME=testvault
```
