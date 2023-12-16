# provider-azureextra

`provider-azureextra` is a minimal [Crossplane](https://crossplane.io/) Provider
that is meant to be used as a azureextra for implementing new Providers. It comes
with the following features that are meant to be refactored:

- A `ProviderConfig` type that only points to a credentials `Secret`.
- A `MyType` resource type that serves as an example managed resource.
- A managed resource controller that reconciles `MyType` objects and simply
  prints their configuration in its `Observe` method.

## Developing

1. Use this repository as a azureextra to create a new one.
1. Run `make submodules` to initialize the "build" Make submodule we use for CI/CD.
1. Rename the provider by running the following command:
```shell
  export provider_name=MyProvider # Camel case, e.g. GitHub
  make provider.prepare provider=${provider_name}
```
4. Add your new type by running the following command:
```shell
  export group=armappservice # lower case e.g. core, cache, database, storage, etc.
  export type=MyType # Camel casee.g. Bucket, Database, CacheCluster, etc.
  make provider.addtype provider=${provider_name} group=${group} kind=${type}
```
5. Replace the *sample* group with your new group in apis/{provider}.go
5. Replace the *mytype* type with your new type in internal/controller/{provider}.go
5. Replace the default controller and ProviderConfig implementations with your own
5. Run `make reviewable` to run code generation, linters, and tests.
5. Run `make build` to build the provider.

Refer to Crossplane's [CONTRIBUTING.md] file for more information on how the
Crossplane community prefers to work. The [Provider Development][provider-dev]
guide may also be of use.

[CONTRIBUTING.md]: https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md
[provider-dev]: https://github.com/crossplane/crossplane/blob/master/contributing/guide-provider-development.md

## Connecting to private Azure Container Registry

```shell
az login

az acr token create --name crossplanecli --registry petlab --repository crossplane/provider-azureextra content/write content/read 

az acr login --name petlab



crossplane --verbose xpkg push -f provider-azureextra-v0.0.0-3.g5dfbf0c.xpkg petlab.azurecr.io/crossplane/provider-azureextra:v0.0.1 