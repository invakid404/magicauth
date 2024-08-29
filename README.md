# Magicauth

[output.webm](https://github.com/user-attachments/assets/20a4ce57-6f4f-4e84-ae24-7c5005d0291a)

Magicauth is a lightweight OpenID Connect server that leverages Tailscale
identity for seamless authentication. It's designed for self-hosted
applications, eliminating the need for complex OpenID Connect provider setups.

## Features

- Zero-interaction authentication for users connected to your Tailnet
- Minimal OpenID Connect implementation compatible with most self-hosted apps
- No external dependencies besides Tailscale
- Easy configuration via YAML, TOML, JSON, or environment variables
- Optional Kubernetes integration for managing OAuth clients

## How It Works

Magicauth utilizes the special identity headers set by Tailscale Serve/Funnel:

1. When a user makes a request, Tailscale adds identity headers (e.g.,
   `Tailscale-User-Login`, `Tailscale-User-Name`)
2. Magicauth checks the `Tailscale-User-Login` header to authenticate the user
3. If the user is authenticated, Magicauth handles the OpenID Connect flow

This approach provides automatic authentication for users within your Tailnet
without additional login steps.

For more information, see [the Magicauth blog post](https://invak.id/magicauth).

## Installation

### Docker

Use the Docker image provided
[here](https://github.com/invakid404/magicauth/pkgs/container/magicauth).

## Configuration

Magicauth can be configured using:

- YAML, TOML, or JSON files, e.g.:

```yaml
base_url: http://localhost:8080
global_secret: redacted
clients:
  outline:
    audience:
      - https://outline.qilin-qilin.ts.net
    public: false
    client_secret: redacted
    redirect_uris:
      - https://outline.qilin-qilin.ts.net/auth/oidc.callback
    response_types:
      - "id_token"
      - "code"
      - "token"
      - "id_token token"
      - "code id_token"
      - "code token"
      - "code id_token token"
    grant_types:
      - "implicit"
      - "refresh_token"
      - "authorization_code"
      - "password"
      - "client_credentials"
    scopes:
      - "openid"
```

- Environment variables:

```bash
MAGICAUTH_BASE_URL="http://localhost:8080"
MAGICAUTH_GLOBAL_SECRET="redacted"
MAGICAUTH_CLIENTS__OUTLINE__AUDIENCE="https://outline.qilin-qilin.ts.net"
MAGICAUTH_CLIENTS__OUTLINE__PUBLIC="false"
MAGICAUTH_CLIENTS__OUTLINE__CLIENT_SECRET="redacted"
MAGICAUTH_CLIENTS__OUTLINE__REDIRECT_URIS="https://outline.qilin-qilin.ts.net/auth/oidc.callback"
MAGICAUTH_CLIENTS__OUTLINE__RESPONSE_TYPES="id_token,code,..."
MAGICAUTH_CLIENTS__OUTLINE__GRANT_TYPES="implicit,refresh_token,..."
MAGICAUTH_CLIENTS__OUTLINE__SCOPES="openid"
```

- [Kubernetes resources](#kubernetes-integration)

## Kubernetes Integration

To enable Kubernetes integration for managing OAuth clients:

1. Enable the Kubernetes controller by either:

- Setting `enable_k8s` to `true` in the configuration file
- Setting the `MAGICAUTH_ENABLE_K8S` environment variable to `true`

2. Install the CRDs provided
   [here](https://github.com/invakid404/magicauth/tree/master/crds)

Now you can create OAuth clients using Kubernetes resources. For example:

```yaml
apiVersion: magicauth.invak.id/v1
kind: OAuthClient
metadata:
  name: outline
spec:
  audience:
    - https://outline.qilin-qilin.ts.net
  public: false
  clientSecret: redacted
  redirectUris:
    - https://outline.qilin-qilin.ts.net/auth/oidc.callback
  responseTypes: ...
  grantTypes: ...
  scopes: ...
```

## Acknowledgements

- [Ory Fosite](https://github.com/ory/fosite) for the OpenID Connect
  implementation and reference code
