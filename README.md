# Redirect For ipv6

Redirect Location is a middleware plugin for [Traefik](https://traefik.io) which adds the ability to fix some issues when handling redirect location in combination with path prefixes and the stripPrefix middleware.

## Default handling

If the redirection target is the same host as the request (either a relative path as redirect location or hostname of the redirect location is the same hostname as in the initial request) a stripped path prefix is reatached to the location path (if not already present).

Default handling can be enabled/disabled by config parameter "default".

## Rewrites

The plugin also supports the definition and handling of some rewrites. A rewrite consists of a regular expression defining what is replacement and a replacement string.

## Configuration

### Static

```yaml
experimental:
  plugins:
    redirectIPv6Location:
      modulename: "github.com/init-object/redirect-ipv6"
      version: "v0.0.1" #replace with newest version
```

### Dynamic

To configure the  plugin you should create a [middleware](https://docs.traefik.io/middlewares/overview/) in your dynamic configuration as explained [here](https://docs.traefik.io/middlewares/overview/).
The following example creates and uses the redirect location middleware plugin to add the prefix removed by the stripPrefix middleware to the redirect location path:

```yaml
http:
  routes:
    my-router:
      rule: "Host(`localhost`)"
      service: "my-service"
      middlewares : 
        - "stripPrefix, redirectIPv6Location"
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1"
  middlewares:
    stripPrefix:
      stripPrefix:
        prefixes: "foo"
    redirectIPv6Location:
      plugin:
        redirectIPv6Location:
          default: true
```

The next example creates and uses the redirect location middleware plugin to modify the scheme in every redirect location from http to https:

```yaml
http:
  routes:
    my-router:
      rule: "Host(`localhost`)"
      service: "my-service"
      middlewares : 
        - "redirectIPv6Location"
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1"
  middlewares:
    redirectIPv6Location:
      plugin:
        redirectIPv6Location:
          default: false
          rewrites:
            - regex: "^http://(.+)$"
              replacement: "https://$1"
```

Configuration can also be set via toml or docker labels.
