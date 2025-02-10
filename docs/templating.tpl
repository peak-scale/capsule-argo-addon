# Templating

For templating you have [Go Sprig](https://masterminds.github.io/sprig/) available. The following custom functions are additionally available:

- `toYaml`
- `fromYaml`
- `toJson`
- `fromJson`
- `toToml`
- `fromToml`

## Context

The follwing data context is available for templating:

```yaml
{{toYaml . }}
```

You can access them via their Map-Path (eg. `.Config.Argo.Namespace`)
