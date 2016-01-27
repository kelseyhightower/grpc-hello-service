# cfssl cert configs

This directory contains the cfssl cert configs required to generate TLS certs
and a RSA key/pair for signing and verifying JWT tokens.

## Usage

Fist install the `cfssl` and `cfssljson` command line tools following the [cfssl installation guide](https://github.com/cloudflare/cfssl#installation).

### Generate the certificates

Generate all TLS certs by running the following script from this directory:

```
$ ./generate-certs
```

You should now have the following keys and certs:

```
auth-key.pem
auth.pem
ca-key.pem
ca.pem
client-key.pem
client.pem
hello-key.pem
hello.pem
jwt-key.pem
jwt.pem
```

### Cleanup

Delete all previously generate certs and certificate signing requests:

```
$ rm *.pem *.csr
```
